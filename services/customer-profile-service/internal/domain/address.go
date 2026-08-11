package domain

import (
	"fmt"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"
)

// AddressLabel classifies a saved delivery address.
type AddressLabel string

const (
	AddressLabelHome     AddressLabel = "home"
	AddressLabelWork     AddressLabel = "work"
	AddressLabelVacation AddressLabel = "vacation"
	AddressLabelCustom   AddressLabel = "custom"
)

func (l AddressLabel) Valid() bool {
	switch l {
	case AddressLabelHome, AddressLabelWork, AddressLabelVacation, AddressLabelCustom:
		return true
	default:
		return false
	}
}

const (
	maxLine1Len       = 200
	maxBuildingLen    = 80
	maxApartmentLen   = 40
	maxEntranceLen    = 40
	maxFloorLen       = 20
	maxDoorLen        = 20
	maxAddressNotes   = 500
	maxCustomLabelLen = 80
)

// Address is a delivery location belonging to a profile.
type Address struct {
	ID              uuid.UUID
	ProfileID       uuid.UUID
	TenantID        uuid.UUID
	Label           AddressLabel
	CustomLabel     string
	Line1           string
	Building        string
	Apartment       string
	Entrance        string
	Floor           string
	Door            string
	Notes           string
	Lat             float64
	Lng             float64
	CityID          *uuid.UUID
	ZoneValidatedAt *time.Time
	IsDefault       bool
	CreatedAt       time.Time
	UpdatedAt       time.Time
	DeletedAt       *time.Time
}

// Validate checks structural invariants including WGS84 lat/lng bounds.
func (a Address) Validate() error {
	if a.ID == uuid.Nil {
		return fmt.Errorf("%w: address id required", ErrInvalidArgument)
	}
	if a.ProfileID == uuid.Nil {
		return fmt.Errorf("%w: profile_id required", ErrInvalidArgument)
	}
	if a.TenantID == uuid.Nil {
		return fmt.Errorf("%w: tenant_id required", ErrInvalidArgument)
	}
	if !a.Label.Valid() {
		return fmt.Errorf("%w: invalid address label %q", ErrInvalidArgument, a.Label)
	}
	if a.Label == AddressLabelCustom && strings.TrimSpace(a.CustomLabel) == "" {
		return fmt.Errorf("%w: custom_label required when label is custom", ErrInvalidArgument)
	}
	if utf8.RuneCountInString(a.CustomLabel) > maxCustomLabelLen {
		return fmt.Errorf("%w: custom_label too long", ErrInvalidArgument)
	}
	if strings.TrimSpace(a.Line1) == "" {
		return fmt.Errorf("%w: line1 required", ErrInvalidArgument)
	}
	if utf8.RuneCountInString(a.Line1) > maxLine1Len {
		return fmt.Errorf("%w: line1 too long", ErrInvalidArgument)
	}
	if utf8.RuneCountInString(a.Building) > maxBuildingLen {
		return fmt.Errorf("%w: building too long", ErrInvalidArgument)
	}
	if utf8.RuneCountInString(a.Apartment) > maxApartmentLen {
		return fmt.Errorf("%w: apartment too long", ErrInvalidArgument)
	}
	if utf8.RuneCountInString(a.Entrance) > maxEntranceLen {
		return fmt.Errorf("%w: entrance too long", ErrInvalidArgument)
	}
	if utf8.RuneCountInString(a.Floor) > maxFloorLen {
		return fmt.Errorf("%w: floor too long", ErrInvalidArgument)
	}
	if utf8.RuneCountInString(a.Door) > maxDoorLen {
		return fmt.Errorf("%w: door too long", ErrInvalidArgument)
	}
	if utf8.RuneCountInString(a.Notes) > maxAddressNotes {
		return fmt.Errorf("%w: notes too long", ErrInvalidArgument)
	}
	if a.Lat < -90 || a.Lat > 90 {
		return fmt.Errorf("%w: lat out of range [-90,90]", ErrInvalidArgument)
	}
	if a.Lng < -180 || a.Lng > 180 {
		return fmt.Errorf("%w: lng out of range [-180,180]", ErrInvalidArgument)
	}
	if a.CityID != nil && *a.CityID == uuid.Nil {
		return fmt.Errorf("%w: city_id must not be nil uuid", ErrInvalidArgument)
	}
	return nil
}

// IsZoneValidated reports whether the address has been validated by geofence.
func (a Address) IsZoneValidated() bool {
	return a.ZoneValidatedAt != nil
}
