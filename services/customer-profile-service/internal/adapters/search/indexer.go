// Package search provides an OpenSearch profile indexer.
package search

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/nexora/customer-profile-service/internal/app/ports"
	"github.com/nexora/customer-profile-service/internal/domain"
)

const indexPrefix = "nexora-profiles-"

// Indexer indexes customer profiles in OpenSearch when URL is set.
type Indexer struct {
	URL        string
	Log        *slog.Logger
	HTTPClient *http.Client
}

// NewIndexer constructs an OpenSearch indexer. Empty URL → no-op with debug logs.
func NewIndexer(searchURL string, log *slog.Logger) *Indexer {
	if log == nil {
		log = slog.Default()
	}
	searchURL = strings.TrimRight(strings.TrimSpace(searchURL), "/")
	i := &Indexer{
		URL:        searchURL,
		Log:        log,
		HTTPClient: &http.Client{Timeout: 10 * time.Second},
	}
	if searchURL != "" {
		log.Info("opensearch.profiles.http", "url", searchURL)
	} else {
		log.Info("opensearch.profiles.noop", "note", "SEARCH_URL/OPENSEARCH_URL empty")
	}
	return i
}

func (i *Indexer) indexName(tenantID uuid.UUID) string {
	return indexPrefix + tenantID.String()
}

// IndexProfile upserts a profile document into the search index.
func (i *Indexer) IndexProfile(ctx context.Context, p domain.CustomerProfile) error {
	if i == nil || i.URL == "" {
		if i != nil && i.Log != nil {
			i.Log.Debug("search.noop.index_profile", "profileId", p.ID.String(), "tenantId", p.TenantID.String())
		}
		return nil
	}
	body, err := json.Marshal(map[string]any{
		"id":          p.ID.String(),
		"tenantId":    p.TenantID.String(),
		"principalId": p.PrincipalID.String(),
		"displayName": p.DisplayName,
		"fullName":    p.FullName,
		"nickname":    p.Nickname,
		"city":        p.City,
		"countryCode": p.CountryCode,
		"language":    p.Language,
		"status":      string(p.Status),
		"updatedAt":   p.UpdatedAt.UTC().Format(time.RFC3339),
	})
	if err != nil {
		return err
	}
	path := fmt.Sprintf("/%s/_doc/%s", i.indexName(p.TenantID), p.ID.String())
	if err := i.doJSON(ctx, http.MethodPut, path, body); err != nil {
		return err
	}
	i.Log.Debug("search.index_profile", "profileId", p.ID.String(), "index", i.indexName(p.TenantID))
	return nil
}

// DeleteProfile removes a profile document.
func (i *Indexer) DeleteProfile(ctx context.Context, tenantID uuid.UUID, profileID uuid.UUID) error {
	if i == nil || i.URL == "" {
		if i != nil && i.Log != nil {
			i.Log.Debug("search.noop.delete_profile", "profileId", profileID.String())
		}
		return nil
	}
	path := fmt.Sprintf("/%s/_doc/%s", i.indexName(tenantID), profileID.String())
	if err := i.doJSON(ctx, http.MethodDelete, path, nil); err != nil {
		i.Log.Warn("search.delete_profile", "err", err, "profileId", profileID.String())
	}
	return nil
}

func (i *Indexer) doJSON(ctx context.Context, method, path string, body []byte) error {
	var rdr io.Reader
	if body != nil {
		rdr = bytes.NewReader(body)
	}
	req, err := http.NewRequestWithContext(ctx, method, i.URL+path, rdr)
	if err != nil {
		return err
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	res, err := i.HTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	raw, _ := io.ReadAll(io.LimitReader(res.Body, 4<<20))
	if res.StatusCode >= 300 && res.StatusCode != http.StatusNotFound {
		return fmt.Errorf("opensearch %s %s: status %d: %s", method, path, res.StatusCode, string(raw))
	}
	return nil
}

var _ ports.ProfileSearchIndexer = (*Indexer)(nil)
