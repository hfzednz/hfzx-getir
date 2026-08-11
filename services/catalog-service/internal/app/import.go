package app

import (
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"strings"

	"github.com/google/uuid"
	"github.com/nexora/catalog-service/internal/domain"
)

// Required CSV columns for product import validation.
var ImportRequiredColumns = []string{"slug", "sku_code", "title", "lang"}

// ValidateImportCSV validates CSV headers and row shape without persisting.
func (d *Deps) ValidateImportCSV(ctx context.Context, tenantID uuid.UUID, r io.Reader) (domain.ImportJob, error) {
	now := d.now()
	job := domain.ImportJob{
		ID:           d.newID(),
		TenantID:     tenantID,
		Kind:         domain.ImportJobKindImport,
		Status:       domain.ImportJobStatusValidating,
		SourceFormat: "csv",
		Errors:       []map[string]any{},
		Options:      map[string]any{},
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	if err := job.Validate(); err != nil {
		return domain.ImportJob{}, err
	}
	if err := d.ImportJobs.Create(ctx, job); err != nil {
		return domain.ImportJob{}, err
	}

	reader := csv.NewReader(r)
	reader.TrimLeadingSpace = true
	header, err := reader.Read()
	if err != nil {
		job.Status = domain.ImportJobStatusFailed
		job.Errors = append(job.Errors, map[string]any{"row": 0, "error": err.Error()})
		job.UpdatedAt = d.now()
		_ = d.ImportJobs.Update(ctx, job)
		return job, nil
	}
	colIndex := map[string]int{}
	for i, h := range header {
		colIndex[strings.ToLower(strings.TrimSpace(h))] = i
	}
	for _, req := range ImportRequiredColumns {
		if _, ok := colIndex[req]; !ok {
			job.Errors = append(job.Errors, map[string]any{
				"row": 1, "error": fmt.Sprintf("missing required column %q", req),
			})
		}
	}
	rowNum := 1
	for {
		row, err := reader.Read()
		if err == io.EOF {
			break
		}
		rowNum++
		if err != nil {
			job.Errors = append(job.Errors, map[string]any{"row": rowNum, "error": err.Error()})
			continue
		}
		job.TotalRows++
		if slugIdx, ok := colIndex["slug"]; ok && slugIdx < len(row) {
			if err := domain.ValidateSlug(strings.TrimSpace(row[slugIdx])); err != nil {
				job.Errors = append(job.Errors, map[string]any{"row": rowNum, "field": "slug", "error": err.Error()})
				job.ErrorRows++
				continue
			}
		}
		job.SuccessRows++
	}
	job.ProcessedRows = job.TotalRows
	if len(job.Errors) > 0 {
		job.Status = domain.ImportJobStatusFailed
	} else {
		job.Status = domain.ImportJobStatusCompleted
	}
	finished := d.now()
	job.FinishedAt = &finished
	job.UpdatedAt = finished
	_ = d.ImportJobs.Update(ctx, job)
	return job, nil
}

// GetImportJob returns an import job by id.
func (d *Deps) GetImportJob(ctx context.Context, tenantID, jobID uuid.UUID) (domain.ImportJob, error) {
	return d.ImportJobs.GetByID(ctx, tenantID, jobID)
}

// ListImportJobs lists import jobs for a tenant.
func (d *Deps) ListImportJobs(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]domain.ImportJob, error) {
	if limit <= 0 {
		limit = 20
	}
	return d.ImportJobs.List(ctx, tenantID, limit, offset)
}
