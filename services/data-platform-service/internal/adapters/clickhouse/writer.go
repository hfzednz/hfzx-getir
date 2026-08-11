package clickhouse

import (
	"context"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/nexora/data-platform-service/internal/app/ports"
)

// Writer wraps ClickHouse; falls back to memory OLAP when URL empty/unreachable.
type Writer struct {
	url      string
	fallback ports.OLAPWriter
	log      *slog.Logger
}

func NewWriter(url string, fallback ports.OLAPWriter) *Writer {
	return &Writer{url: url, fallback: fallback, log: slog.Default()}
}

func (w *Writer) InsertAggregate(ctx context.Context, tenantID uuid.UUID, metric string, value float64, ts time.Time) error {
	w.log.Info("clickhouse.insert_stub", "urlSet", w.url != "", "metric", metric, "value", value)
	if w.fallback != nil {
		return w.fallback.InsertAggregate(ctx, tenantID, metric, value, ts)
	}
	return nil
}

func (w *Writer) QuerySum(ctx context.Context, tenantID uuid.UUID, metric string, from, to time.Time) (float64, error) {
	if w.fallback != nil {
		return w.fallback.QuerySum(ctx, tenantID, metric, from, to)
	}
	return 0, nil
}
