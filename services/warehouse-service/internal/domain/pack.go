package domain

import (
	"fmt"
	"math"
	"time"

	"github.com/google/uuid"
)

// PackSessionStatus is the pack workflow status.
type PackSessionStatus string

const (
	PackSessionStatusQueued    PackSessionStatus = "queued"
	PackSessionStatusClaimed   PackSessionStatus = "claimed"
	PackSessionStatusVerified  PackSessionStatus = "verified"
	PackSessionStatusSealed    PackSessionStatus = "sealed"
	PackSessionStatusLabeled   PackSessionStatus = "labeled"
	PackSessionStatusCompleted PackSessionStatus = "completed"
	PackSessionStatusCancelled PackSessionStatus = "cancelled"
)

func (s PackSessionStatus) Valid() bool {
	switch s {
	case PackSessionStatusQueued, PackSessionStatusClaimed, PackSessionStatusVerified,
		PackSessionStatusSealed, PackSessionStatusLabeled, PackSessionStatusCompleted,
		PackSessionStatusCancelled:
		return true
	default:
		return false
	}
}

// PackSession is a packing run at a station.
type PackSession struct {
	ID              uuid.UUID
	TenantID        uuid.UUID
	TaskID          uuid.UUID
	WarehouseID     uuid.UUID
	FulfillmentID   uuid.UUID
	StationID       *uuid.UUID
	PackerID        *uuid.UUID
	Status          PackSessionStatus
	ExpectedWeightG int64
	WeightTolerance int64
	ActualWeightG   *int64
	WeightG         *int
	LengthMM        *int
	WidthMM         *int
	HeightMM        *int
	Materials       []string
	ColdChain       bool
	Fragile         bool
	Hazard          bool
	SealedAt        *time.Time
	LabeledAt       *time.Time
	LabelID         *uuid.UUID
	LabelPayload    map[string]any
	StartedAt       *time.Time
	CompletedAt     *time.Time
	Metadata        map[string]any
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

// Validate checks pack session invariants.
func (s PackSession) Validate() error {
	if s.ID == uuid.Nil {
		return fmt.Errorf("%w: pack session id required", ErrInvalidArgument)
	}
	if s.TaskID == uuid.Nil {
		return fmt.Errorf("%w: task_id required", ErrInvalidArgument)
	}
	if s.Status != "" && !s.Status.Valid() {
		return fmt.Errorf("%w: invalid pack status %q", ErrInvalidArgument, s.Status)
	}
	return nil
}

// VerifyWeight checks actual weight within expected ± tolerance.
func (s *PackSession) VerifyWeight(actualWeightG int64) error {
	if s.Status != PackSessionStatusClaimed && s.Status != PackSessionStatusVerified {
		return fmt.Errorf("%w: verify from %s", ErrInvalidTransition, s.Status)
	}
	tol := s.WeightTolerance
	if tol <= 0 {
		tol = 50
	}
	diff := int64(math.Abs(float64(actualWeightG - s.ExpectedWeightG)))
	if diff > tol {
		return fmt.Errorf("%w: expected %d±%d got %d", ErrWeightMismatch, s.ExpectedWeightG, tol, actualWeightG)
	}
	s.ActualWeightG = &actualWeightG
	s.Status = PackSessionStatusVerified
	s.UpdatedAt = time.Now().UTC()
	return nil
}

// Seal marks the package sealed after weight verification.
func (s *PackSession) Seal() error {
	if s.Status != PackSessionStatusVerified {
		return fmt.Errorf("%w: seal from %s", ErrInvalidTransition, s.Status)
	}
	now := time.Now().UTC()
	s.Status = PackSessionStatusSealed
	s.SealedAt = &now
	s.UpdatedAt = now
	return nil
}

// AttachLabel stores label print payload after seal.
func (s *PackSession) AttachLabel(labelID uuid.UUID, payload map[string]any) error {
	if s.Status != PackSessionStatusSealed {
		return fmt.Errorf("%w: label from %s", ErrNotSealed, s.Status)
	}
	now := time.Now().UTC()
	s.Status = PackSessionStatusLabeled
	s.LabelID = &labelID
	s.LabeledAt = &now
	if payload == nil {
		payload = map[string]any{}
	}
	s.LabelPayload = payload
	s.UpdatedAt = now
	return nil
}

// Complete finishes the pack session.
func (s *PackSession) Complete() error {
	if s.Status != PackSessionStatusLabeled && s.Status != PackSessionStatusSealed {
		return fmt.Errorf("%w: complete from %s", ErrInvalidTransition, s.Status)
	}
	now := time.Now().UTC()
	s.Status = PackSessionStatusCompleted
	s.CompletedAt = &now
	s.UpdatedAt = now
	return nil
}

// IsSealed reports whether the package is sealed.
func (s PackSession) IsSealed() bool {
	return s.SealedAt != nil || s.Status == PackSessionStatusSealed ||
		s.Status == PackSessionStatusLabeled || s.Status == PackSessionStatusCompleted
}
