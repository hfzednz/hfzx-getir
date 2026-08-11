package domain

import (
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

// MaxHistoryPerSubject caps retained location history rows per subject.
const MaxHistoryPerSubject = 100

// SubjectType classifies a history subject.
type SubjectType string

const (
	SubjectDevice    SubjectType = "device"
	SubjectCustomer  SubjectType = "customer"
	SubjectCourier   SubjectType = "courier"
	SubjectWarehouse SubjectType = "warehouse"
)

// Valid reports whether the subject type is recognized.
func (s SubjectType) Valid() bool {
	switch s {
	case SubjectDevice, SubjectCustomer, SubjectCourier, SubjectWarehouse:
		return true
	default:
		return false
	}
}

// LocationHistory is a privacy-scoped historical location ping.
type LocationHistory struct {
	ID          uuid.UUID
	TenantID    uuid.UUID
	SubjectType SubjectType
	SubjectID   string
	Lat         float64
	Lng         float64
	RecordedAt  time.Time
}

// Validate checks history invariants.
func (h LocationHistory) Validate() error {
	if h.ID == uuid.Nil {
		return fmt.Errorf("%w: history id required", ErrInvalidArgument)
	}
	if h.TenantID == uuid.Nil {
		return fmt.Errorf("%w: tenant_id required", ErrInvalidArgument)
	}
	if !h.SubjectType.Valid() {
		return fmt.Errorf("%w: invalid subject_type %q", ErrInvalidArgument, h.SubjectType)
	}
	if strings.TrimSpace(h.SubjectID) == "" {
		return fmt.Errorf("%w: subject_id required", ErrInvalidArgument)
	}
	if !ValidLatLng(h.Lat, h.Lng) {
		return fmt.Errorf("%w: lat/lng out of range", ErrInvalidArgument)
	}
	if h.RecordedAt.IsZero() {
		return fmt.Errorf("%w: recorded_at required", ErrInvalidArgument)
	}
	return nil
}
