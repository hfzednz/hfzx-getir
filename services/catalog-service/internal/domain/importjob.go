package domain

import (
	"fmt"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"
)

// ImportJobKind distinguishes import vs export jobs.
type ImportJobKind string

const (
	ImportJobKindImport ImportJobKind = "import"
	ImportJobKindExport ImportJobKind = "export"
)

func (k ImportJobKind) Valid() bool {
	switch k {
	case ImportJobKindImport, ImportJobKindExport:
		return true
	default:
		return false
	}
}

// ImportJobStatus is the async job lifecycle.
type ImportJobStatus string

const (
	ImportJobStatusPending    ImportJobStatus = "pending"
	ImportJobStatusValidating ImportJobStatus = "validating"
	ImportJobStatusRunning    ImportJobStatus = "running"
	ImportJobStatusCompleted  ImportJobStatus = "completed"
	ImportJobStatusFailed     ImportJobStatus = "failed"
	ImportJobStatusCancelled  ImportJobStatus = "cancelled"
	ImportJobStatusPartial    ImportJobStatus = "partial"
)

func (s ImportJobStatus) Valid() bool {
	switch s {
	case ImportJobStatusPending, ImportJobStatusValidating, ImportJobStatusRunning,
		ImportJobStatusCompleted, ImportJobStatusFailed, ImportJobStatusCancelled,
		ImportJobStatusPartial:
		return true
	default:
		return false
	}
}

var importJobTransitions = map[ImportJobStatus][]ImportJobStatus{
	ImportJobStatusPending: {
		ImportJobStatusValidating, ImportJobStatusRunning, ImportJobStatusCancelled,
	},
	ImportJobStatusValidating: {
		ImportJobStatusRunning, ImportJobStatusFailed, ImportJobStatusCancelled,
	},
	ImportJobStatusRunning: {
		ImportJobStatusCompleted, ImportJobStatusFailed, ImportJobStatusPartial, ImportJobStatusCancelled,
	},
	ImportJobStatusCompleted: {},
	ImportJobStatusFailed:    {},
	ImportJobStatusCancelled: {},
	ImportJobStatusPartial:   {},
}

// CanTransitionTo reports whether from → to is allowed for import jobs.
func (s ImportJobStatus) CanTransitionTo(to ImportJobStatus) bool {
	if !s.Valid() || !to.Valid() {
		return false
	}
	for _, next := range importJobTransitions[s] {
		if next == to {
			return true
		}
	}
	return false
}

const (
	maxSourceFormatLen = 32
	maxURILen          = 2048
)

// ImportJob tracks async catalog import/export work.
type ImportJob struct {
	ID            uuid.UUID
	TenantID      uuid.UUID
	Kind          ImportJobKind
	Status        ImportJobStatus
	SourceFormat  string
	SourceURI     string
	ResultURI     string
	TotalRows     int
	ProcessedRows int
	SuccessRows   int
	ErrorRows     int
	Errors        []map[string]any
	Options       map[string]any
	CreatedBy     *uuid.UUID
	StartedAt     *time.Time
	FinishedAt    *time.Time
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

// Validate checks structural invariants.
func (j ImportJob) Validate() error {
	if j.ID == uuid.Nil {
		return fmt.Errorf("%w: import_job id required", ErrInvalidArgument)
	}
	if j.TenantID == uuid.Nil {
		return fmt.Errorf("%w: tenant_id required", ErrInvalidArgument)
	}
	if !j.Kind.Valid() {
		return fmt.Errorf("%w: invalid import job kind %q", ErrInvalidArgument, j.Kind)
	}
	if !j.Status.Valid() {
		return fmt.Errorf("%w: invalid import job status %q", ErrInvalidArgument, j.Status)
	}
	if utf8.RuneCountInString(j.SourceFormat) > maxSourceFormatLen {
		return fmt.Errorf("%w: source_format too long", ErrInvalidArgument)
	}
	if utf8.RuneCountInString(j.SourceURI) > maxURILen {
		return fmt.Errorf("%w: source_uri too long", ErrInvalidArgument)
	}
	if utf8.RuneCountInString(j.ResultURI) > maxURILen {
		return fmt.Errorf("%w: result_uri too long", ErrInvalidArgument)
	}
	if j.TotalRows < 0 || j.ProcessedRows < 0 || j.SuccessRows < 0 || j.ErrorRows < 0 {
		return fmt.Errorf("%w: row counters must be >= 0", ErrInvalidArgument)
	}
	if j.ProcessedRows > j.TotalRows && j.TotalRows > 0 {
		return fmt.Errorf("%w: processed_rows exceeds total_rows", ErrInvariant)
	}
	if j.Errors == nil {
		return fmt.Errorf("%w: errors required (may be empty)", ErrInvalidArgument)
	}
	if j.Options == nil {
		return fmt.Errorf("%w: options required", ErrInvalidArgument)
	}
	return nil
}

// ValidateImportJobStatusTransition returns ErrInvalidTransition when disallowed.
func ValidateImportJobStatusTransition(from, to ImportJobStatus) error {
	if from == to {
		return nil
	}
	if !from.CanTransitionTo(to) {
		return fmt.Errorf("%w: %s → %s", ErrInvalidTransition, from, to)
	}
	return nil
}
