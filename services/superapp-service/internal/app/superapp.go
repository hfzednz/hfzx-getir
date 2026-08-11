package app

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/nexora/superapp-service/internal/domain"
)

func (d *Deps) UpsertModule(ctx context.Context, m domain.Module) (domain.Module, error) {
	if err := domain.ValidateModule(m); err != nil {
		return domain.Module{}, err
	}
	now := d.now()
	if m.ID == uuid.Nil {
		if existing, err := d.Modules.GetByKey(ctx, m.TenantID, m.Key); err == nil {
			m.ID = existing.ID
			m.CreatedAt = existing.CreatedAt
		} else {
			m.ID = d.newID()
			m.CreatedAt = now
		}
	}
	if m.Status == "" {
		m.Status = domain.ModuleDraft
	}
	if m.LatestVersion == "" {
		m.LatestVersion = "0.1.0"
	}
	m.UpdatedAt = now
	if err := d.Modules.Save(ctx, m); err != nil {
		return domain.Module{}, err
	}
	return m, nil
}

func (d *Deps) PublishManifest(ctx context.Context, man domain.ModuleManifest) (domain.ModuleManifest, error) {
	if err := domain.ValidateManifest(man); err != nil {
		return domain.ModuleManifest{}, err
	}
	if err := domain.ValidatePermissions(man.Permissions); err != nil {
		return domain.ModuleManifest{}, err
	}
	mod, err := d.Modules.Get(ctx, man.TenantID, man.ModuleID)
	if err != nil {
		return domain.ModuleManifest{}, err
	}
	shell := d.ShellVersion
	if shell == "" {
		shell = "1.0.0"
	}
	man.Compatible = domain.SemverCompatible(shell, man.MinShellVersion)
	if !man.Compatible {
		return domain.ModuleManifest{}, domain.ErrIncompatible
	}
	man.ID = d.newID()
	man.CreatedAt = d.now()
	if err := d.Manifests.Save(ctx, man); err != nil {
		return domain.ModuleManifest{}, err
	}
	mod.LatestVersion = man.Version
	mod.Status = domain.ModulePublished
	mod.UpdatedAt = d.now()
	_ = d.Modules.Save(ctx, mod)
	d.emit(ctx, man.TenantID, mod.ID, domain.EventModuleActivated, map[string]any{
		"key": mod.Key, "version": man.Version, "kind": string(mod.Kind),
	})
	return man, nil
}

func (d *Deps) SeedMiniApps(ctx context.Context, tenantID uuid.UUID) error {
	for _, m := range domain.DefaultMiniAppCatalog() {
		m.TenantID = tenantID
		m.PublisherID = "nexora"
		mod, err := d.UpsertModule(ctx, m)
		if err != nil {
			return err
		}
		_, err = d.PublishManifest(ctx, domain.ModuleManifest{
			TenantID: tenantID, ModuleID: mod.ID, Version: "1.0.0",
			EntryPoint: "flutter://mini/" + mod.Key, MinShellVersion: "1.0.0",
			Permissions: []string{domain.PermNavigation, domain.PermSearch},
			Hooks:       []string{"navigation", "search"},
			Signature:   "sig-" + mod.Key, Checksum: "sha256-" + mod.Key,
			BundleURI:   "https://cdn.nexora.example/mini/" + mod.Key + "/1.0.0.apk",
		})
		if err != nil {
			return err
		}
		_, _ = d.UpsertListing(ctx, domain.StoreListing{
			TenantID: tenantID, ModuleID: mod.ID, Featured: mod.Key == "qc", Currency: "TRY",
		})
	}
	return nil
}

func (d *Deps) UpsertListing(ctx context.Context, l domain.StoreListing) (domain.StoreListing, error) {
	if l.TenantID == uuid.Nil || l.ModuleID == uuid.Nil {
		return domain.StoreListing{}, domain.ErrInvalidArgument
	}
	if existing, err := d.Listings.GetByModule(ctx, l.TenantID, l.ModuleID); err == nil {
		l.ID = existing.ID
		l.RatingAvg = existing.RatingAvg
		l.RatingCount = existing.RatingCount
		l.Installs = existing.Installs
	} else {
		l.ID = d.newID()
	}
	if l.Currency == "" {
		l.Currency = "TRY"
	}
	l.UpdatedAt = d.now()
	if err := d.Listings.Save(ctx, l); err != nil {
		return domain.StoreListing{}, err
	}
	return l, nil
}

func (d *Deps) InstallModule(ctx context.Context, tenantID uuid.UUID, subjectID string, moduleKey string) (domain.Install, error) {
	if subjectID == "" || moduleKey == "" {
		return domain.Install{}, domain.ErrInvalidArgument
	}
	mod, err := d.Modules.GetByKey(ctx, tenantID, moduleKey)
	if err != nil {
		return domain.Install{}, err
	}
	if mod.Status != domain.ModulePublished {
		return domain.Install{}, domain.ErrForbidden
	}
	if d.LiveOps != nil {
		if ok, err := d.LiveOps.ModuleEnabled(ctx, tenantID, moduleKey, subjectID); err == nil && !ok {
			return domain.Install{}, domain.ErrForbidden
		}
	}
	man, err := d.Manifests.Get(ctx, tenantID, mod.ID, mod.LatestVersion)
	if err != nil {
		man, err = d.Manifests.Latest(ctx, tenantID, mod.ID)
		if err != nil {
			return domain.Install{}, err
		}
	}
	inst, err := d.Installs.Get(ctx, tenantID, subjectID, mod.ID)
	now := d.now()
	if err != nil {
		inst = domain.Install{
			ID: d.newID(), TenantID: tenantID, SubjectID: subjectID, ModuleID: mod.ID,
			Version: man.Version, Status: domain.InstallActive, InstalledAt: now, UpdatedAt: now,
		}
	} else {
		inst.PreviousVersion = inst.Version
		inst.Version = man.Version
		inst.Status = domain.InstallActive
		inst.UpdatedAt = now
	}
	if err := d.Installs.Save(ctx, inst); err != nil {
		return domain.Install{}, err
	}
	if listing, err := d.Listings.GetByModule(ctx, tenantID, mod.ID); err == nil {
		listing.Installs++
		listing.UpdatedAt = now
		_ = d.Listings.Save(ctx, listing)
	}
	event := domain.EventPluginInstalled
	if mod.Kind == domain.KindMiniApp {
		event = domain.EventPluginInstalled // still emit install; launch is separate
	}
	d.emit(ctx, tenantID, inst.ID, event, map[string]any{
		"moduleKey": mod.Key, "version": man.Version, "kind": string(mod.Kind), "subjectId": subjectID,
	})
	if d.Metrics != nil {
		_ = d.Metrics.Record(ctx, "superapp.install", map[string]string{"key": mod.Key}, 1)
	}
	return inst, nil
}

func (d *Deps) UpdateModule(ctx context.Context, tenantID uuid.UUID, subjectID, moduleKey string) (domain.Install, error) {
	mod, err := d.Modules.GetByKey(ctx, tenantID, moduleKey)
	if err != nil {
		return domain.Install{}, err
	}
	inst, err := d.Installs.Get(ctx, tenantID, subjectID, mod.ID)
	if err != nil {
		return domain.Install{}, err
	}
	man, err := d.Manifests.Get(ctx, tenantID, mod.ID, mod.LatestVersion)
	if err != nil {
		man, err = d.Manifests.Latest(ctx, tenantID, mod.ID)
		if err != nil {
			return domain.Install{}, err
		}
	}
	inst.PreviousVersion = inst.Version
	inst.Version = man.Version
	inst.Status = domain.InstallActive
	inst.UpdatedAt = d.now()
	if err := d.Installs.Save(ctx, inst); err != nil {
		return domain.Install{}, err
	}
	d.emit(ctx, tenantID, inst.ID, domain.EventPluginUpdated, map[string]any{
		"moduleKey": mod.Key, "version": man.Version, "subjectId": subjectID,
	})
	return inst, nil
}

func (d *Deps) RemoveModule(ctx context.Context, tenantID uuid.UUID, subjectID, moduleKey string) (domain.Install, error) {
	mod, err := d.Modules.GetByKey(ctx, tenantID, moduleKey)
	if err != nil {
		return domain.Install{}, err
	}
	inst, err := d.Installs.Get(ctx, tenantID, subjectID, mod.ID)
	if err != nil {
		return domain.Install{}, err
	}
	inst.Status = domain.InstallRemoved
	inst.UpdatedAt = d.now()
	if err := d.Installs.Save(ctx, inst); err != nil {
		return domain.Install{}, err
	}
	d.emit(ctx, tenantID, inst.ID, domain.EventPluginRemoved, map[string]any{
		"moduleKey": mod.Key, "subjectId": subjectID,
	})
	return inst, nil
}

func (d *Deps) RollbackInstall(ctx context.Context, tenantID uuid.UUID, subjectID, moduleKey string) (domain.Install, error) {
	mod, err := d.Modules.GetByKey(ctx, tenantID, moduleKey)
	if err != nil {
		return domain.Install{}, err
	}
	inst, err := d.Installs.Get(ctx, tenantID, subjectID, mod.ID)
	if err != nil {
		return domain.Install{}, err
	}
	if inst.PreviousVersion == "" {
		return domain.Install{}, domain.ErrIllegalTransition
	}
	if _, err := d.Manifests.Get(ctx, tenantID, mod.ID, inst.PreviousVersion); err != nil {
		return domain.Install{}, err
	}
	cur := inst.Version
	inst.Version = inst.PreviousVersion
	inst.PreviousVersion = cur
	inst.Status = domain.InstallRolledBack
	inst.UpdatedAt = d.now()
	if err := d.Installs.Save(ctx, inst); err != nil {
		return domain.Install{}, err
	}
	return inst, nil
}

func (d *Deps) GrantPermission(ctx context.Context, tenantID uuid.UUID, subjectID string, moduleID uuid.UUID, perm string) (domain.PermissionGrant, error) {
	if err := domain.ValidatePermissions([]string{perm}); err != nil {
		return domain.PermissionGrant{}, err
	}
	g := domain.PermissionGrant{
		ID: d.newID(), TenantID: tenantID, SubjectID: subjectID, ModuleID: moduleID,
		Permission: perm, Granted: true, GrantedAt: d.now(),
	}
	if err := d.Permissions.Save(ctx, g); err != nil {
		return domain.PermissionGrant{}, err
	}
	d.emit(ctx, tenantID, g.ID, domain.EventPermissionGranted, map[string]any{
		"moduleId": moduleID.String(), "permission": perm, "subjectId": subjectID,
	})
	return g, nil
}

func (d *Deps) RateModule(ctx context.Context, r domain.StoreRating) (domain.StoreRating, error) {
	if r.TenantID == uuid.Nil || r.ModuleID == uuid.Nil || r.SubjectID == "" || r.Score < 1 || r.Score > 5 {
		return domain.StoreRating{}, domain.ErrInvalidArgument
	}
	r.ID = d.newID()
	r.CreatedAt = d.now()
	if err := d.Ratings.Save(ctx, r); err != nil {
		return domain.StoreRating{}, err
	}
	if listing, err := d.Listings.GetByModule(ctx, r.TenantID, r.ModuleID); err == nil {
		listing.RatingAvg, listing.RatingCount = domain.UpdateRatingAvg(listing.RatingAvg, listing.RatingCount, r.Score)
		listing.UpdatedAt = d.now()
		_ = d.Listings.Save(ctx, listing)
	}
	return r, nil
}

func (d *Deps) AddWidget(ctx context.Context, w domain.WidgetPlacement) (domain.WidgetPlacement, error) {
	if w.TenantID == uuid.Nil || w.SubjectID == "" || w.ModuleID == uuid.Nil || w.Slot == "" {
		return domain.WidgetPlacement{}, domain.ErrInvalidArgument
	}
	mod, err := d.Modules.Get(ctx, w.TenantID, w.ModuleID)
	if err != nil {
		return domain.WidgetPlacement{}, err
	}
	if mod.Kind != domain.KindWidget && mod.Kind != domain.KindMiniApp {
		// allow mini apps to expose widgets
	}
	w.ID = d.newID()
	w.CreatedAt = d.now()
	if err := d.Widgets.Save(ctx, w); err != nil {
		return domain.WidgetPlacement{}, err
	}
	d.emit(ctx, w.TenantID, w.ID, domain.EventWidgetAdded, map[string]any{
		"slot": w.Slot, "moduleId": w.ModuleID.String(), "subjectId": w.SubjectID,
	})
	return w, nil
}

func (d *Deps) LaunchMiniApp(ctx context.Context, tenantID uuid.UUID, subjectID, moduleKey string) (domain.LaunchEvent, error) {
	mod, err := d.Modules.GetByKey(ctx, tenantID, moduleKey)
	if err != nil {
		return domain.LaunchEvent{}, err
	}
	inst, err := d.Installs.Get(ctx, tenantID, subjectID, mod.ID)
	if err != nil || inst.Status == domain.InstallRemoved {
		return domain.LaunchEvent{}, domain.ErrForbidden
	}
	ev := domain.LaunchEvent{
		ID: d.newID(), TenantID: tenantID, SubjectID: subjectID, ModuleID: mod.ID,
		Kind: mod.Kind, CreatedAt: d.now(),
	}
	_ = d.Launches.Save(ctx, ev)
	d.emit(ctx, tenantID, ev.ID, domain.EventMiniAppLaunched, map[string]any{
		"moduleKey": mod.Key, "subjectId": subjectID,
	})
	return ev, nil
}

func (d *Deps) UpsertMonetization(ctx context.Context, r domain.MonetizationRule) (domain.MonetizationRule, error) {
	if r.TenantID == uuid.Nil || r.ModuleID == uuid.Nil {
		return domain.MonetizationRule{}, domain.ErrInvalidArgument
	}
	if r.CommissionBps+r.PartnerShareBps > 10000 {
		return domain.MonetizationRule{}, domain.ErrInvalidArgument
	}
	if existing, err := d.Monetization.GetByModule(ctx, r.TenantID, r.ModuleID); err == nil {
		r.ID = existing.ID
	} else {
		r.ID = d.newID()
	}
	r.Active = true
	r.UpdatedAt = d.now()
	if err := d.Monetization.Save(ctx, r); err != nil {
		return domain.MonetizationRule{}, err
	}
	return r, nil
}

func (d *Deps) ResolveShell(ctx context.Context, tenantID uuid.UUID, subjectID, shellVersion string) (domain.ShellResolve, error) {
	if subjectID == "" {
		return domain.ShellResolve{}, domain.ErrInvalidArgument
	}
	if shellVersion == "" {
		shellVersion = d.ShellVersion
		if shellVersion == "" {
			shellVersion = "1.0.0"
		}
	}
	installs, err := d.Installs.ListBySubject(ctx, tenantID, subjectID)
	if err != nil {
		return domain.ShellResolve{}, err
	}
	resolved := []domain.ResolvedModule{}
	for _, inst := range installs {
		if inst.Status != domain.InstallActive && inst.Status != domain.InstallRolledBack {
			continue
		}
		mod, err := d.Modules.Get(ctx, tenantID, inst.ModuleID)
		if err != nil {
			continue
		}
		man, err := d.Manifests.Get(ctx, tenantID, mod.ID, inst.Version)
		if err != nil {
			man, err = d.Manifests.Latest(ctx, tenantID, mod.ID)
			if err != nil {
				continue
			}
		}
		if !domain.SemverCompatible(shellVersion, man.MinShellVersion) {
			continue
		}
		if d.LiveOps != nil {
			if ok, _ := d.LiveOps.ModuleEnabled(ctx, tenantID, mod.Key, subjectID); !ok {
				continue
			}
		}
		resolved = append(resolved, domain.ResolvedModule{
			Key: mod.Key, Kind: mod.Kind, Version: man.Version, EntryPoint: man.EntryPoint,
			Permissions: man.Permissions, Hooks: man.Hooks, BundleURI: man.BundleURI,
		})
	}
	widgets, _ := d.Widgets.ListBySubject(ctx, tenantID, subjectID)
	return domain.ShellResolve{Modules: resolved, Widgets: widgets, ShellHint: fmt.Sprintf("shell=%s", shellVersion)}, nil
}

func (d *Deps) Recommend(ctx context.Context, tenantID uuid.UUID, subjectID string, limit int) ([]string, error) {
	if d.AI != nil {
		return d.AI.RecommendModules(ctx, tenantID, subjectID, limit)
	}
	list, _ := d.Listings.List(ctx, tenantID)
	out := []string{}
	for _, l := range list {
		if l.Featured {
			if m, err := d.Modules.Get(ctx, tenantID, l.ModuleID); err == nil {
				out = append(out, m.Key)
			}
		}
		if limit > 0 && len(out) >= limit {
			break
		}
	}
	return out, nil
}

func (d *Deps) BrowseStore(ctx context.Context, tenantID uuid.UUID, category string) ([]map[string]any, error) {
	list, err := d.Listings.List(ctx, tenantID)
	if err != nil {
		return nil, err
	}
	out := []map[string]any{}
	for _, l := range list {
		m, err := d.Modules.Get(ctx, tenantID, l.ModuleID)
		if err != nil {
			continue
		}
		if category != "" && !strings.EqualFold(m.Category, category) {
			continue
		}
		out = append(out, map[string]any{
			"key": m.Key, "name": m.Name, "kind": m.Kind, "category": m.Category,
			"ratingAvg": l.RatingAvg, "installs": l.Installs, "featured": l.Featured,
			"priceMinor": l.PriceMinor, "currency": l.Currency,
		})
	}
	return out, nil
}

func (d *Deps) SearchModules(ctx context.Context, tenantID uuid.UUID, q string) ([]domain.Module, error) {
	list, err := d.Modules.List(ctx, tenantID)
	if err != nil {
		return nil, err
	}
	q = strings.ToLower(strings.TrimSpace(q))
	if q == "" {
		return list, nil
	}
	out := []domain.Module{}
	for _, m := range list {
		if strings.Contains(strings.ToLower(m.Key), q) ||
			strings.Contains(strings.ToLower(m.Name), q) ||
			strings.Contains(strings.ToLower(m.Category), q) ||
			strings.Contains(strings.ToLower(m.Description), q) {
			out = append(out, m)
		}
	}
	return out, nil
}

func (d *Deps) AdminStats(ctx context.Context, tenantID uuid.UUID) (map[string]any, error) {
	mods, _ := d.Modules.List(ctx, tenantID)
	listings, _ := d.Listings.List(ctx, tenantID)
	mini, plugins, widgets := 0, 0, 0
	for _, m := range mods {
		switch m.Kind {
		case domain.KindMiniApp:
			mini++
		case domain.KindPlugin:
			plugins++
		case domain.KindWidget:
			widgets++
		}
	}
	return map[string]any{
		"modules": len(mods), "miniApps": mini, "plugins": plugins, "widgets": widgets, "listings": len(listings),
	}, nil
}
