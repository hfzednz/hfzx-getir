package app

import (
	"context"
	"math"
	"time"

	"github.com/google/uuid"
	"github.com/nexora/data-platform-service/internal/domain"
)

// RegisterSchema registers or versions an event schema.
func (d *Deps) RegisterSchema(ctx context.Context, s domain.EventSchema) (domain.EventSchema, error) {
	if s.TenantID == uuid.Nil || s.Name == "" || !domain.ValidFamily(s.Family) {
		return s, domain.ErrInvalidArgument
	}
	s.Name = domain.NormalizeEventName(s.Name)
	if s.Compatibility == "" {
		s.Compatibility = domain.CompatBackward
	}
	if s.Version <= 0 {
		if latest, err := d.Schemas.GetLatest(ctx, s.TenantID, s.Name); err == nil {
			s.Version = latest.Version + 1
			// simple compatibility: required fields can only shrink under backward
			oldReq := domain.RequiredFieldsFromSchema(latest.JSONSchema)
			newReq := domain.RequiredFieldsFromSchema(s.JSONSchema)
			if s.Compatibility == domain.CompatBackward || s.Compatibility == domain.CompatFull {
				oldSet := map[string]struct{}{}
				for _, r := range oldReq {
					oldSet[r] = struct{}{}
				}
				for _, r := range newReq {
					if _, ok := oldSet[r]; !ok && len(oldReq) > 0 {
						return s, domain.ErrSchemaIncompatible
					}
				}
			}
		} else {
			s.Version = 1
		}
	}
	if s.ID == uuid.Nil {
		s.ID = d.newID()
	}
	s.Active = true
	s.CreatedAt = d.now()
	if err := d.Schemas.Save(ctx, s); err != nil {
		return s, err
	}
	d.emit(ctx, s.TenantID, s.ID, domain.EventSchemaRegistered, map[string]any{
		"name": s.Name, "version": s.Version, "family": s.Family,
	})
	_ = d.Catalog.SaveAsset(ctx, domain.CatalogAsset{
		ID: d.newID(), TenantID: s.TenantID, Name: s.Name, Type: "event",
		Owner: "data-platform", Description: "schema v" + itoa(s.Version),
		Classification: "internal", UpdatedAt: d.now(),
	})
	return s, nil
}

// IngestEvent validates and stores a bronze event, then runs stream + realtime hooks.
func (d *Deps) IngestEvent(ctx context.Context, e domain.AnalyticsEvent) (domain.AnalyticsEvent, error) {
	if e.TenantID == uuid.Nil || e.Name == "" {
		return e, domain.ErrInvalidArgument
	}
	e.Name = domain.NormalizeEventName(e.Name)
	if e.Family == "" {
		e.Family = domain.FamilySystem
	}
	if !domain.ValidFamily(e.Family) {
		return e, domain.ErrInvalidArgument
	}
	if e.IdempotencyKey != "" {
		if existing, ok, err := d.Events.GetByIdempotency(ctx, e.TenantID, e.IdempotencyKey); err != nil {
			return e, err
		} else if ok {
			return existing, nil
		}
	}
	now := d.now()
	if e.ID == uuid.Nil {
		e.ID = d.newID()
	}
	if e.OccurredAt.IsZero() {
		e.OccurredAt = now
	}
	e.IngestedAt = now
	e.Layer = domain.LayerBronze
	if e.Payload == nil {
		e.Payload = map[string]any{}
	}
	e.PayloadHash = domain.HashPayload(e.Payload)
	e.Valid = true

	if schema, err := d.Schemas.GetLatest(ctx, e.TenantID, e.Name); err == nil {
		e.SchemaVersion = schema.Version
		e.Family = schema.Family
		if req := domain.RequiredFieldsFromSchema(schema.JSONSchema); len(req) > 0 {
			if err := domain.ValidateRequired(e.Payload, req); err != nil {
				e.Valid = false
				e.Error = "missing_required_fields"
			}
		}
	}

	// dedupe identical payload within 1h (quality)
	if n, _ := d.Events.CountByHash(ctx, e.TenantID, e.PayloadHash, now.Add(-time.Hour)); n > 0 && e.IdempotencyKey == "" {
		// allow but flag
		e.Payload["_dedupeHint"] = true
	}

	if err := d.Events.Save(ctx, e); err != nil {
		return e, err
	}
	d.emit(ctx, e.TenantID, e.ID, domain.EventEventIngested, map[string]any{
		"name": e.Name, "family": e.Family, "valid": e.Valid,
	})

	if e.Valid {
		_ = d.processStreams(ctx, e)
		_ = d.bumpRealtime(ctx, e)
		_ = d.maybeFact(ctx, e)
		_ = d.promoteSilver(ctx, e)
	} else {
		_ = d.Quality.Save(ctx, domain.QualityCheck{
			ID: d.newID(), TenantID: e.TenantID, AssetName: e.Name, CheckType: "validity",
			Passed: false, Score: 0, Details: e.Error, CreatedAt: now,
		})
		d.emit(ctx, e.TenantID, e.ID, domain.EventQualityFailed, map[string]any{"name": e.Name, "error": e.Error})
	}
	_ = d.evaluateAlerts(ctx, e.TenantID)
	return e, nil
}

func (d *Deps) processStreams(ctx context.Context, e domain.AnalyticsEvent) error {
	jobs, err := d.Streams.ListJobs(ctx, e.TenantID)
	if err != nil {
		return err
	}
	now := d.now()
	for _, job := range jobs {
		if !job.Enabled || job.EventName != e.Name {
			continue
		}
		win := time.Duration(job.WindowSec) * time.Second
		if win <= 0 {
			win = time.Minute
		}
		start := e.OccurredAt.Truncate(win)
		end := start.Add(win)
		delta := 1.0
		if job.Agg == "sum" || job.Agg == "avg" {
			if job.MetricField != "" {
				if v, ok := e.Payload[job.MetricField]; ok {
					delta = asFloat(v)
				}
			}
		}
		aggs, _ := d.Streams.ListAggregates(ctx, e.TenantID, job.ID, 20)
		var cur domain.AggregateWindow
		found := false
		for _, a := range aggs {
			if a.WindowStart.Equal(start) {
				cur = a
				found = true
				break
			}
		}
		if !found {
			cur = domain.AggregateWindow{TenantID: e.TenantID, JobID: job.ID, WindowStart: start, WindowEnd: end}
		}
		cur.Count++
		cur.Value += delta
		if job.Agg == "avg" && cur.Count > 0 {
			// store sum in Value; consumers divide — keep running sum
		}
		cur.UpdatedAt = now
		_ = d.Streams.UpsertAggregate(ctx, cur)
		if d.OLAP != nil {
			_ = d.OLAP.InsertAggregate(ctx, e.TenantID, job.Name, delta, e.OccurredAt)
		}
		d.emit(ctx, e.TenantID, job.ID, domain.EventAggregateUpdated, map[string]any{
			"job": job.Name, "value": cur.Value, "count": cur.Count,
		})
	}
	return nil
}

func (d *Deps) bumpRealtime(ctx context.Context, e domain.AnalyticsEvent) error {
	_, _ = d.Realtime.Incr(ctx, e.TenantID, "events.total", 1, d.now())
	_, _ = d.Realtime.Incr(ctx, e.TenantID, "events."+e.Family, 1, d.now())
	switch e.Name {
	case "order.placed":
		_, _ = d.Realtime.Incr(ctx, e.TenantID, "live.orders", 1, d.now())
		if v, ok := e.Payload["amountMinor"]; ok {
			_, _ = d.Realtime.Incr(ctx, e.TenantID, "live.revenue", asFloat(v), d.now())
		}
	case "delivery.completed":
		_, _ = d.Realtime.Incr(ctx, e.TenantID, "live.deliveries", 1, d.now())
	case "user.session_start":
		_, _ = d.Realtime.Incr(ctx, e.TenantID, "live.users", 1, d.now())
	}
	return nil
}

func (d *Deps) maybeFact(ctx context.Context, e domain.AnalyticsEvent) error {
	fact := ""
	measures := map[string]float64{}
	dims := map[string]string{"event": e.Name, "family": e.Family}
	switch e.Family {
	case domain.FamilyOrder:
		fact = "fact_orders"
		measures["orders"] = 1
		if v, ok := e.Payload["amountMinor"]; ok {
			measures["revenue_minor"] = asFloat(v)
		}
	case domain.FamilyPayment:
		fact = "fact_payments"
		measures["payments"] = 1
	case domain.FamilyDelivery:
		fact = "fact_deliveries"
		measures["deliveries"] = 1
		if v, ok := e.Payload["onTime"]; ok && asFloat(v) > 0 {
			measures["on_time"] = 1
		}
	case domain.FamilySearch:
		fact = "fact_search"
		measures["searches"] = 1
		if z, ok := e.Payload["zeroResult"]; ok && asBool(z) {
			measures["zero_results"] = 1
		}
	case domain.FamilySupport:
		fact = "fact_support"
		measures["tickets"] = 1
	default:
		return nil
	}
	f := domain.FactSnapshot{
		ID: d.newID(), TenantID: e.TenantID, FactTable: fact, GrainKey: e.ID.String(),
		Measures: measures, Dims: dims, AsOf: e.OccurredAt, CreatedAt: d.now(),
	}
	return d.Warehouse.SaveFact(ctx, f)
}

func (d *Deps) promoteSilver(ctx context.Context, e domain.AnalyticsEvent) error {
	// silver = validated bronze copy marker via lake dataset touch
	ds := domain.LakeDataset{
		ID: d.newID(), TenantID: e.TenantID, Name: "events_" + e.Family,
		Layer: domain.LayerSilver, Format: "parquet",
		Location: "s3://nexora-lake/" + e.TenantID.String() + "/silver/" + e.Family + "/",
		PartitionBy: []string{"dt", "city_id"}, RetentionDays: 90, UpdatedAt: d.now(),
	}
	_ = d.Lake.SaveDataset(ctx, ds)
	_ = d.Catalog.SaveLineage(ctx, domain.LineageEdge{
		TenantID: e.TenantID, FromName: e.Name, ToName: ds.Name, Kind: "ingests",
	})
	return nil
}

// UpsertStreamJob creates a streaming aggregation job.
func (d *Deps) UpsertStreamJob(ctx context.Context, j domain.StreamJob) (domain.StreamJob, error) {
	if j.TenantID == uuid.Nil || j.Name == "" || j.EventName == "" {
		return j, domain.ErrInvalidArgument
	}
	j.EventName = domain.NormalizeEventName(j.EventName)
	if j.Agg == "" {
		j.Agg = "count"
	}
	if j.WindowSec <= 0 {
		j.WindowSec = 60
	}
	if j.ID == uuid.Nil {
		j.ID = d.newID()
	}
	j.Enabled = true
	j.UpdatedAt = d.now()
	return j, d.Streams.SaveJob(ctx, j)
}

// RefreshMarts computes gold KPIs from facts.
func (d *Deps) RefreshMarts(ctx context.Context, tenantID uuid.UUID) ([]domain.KPIValue, error) {
	now := d.now()
	facts, err := d.Warehouse.ListFacts(ctx, tenantID, "", 10000)
	if err != nil {
		return nil, err
	}
	var orders, revenue, deliveries, onTime, searches, zero, tickets float64
	for _, f := range facts {
		orders += f.Measures["orders"]
		revenue += f.Measures["revenue_minor"]
		deliveries += f.Measures["deliveries"]
		onTime += f.Measures["on_time"]
		searches += f.Measures["searches"]
		zero += f.Measures["zero_results"]
		tickets += f.Measures["tickets"]
	}
	kpis := []domain.KPIValue{
		{TenantID: tenantID, Key: domain.KPIOrders, Value: orders, Unit: "count", AsOf: now},
		{TenantID: tenantID, Key: domain.KPIRevenue, Value: revenue, Unit: "minor", AsOf: now},
	}
	if deliveries > 0 {
		kpis = append(kpis, domain.KPIValue{TenantID: tenantID, Key: domain.KPIDeliverySLA, Value: onTime / deliveries, Unit: "ratio", AsOf: now})
		kpis = append(kpis, domain.KPIValue{TenantID: tenantID, Key: domain.KPIFulfillment, Value: math.Min(1, deliveries/math.Max(1, orders)), Unit: "ratio", AsOf: now})
	}
	if searches > 0 {
		// conversion proxy: orders/searches
		kpis = append(kpis, domain.KPIValue{TenantID: tenantID, Key: domain.KPIConversion, Value: orders / searches, Unit: "ratio", AsOf: now})
	}
	for _, k := range kpis {
		_ = d.Warehouse.SaveKPI(ctx, k)
	}
	_ = d.Lake.SaveDataset(ctx, domain.LakeDataset{
		ID: d.newID(), TenantID: tenantID, Name: "mart_kpis", Layer: domain.LayerGold,
		Format: "parquet", Location: "s3://nexora-lake/" + tenantID.String() + "/gold/mart_kpis/",
		RetentionDays: 365, UpdatedAt: now,
	})
	d.emit(ctx, tenantID, d.newID(), domain.EventMartRefreshed, map[string]any{"kpis": len(kpis)})
	return kpis, nil
}

func asFloat(v any) float64 {
	switch t := v.(type) {
	case float64:
		return t
	case float32:
		return float64(t)
	case int:
		return float64(t)
	case int64:
		return float64(t)
	default:
		return 0
	}
}

func asBool(v any) bool {
	switch t := v.(type) {
	case bool:
		return t
	case float64:
		return t != 0
	case string:
		return t == "true" || t == "1"
	default:
		return false
	}
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	digits := []byte{}
	for n > 0 {
		digits = append([]byte{byte('0' + n%10)}, digits...)
		n /= 10
	}
	return string(digits)
}
