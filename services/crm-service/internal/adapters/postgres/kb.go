package postgres

import (
	"context"
	"database/sql"
	"strings"

	"github.com/google/uuid"
	"github.com/nexora/crm-service/internal/app/ports"
	"github.com/nexora/crm-service/internal/domain"
)

// KBRepo persists knowledge-base articles and versions.
type KBRepo struct{ DB *sql.DB }

const articleSelect = `
	SELECT id, tenant_id, slug, title, body, locale, tags, status, version, published_at, created_at, updated_at
	FROM kb_articles`

func (r *KBRepo) SaveArticle(ctx context.Context, a domain.Article) error {
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO kb_articles (
			id, tenant_id, slug, title, body, locale, tags, status, version, published_at, created_at, updated_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)
		ON CONFLICT (id) DO UPDATE SET
			slug=EXCLUDED.slug,
			title=EXCLUDED.title,
			body=EXCLUDED.body,
			locale=EXCLUDED.locale,
			tags=EXCLUDED.tags,
			status=EXCLUDED.status,
			version=EXCLUDED.version,
			published_at=EXCLUDED.published_at,
			updated_at=EXCLUDED.updated_at`,
		a.ID, a.TenantID, a.Slug, a.Title, a.Body, a.Locale, TextArray(a.Tags),
		a.Status, a.Version, nullTime(a.PublishedAt), a.CreatedAt.UTC(), a.UpdatedAt.UTC())
	return mapUniqueViolation(err)
}

func (r *KBRepo) GetArticle(ctx context.Context, tenantID, id uuid.UUID) (domain.Article, error) {
	return r.scanArticle(r.DB.QueryRowContext(ctx, articleSelect+` WHERE id=$1 AND tenant_id=$2`, id, tenantID))
}

func (r *KBRepo) GetBySlug(ctx context.Context, tenantID uuid.UUID, slug string) (domain.Article, error) {
	return r.scanArticle(r.DB.QueryRowContext(ctx, articleSelect+` WHERE tenant_id=$1 AND slug=$2`, tenantID, slug))
}

func (r *KBRepo) Search(ctx context.Context, tenantID uuid.UUID, query string) ([]domain.Article, error) {
	q := strings.ToLower(strings.TrimSpace(query))
	rows, err := r.DB.QueryContext(ctx, articleSelect+`
		WHERE tenant_id=$1 AND status=$2
		ORDER BY updated_at DESC`, tenantID, domain.ArticleStatusPublished)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.Article{}
	for rows.Next() {
		a, err := scanArticleRow(rows)
		if err != nil {
			return nil, err
		}
		if q == "" ||
			strings.Contains(strings.ToLower(a.Title), q) ||
			strings.Contains(strings.ToLower(a.Body), q) ||
			strings.Contains(strings.ToLower(a.Slug), q) {
			out = append(out, a)
			continue
		}
		for _, tag := range a.Tags {
			if strings.Contains(strings.ToLower(tag), q) {
				out = append(out, a)
				break
			}
		}
	}
	return out, rows.Err()
}

func (r *KBRepo) SaveVersion(ctx context.Context, v domain.ArticleVersion) error {
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO kb_versions (id, tenant_id, article_id, version, title, body, created_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7)`,
		v.ID, v.TenantID, v.ArticleID, v.Version, v.Title, v.Body, v.CreatedAt.UTC())
	return err
}

func (r *KBRepo) scanArticle(row scannable) (domain.Article, error) {
	a, err := scanArticleRow(row)
	if err != nil {
		return domain.Article{}, mapNotFound(err)
	}
	return a, nil
}

func scanArticleRow(row scannable) (domain.Article, error) {
	var a domain.Article
	var tags TextArray
	var published sql.NullTime
	err := row.Scan(
		&a.ID, &a.TenantID, &a.Slug, &a.Title, &a.Body, &a.Locale, &tags,
		&a.Status, &a.Version, &published, &a.CreatedAt, &a.UpdatedAt)
	if err != nil {
		return domain.Article{}, err
	}
	a.Tags = []string(tags)
	a.PublishedAt = scanNullTime(published)
	a.CreatedAt = a.CreatedAt.UTC()
	a.UpdatedAt = a.UpdatedAt.UTC()
	return a, nil
}

var _ ports.KBRepo = (*KBRepo)(nil)
