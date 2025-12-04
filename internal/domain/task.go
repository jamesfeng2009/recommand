package domain

import "time"

type CrawlMode string

const (
	CrawlModeFull        CrawlMode = "full"
	CrawlModeIncremental CrawlMode = "incremental"
)

type CrawlStatus string

const (
	StatusPending   CrawlStatus = "pending"
	StatusRunning   CrawlStatus = "running"
	StatusCompleted CrawlStatus = "completed"
	StatusFailed    CrawlStatus = "failed"
	StatusStopped   CrawlStatus = "stopped"
)

type CrawlTask struct {
	TaskID            string      `db:"task_id" json:"task_id"`
	SourceID          int64       `db:"source_id" json:"source_id"`
	SourceName        string      `db:"source_name" json:"source_name"`
	Mode              CrawlMode   `db:"mode" json:"mode"`
	Since             *time.Time  `db:"since" json:"since,omitempty"`
	MaxPages          *int        `db:"max_pages" json:"max_pages,omitempty"`
	Status            CrawlStatus `db:"status" json:"status"`
	Progress          float64     `db:"progress" json:"progress"`
	PagesCrawled      int         `db:"pages_crawled" json:"pages_crawled"`
	ArticlesFound     int         `db:"articles_found" json:"articles_found"`
	ArticlesSaved     int         `db:"articles_saved" json:"articles_saved"`
	DuplicatesSkipped int         `db:"duplicates_skipped" json:"duplicates_skipped"`
	Errors            int         `db:"errors" json:"errors"`
	StartedAt         *time.Time  `db:"started_at" json:"started_at,omitempty"`
	CompletedAt       *time.Time  `db:"completed_at" json:"completed_at,omitempty"`
	ErrorMessage      *string     `db:"error_message" json:"error_message,omitempty"`
	CreatedBy         *int64      `db:"created_by" json:"created_by,omitempty"`
	CreatedAt         time.Time   `db:"created_at" json:"created_at"`
	UpdatedAt         time.Time   `db:"updated_at" json:"updated_at"`
}
