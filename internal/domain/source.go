package domain

import "time"

type NewsSource struct {
	ID               int64      `db:"id" json:"id"`
	Name             string     `db:"name" json:"name"`
	Code             string     `db:"code" json:"code"`
	BaseURL          string     `db:"base_url" json:"base_url"`
	Language         string     `db:"language" json:"language"`
	Category         string     `db:"category" json:"category"`
	Enabled          bool       `db:"enabled" json:"enabled"`
	CrawlIntervalMin int        `db:"crawl_interval_minutes" json:"crawl_interval_minutes"`
	MaxConcurrency   int        `db:"max_concurrency" json:"max_concurrency"`
	LastCrawlAt      *time.Time `db:"last_crawl_at" json:"last_crawl_at,omitempty"`
	LastCrawlStatus  *string    `db:"last_crawl_status" json:"last_crawl_status,omitempty"`
	CreatedAt        time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt        time.Time  `db:"updated_at" json:"updated_at"`
}
