package repository

import (
	"context"
	"database/sql"
	"strings"

	"recommand/internal/domain"
)

type TaskRepo struct {
	db *sql.DB
}

func NewTaskRepo(db *sql.DB) *TaskRepo {
	return &TaskRepo{db: db}
}

func (r *TaskRepo) Create(ctx context.Context, t *domain.CrawlTask) error {
	_, err := r.db.ExecContext(ctx, `INSERT INTO crawl_tasks (task_id, source_id, source_name, mode, since, max_pages, status, progress, pages_crawled, articles_found, articles_saved, duplicates_skipped, errors, started_at, completed_at, error_message, created_by) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17)`, t.TaskID, t.SourceID, t.SourceName, t.Mode, t.Since, t.MaxPages, t.Status, t.Progress, t.PagesCrawled, t.ArticlesFound, t.ArticlesSaved, t.DuplicatesSkipped, t.Errors, t.StartedAt, t.CompletedAt, t.ErrorMessage, t.CreatedBy)
	return err
}

func (r *TaskRepo) GetByID(ctx context.Context, id string) (*domain.CrawlTask, error) {
	row := r.db.QueryRowContext(ctx, `SELECT task_id, source_id, source_name, mode, since, max_pages, status, progress, pages_crawled, articles_found, articles_saved, duplicates_skipped, errors, started_at, completed_at, error_message, created_by, created_at, updated_at FROM crawl_tasks WHERE task_id=$1`, id)
	var t domain.CrawlTask
	if err := row.Scan(&t.TaskID, &t.SourceID, &t.SourceName, &t.Mode, &t.Since, &t.MaxPages, &t.Status, &t.Progress, &t.PagesCrawled, &t.ArticlesFound, &t.ArticlesSaved, &t.DuplicatesSkipped, &t.Errors, &t.StartedAt, &t.CompletedAt, &t.ErrorMessage, &t.CreatedBy, &t.CreatedAt, &t.UpdatedAt); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &t, nil
}

func (r *TaskRepo) List(ctx context.Context, sourceID *int64, status *domain.CrawlStatus) ([]domain.CrawlTask, error) {
	query := `SELECT task_id, source_id, source_name, mode, since, max_pages, status, progress, pages_crawled, articles_found, articles_saved, duplicates_skipped, errors, started_at, completed_at, error_message, created_by, created_at, updated_at FROM crawl_tasks`
	args := []any{}
	conditions := []string{}
	if sourceID != nil {
		conditions = append(conditions, "source_id=$1")
		args = append(args, *sourceID)
	}
	if status != nil {
		placeholder := "$1"
		if len(args) == 1 {
			placeholder = "$2"
		}
		conditions = append(conditions, "status="+placeholder)
		args = append(args, *status)
	}
	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}
	query += " ORDER BY created_at DESC LIMIT 100"

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var res []domain.CrawlTask
	for rows.Next() {
		var t domain.CrawlTask
		if err := rows.Scan(&t.TaskID, &t.SourceID, &t.SourceName, &t.Mode, &t.Since, &t.MaxPages, &t.Status, &t.Progress, &t.PagesCrawled, &t.ArticlesFound, &t.ArticlesSaved, &t.DuplicatesSkipped, &t.Errors, &t.StartedAt, &t.CompletedAt, &t.ErrorMessage, &t.CreatedBy, &t.CreatedAt, &t.UpdatedAt); err != nil {
			return nil, err
		}
		res = append(res, t)
	}
	return res, rows.Err()
}

func (r *TaskRepo) UpdateStatus(ctx context.Context, id string, status domain.CrawlStatus) error {
	_, err := r.db.ExecContext(ctx, `UPDATE crawl_tasks SET status=$1, updated_at=NOW() WHERE task_id=$2`, status, id)
	return err
}

// UpdateStatusAndProgress updates status, progress and pages_crawled for a task.
func (r *TaskRepo) UpdateStatusAndProgress(ctx context.Context, id string, status domain.CrawlStatus, progress float64, pages int) error {
	_, err := r.db.ExecContext(ctx, `UPDATE crawl_tasks SET status=$1, progress=$2, pages_crawled=$3, updated_at=NOW() WHERE task_id=$4`, status, progress, pages, id)
	return err
}
