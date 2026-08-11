package postgres

import "database/sql"

type Repos struct {
	Modules      *ModuleRepo
	Manifests    *ManifestRepo
	Installs     *InstallRepo
	Permissions  *PermissionRepo
	Listings     *ListingRepo
	Ratings      *RatingRepo
	Widgets      *WidgetRepo
	Monetization *MonetizationRepo
	Launches     *LaunchRepo
	Outbox       *OutboxRepo
}

func NewRepos(db *sql.DB) *Repos {
	return &Repos{
		Modules: &ModuleRepo{DB: db}, Manifests: &ManifestRepo{DB: db}, Installs: &InstallRepo{DB: db},
		Permissions: &PermissionRepo{DB: db}, Listings: &ListingRepo{DB: db}, Ratings: &RatingRepo{DB: db},
		Widgets: &WidgetRepo{DB: db}, Monetization: &MonetizationRepo{DB: db}, Launches: &LaunchRepo{DB: db},
		Outbox: &OutboxRepo{DB: db},
	}
}
