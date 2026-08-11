package postgres

import "database/sql"

// Repos groups notification persistence adapters.
type Repos struct {
	Templates   *TemplateRepo
	Messages    *MessageRepo
	Preferences *PreferenceRepo
	Devices     *DeviceRepo
	Inbox       *InboxRepo
	Schedules   *ScheduleRepo
	Deliveries  *DeliveryRepo
	Outbox      *OutboxRepo
}

// NewRepos constructs postgres-backed repositories.
func NewRepos(db *sql.DB) *Repos {
	return &Repos{
		Templates:   &TemplateRepo{DB: db},
		Messages:    &MessageRepo{DB: db},
		Preferences: &PreferenceRepo{DB: db},
		Devices:     &DeviceRepo{DB: db},
		Inbox:       &InboxRepo{DB: db},
		Schedules:   &ScheduleRepo{DB: db},
		Deliveries:  &DeliveryRepo{DB: db},
		Outbox:      &OutboxRepo{DB: db},
	}
}
