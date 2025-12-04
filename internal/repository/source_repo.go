package repository

import (
	"context"
	"database/sql"

	"recommand/internal/domain"
)

type SourceRepo struct {
	db *sql.DB
}

func NewSourceRepo(db *sql.DB) *SourceRepo {
	return &SourceRepo{db: db}
}

func (r *SourceRepo) List(ctx context.Context) ([]domain.NewsSource, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT id, name, code, base_url, language, category, enabled, crawl_interval_minutes, max_concurrency, last_crawl_at, last_crawl_status, created_at, updated_at FROM news_sources ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var res []domain.NewsSource
	for rows.Next() {
		var ns domain.NewsSource
		if err := rows.Scan(&ns.ID, &ns.Name, &ns.Code, &ns.BaseURL, &ns.Language, &ns.Category, &ns.Enabled, &ns.CrawlIntervalMin, &ns.MaxConcurrency, &ns.LastCrawlAt, &ns.LastCrawlStatus, &ns.CreatedAt, &ns.UpdatedAt); err != nil {
			return nil, err
		}
		res = append(res, ns)
	}
	return res, rows.Err()
}

func (r *SourceRepo) GetByID(ctx context.Context, id int64) (*domain.NewsSource, error) {
	row := r.db.QueryRowContext(ctx, `SELECT id, name, code, base_url, language, category, enabled, crawl_interval_minutes, max_concurrency, last_crawl_at, last_crawl_status, created_at, updated_at FROM news_sources WHERE id=$1`, id)
	var ns domain.NewsSource
	if err := row.Scan(&ns.ID, &ns.Name, &ns.Code, &ns.BaseURL, &ns.Language, &ns.Category, &ns.Enabled, &ns.CrawlIntervalMin, &ns.MaxConcurrency, &ns.LastCrawlAt, &ns.LastCrawlStatus, &ns.CreatedAt, &ns.UpdatedAt); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &ns, nil
}

func (r *SourceRepo) Create(ctx context.Context, ns *domain.NewsSource) error {
	row := r.db.QueryRowContext(ctx, `INSERT INTO news_sources (name, code, base_url, language, category, enabled, crawl_interval_minutes, max_concurrency) VALUES ($1,$2,$3,$4,$5,$6,$7,$8) RETURNING id, created_at, updated_at`, ns.Name, ns.Code, ns.BaseURL, ns.Language, ns.Category, ns.Enabled, ns.CrawlIntervalMin, ns.MaxConcurrency)
	return row.Scan(&ns.ID, &ns.CreatedAt, &ns.UpdatedAt)
}

func (r *SourceRepo) Update(ctx context.Context, ns *domain.NewsSource) error {
	_, err := r.db.ExecContext(ctx, `UPDATE news_sources SET name=$1, code=$2, base_url=$3, language=$4, category=$5, enabled=$6, crawl_interval_minutes=$7, max_concurrency=$8, updated_at=NOW() WHERE id=$9`, ns.Name, ns.Code, ns.BaseURL, ns.Language, ns.Category, ns.Enabled, ns.CrawlIntervalMin, ns.MaxConcurrency, ns.ID)
	return err
}

func (r *SourceRepo) UpdateEnabled(ctx context.Context, id int64, enabled bool) error {
	_, err := r.db.ExecContext(ctx, `UPDATE news_sources SET enabled=$1, updated_at=NOW() WHERE id=$2`, enabled, id)
	return err
}
