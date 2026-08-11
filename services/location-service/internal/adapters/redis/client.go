// Package redis provides a Redis GEO-backed POI spatial index.
package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"math"
	"time"

	"github.com/google/uuid"
	goredis "github.com/redis/go-redis/v9"

	"github.com/nexora/location-service/internal/app/ports"
	"github.com/nexora/location-service/internal/domain"
)

const (
	// Key prefixes for the location POI GEO index.
	geoKeyPrefix = "location:poi:geo:" // + tenantID[+":"+kind]
	docKeyPrefix = "location:poi:doc:" // + poiID
	idsKeyPrefix = "location:poi:ids:" // + tenantID  (SET of poi IDs)

	// nearestWorldRadiusM covers the globe for NearestOfKind (no radius in port).
	nearestWorldRadiusM = 20_000_000.0
)

// Client is a Redis GEO adapter implementing ports.POIRepo for hot nearby queries.
type Client struct {
	url    string
	client *goredis.Client
	log    *slog.Logger
}

// Open dials Redis from redisURL, verifies connectivity with Ping, and returns a ready Client.
func Open(redisURL string, log *slog.Logger) (*Client, error) {
	if redisURL == "" {
		return nil, fmt.Errorf("REDIS_URL required")
	}
	if log == nil {
		log = slog.Default()
	}
	opt, err := goredis.ParseURL(redisURL)
	if err != nil {
		return nil, fmt.Errorf("redis: parse url: %w", err)
	}
	rdb := goredis.NewClient(opt)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := rdb.Ping(ctx).Err(); err != nil {
		_ = rdb.Close()
		return nil, fmt.Errorf("redis: ping: %w", err)
	}
	log.Info("redis.connected", "addr", opt.Addr, "adapter", "geo-poi")
	return &Client{url: redisURL, client: rdb, log: log}, nil
}

// Close closes the underlying Redis client.
func (c *Client) Close() error {
	if c == nil || c.client == nil {
		return nil
	}
	return c.client.Close()
}

// Ping verifies Redis connectivity.
func (c *Client) Ping(ctx context.Context) error {
	if c == nil || c.client == nil {
		return fmt.Errorf("redis: not connected")
	}
	return c.client.Ping(ctx).Err()
}

func geoKey(tenantID uuid.UUID) string {
	return geoKeyPrefix + tenantID.String()
}

func geoKindKey(tenantID uuid.UUID, kind domain.POIKind) string {
	return geoKeyPrefix + tenantID.String() + ":" + string(kind)
}

func docKey(id uuid.UUID) string {
	return docKeyPrefix + id.String()
}

func idsKey(tenantID uuid.UUID) string {
	return idsKeyPrefix + tenantID.String()
}

type poiDoc struct {
	ID        uuid.UUID      `json:"id"`
	TenantID  uuid.UUID      `json:"tenantId"`
	Kind      domain.POIKind `json:"kind"`
	RefID     string         `json:"refId"`
	Name      string         `json:"name"`
	Lat       float64        `json:"lat"`
	Lng       float64        `json:"lng"`
	Meta      map[string]any `json:"meta,omitempty"`
	Active    bool           `json:"active"`
	CreatedAt time.Time      `json:"createdAt"`
	UpdatedAt time.Time      `json:"updatedAt"`
}

func toDoc(p domain.POI) poiDoc {
	return poiDoc{
		ID: p.ID, TenantID: p.TenantID, Kind: p.Kind, RefID: p.RefID, Name: p.Name,
		Lat: p.Lat, Lng: p.Lng, Meta: p.Meta, Active: p.Active,
		CreatedAt: p.CreatedAt.UTC(), UpdatedAt: p.UpdatedAt.UTC(),
	}
}

func fromDoc(d poiDoc) domain.POI {
	return domain.POI{
		ID: d.ID, TenantID: d.TenantID, Kind: d.Kind, RefID: d.RefID, Name: d.Name,
		Lat: d.Lat, Lng: d.Lng, Meta: d.Meta, Active: d.Active,
		CreatedAt: d.CreatedAt.UTC(), UpdatedAt: d.UpdatedAt.UTC(),
	}
}

func (c *Client) loadDoc(ctx context.Context, id uuid.UUID) (domain.POI, error) {
	raw, err := c.client.Get(ctx, docKey(id)).Bytes()
	if err == goredis.Nil {
		return domain.POI{}, fmt.Errorf("%w: poi", domain.ErrNotFound)
	}
	if err != nil {
		return domain.POI{}, err
	}
	var d poiDoc
	if err := json.Unmarshal(raw, &d); err != nil {
		return domain.POI{}, fmt.Errorf("redis: decode poi: %w", err)
	}
	return fromDoc(d), nil
}

func (c *Client) saveDoc(ctx context.Context, p domain.POI) error {
	b, err := json.Marshal(toDoc(p))
	if err != nil {
		return err
	}
	return c.client.Set(ctx, docKey(p.ID), b, 0).Err()
}

// Upsert indexes a POI document and GEOADD (or GEOREM when inactive).
func (c *Client) Upsert(ctx context.Context, p domain.POI) error {
	var prev domain.POI
	if existing, err := c.loadDoc(ctx, p.ID); err == nil {
		prev = existing
	}

	if err := c.saveDoc(ctx, p); err != nil {
		return err
	}
	if err := c.client.SAdd(ctx, idsKey(p.TenantID), p.ID.String()).Err(); err != nil {
		return err
	}

	member := p.ID.String()
	tenantGeo := geoKey(p.TenantID)

	// Drop stale kind index when kind changes.
	if prev.ID != uuid.Nil && prev.Kind != p.Kind {
		_ = c.client.ZRem(ctx, geoKindKey(prev.TenantID, prev.Kind), member).Err()
	}

	if !p.Active {
		_ = c.client.ZRem(ctx, tenantGeo, member).Err()
		_ = c.client.ZRem(ctx, geoKindKey(p.TenantID, p.Kind), member).Err()
		if prev.ID != uuid.Nil && prev.Kind != p.Kind {
			_ = c.client.ZRem(ctx, geoKindKey(prev.TenantID, prev.Kind), member).Err()
		}
		return nil
	}

	loc := &goredis.GeoLocation{
		Name:      member,
		Longitude: p.Lng,
		Latitude:  p.Lat,
	}
	if err := c.client.GeoAdd(ctx, tenantGeo, loc).Err(); err != nil {
		return err
	}
	return c.client.GeoAdd(ctx, geoKindKey(p.TenantID, p.Kind), loc).Err()
}

// Get returns a POI document by id for the tenant.
func (c *Client) Get(ctx context.Context, tenantID, id uuid.UUID) (domain.POI, error) {
	p, err := c.loadDoc(ctx, id)
	if err != nil {
		return domain.POI{}, err
	}
	if p.TenantID != tenantID {
		return domain.POI{}, fmt.Errorf("%w: poi", domain.ErrNotFound)
	}
	return p, nil
}

// Nearby uses GEOSEARCH (radius meters) matching ports.POIRepo.
func (c *Client) Nearby(ctx context.Context, q domain.NearbyQuery) ([]domain.POI, error) {
	limit := q.Limit
	if limit <= 0 {
		limit = 50
	}
	key := geoKey(q.TenantID)
	if q.Kind != nil {
		key = geoKindKey(q.TenantID, *q.Kind)
	}
	return c.geoSearchHydrate(ctx, key, q.TenantID, q.Center.Lng, q.Center.Lat, q.RadiusM, limit, q.Kind)
}

// NearestOfKind uses GEOSEARCH with a world-scale radius and COUNT.
func (c *Client) NearestOfKind(ctx context.Context, tenantID uuid.UUID, kind domain.POIKind, lat, lng float64, limit int) ([]domain.POI, error) {
	if limit <= 0 {
		limit = 1
	}
	k := kind
	return c.geoSearchHydrate(ctx, geoKindKey(tenantID, kind), tenantID, lng, lat, nearestWorldRadiusM, limit, &k)
}

func (c *Client) geoSearchHydrate(
	ctx context.Context,
	key string,
	tenantID uuid.UUID,
	lng, lat, radiusM float64,
	limit int,
	kind *domain.POIKind,
) ([]domain.POI, error) {
	// Over-fetch slightly so inactive/mismatched docs can be filtered.
	fetch := limit * 3
	if fetch < limit {
		fetch = limit
	}
	if fetch > 500 {
		fetch = 500
	}

	names, err := c.client.GeoSearch(ctx, key, &goredis.GeoSearchQuery{
		Longitude:  lng,
		Latitude:   lat,
		Radius:     radiusM,
		RadiusUnit: "m",
		Sort:       "ASC",
		Count:      fetch,
	}).Result()
	if err != nil {
		return nil, err
	}

	out := make([]domain.POI, 0, limit)
	for _, name := range names {
		id, err := uuid.Parse(name)
		if err != nil {
			continue
		}
		p, err := c.loadDoc(ctx, id)
		if err != nil {
			continue
		}
		if p.TenantID != tenantID || !p.Active {
			continue
		}
		if kind != nil && p.Kind != *kind {
			continue
		}
		out = append(out, p)
		if len(out) >= limit {
			break
		}
	}
	return out, nil
}

// InBBox uses GEOSEARCH BYBOX around the bbox center.
func (c *Client) InBBox(ctx context.Context, tenantID uuid.UUID, bbox domain.BBox, kind *domain.POIKind, limit int) ([]domain.POI, error) {
	if limit <= 0 {
		limit = 50
	}
	key := geoKey(tenantID)
	if kind != nil {
		key = geoKindKey(tenantID, *kind)
	}

	centerLat := (bbox.MinLat + bbox.MaxLat) / 2
	centerLng := (bbox.MinLng + bbox.MaxLng) / 2
	// Approximate meters-per-degree at center latitude.
	latM := (bbox.MaxLat - bbox.MinLat) * 111_320
	cosLat := math.Cos(centerLat * math.Pi / 180)
	if cosLat < 0.01 {
		cosLat = 0.01
	}
	lngM := (bbox.MaxLng - bbox.MinLng) * 111_320 * cosLat
	if latM <= 0 {
		latM = 1
	}
	if lngM <= 0 {
		lngM = 1
	}

	fetch := limit * 3
	if fetch > 500 {
		fetch = 500
	}
	names, err := c.client.GeoSearch(ctx, key, &goredis.GeoSearchQuery{
		Longitude: centerLng,
		Latitude:  centerLat,
		BoxWidth:  lngM,
		BoxHeight: latM,
		BoxUnit:   "m",
		Sort:      "ASC",
		Count:     fetch,
	}).Result()
	if err != nil {
		return nil, err
	}

	out := make([]domain.POI, 0, limit)
	for _, name := range names {
		id, err := uuid.Parse(name)
		if err != nil {
			continue
		}
		p, err := c.loadDoc(ctx, id)
		if err != nil {
			continue
		}
		if p.TenantID != tenantID || !p.Active {
			continue
		}
		if kind != nil && p.Kind != *kind {
			continue
		}
		if !bbox.Contains(domain.LatLng{Lat: p.Lat, Lng: p.Lng}) {
			continue
		}
		out = append(out, p)
		if len(out) >= limit {
			break
		}
	}
	return out, nil
}

// Count returns active POIs for a tenant.
func (c *Client) Count(ctx context.Context, tenantID uuid.UUID) (int, error) {
	ids, err := c.client.SMembers(ctx, idsKey(tenantID)).Result()
	if err != nil {
		return 0, err
	}
	n := 0
	for _, s := range ids {
		id, err := uuid.Parse(s)
		if err != nil {
			continue
		}
		p, err := c.loadDoc(ctx, id)
		if err != nil {
			continue
		}
		if p.TenantID == tenantID && p.Active {
			n++
		}
	}
	return n, nil
}

var _ ports.POIRepo = (*Client)(nil)
