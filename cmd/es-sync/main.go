package main

import (
	"bytes"
	"context"
	"crypto/tls"
	"database/sql"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/elastic/go-elasticsearch/v8"

	"recommand/internal/config"
	"recommand/internal/db"
)

type NewsRow struct {
	ID          string
	Hash        sql.NullString
	SourceCode  string
	URL         string
	Title       string
	Content     string
	PublishTime sql.NullTime
	CrawlTime   time.Time
	UpdatedAt   time.Time
}

func main() {
	var cfg config.Config
	if err := config.Load(&cfg); err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	// DB
	sqldb, err := db.NewPostgres(cfg.Database)
	if err != nil {
		log.Fatalf("failed to connect postgres: %v", err)
	}
	defer sqldb.Close()

	// ES client (dev only: skip TLS verification for local self-signed cert)
	es, err := elasticsearch.NewClient(elasticsearch.Config{
		Addresses: []string{cfg.ES.Address},
		Username:  cfg.ES.Username,
		Password:  cfg.ES.Password,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	})
	if err != nil {
		log.Fatalf("failed to create ES client: %v", err)
	}

	log.Printf("es-sync started: db=%s es=%s index=%s", cfg.Database.DSN, cfg.ES.Address, cfg.ES.Index)

	ctx := context.Background()
	lastSync := time.Time{}

	for {
		rows, err := fetchNewsSince(ctx, sqldb, lastSync)
		if err != nil {
			log.Printf("fetchNewsSince error: %v", err)
			time.Sleep(5 * time.Second)
			continue
		}
		if len(rows) == 0 {
			time.Sleep(10 * time.Second)
			continue
		}

		if err := bulkIndexNews(ctx, es, cfg.ES.Index, rows); err != nil {
			log.Printf("bulkIndexNews error: %v", err)
		}

		for _, r := range rows {
			if r.UpdatedAt.After(lastSync) {
				lastSync = r.UpdatedAt
			}
		}
	}
}

func fetchNewsSince(ctx context.Context, db *sql.DB, since time.Time) ([]NewsRow, error) {
	const q = `
SELECT id, hash, source_code, url, title, content, publish_time, crawl_time, updated_at
FROM news
WHERE updated_at > $1
ORDER BY updated_at ASC
LIMIT 500
`
	rows, err := db.QueryContext(ctx, q, since)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []NewsRow
	for rows.Next() {
		var r NewsRow
		if err := rows.Scan(&r.ID, &r.Hash, &r.SourceCode, &r.URL, &r.Title, &r.Content, &r.PublishTime, &r.CrawlTime, &r.UpdatedAt); err != nil {
			return nil, err
		}
		result = append(result, r)
	}
	return result, rows.Err()
}

func bulkIndexNews(ctx context.Context, es *elasticsearch.Client, index string, rows []NewsRow) error {
	if len(rows) == 0 {
		return nil
	}
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)

	for _, r := range rows {
		id := r.ID
		if r.Hash.Valid && r.Hash.String != "" {
			id = r.Hash.String
		}

		meta := map[string]map[string]string{
			"index": {
				"_index": index,
				"_id":    id,
			},
		}
		if err := enc.Encode(meta); err != nil {
			return err
		}

		body := map[string]any{
			"id":          r.ID,
			"hash":        r.Hash.String,
			"source_code": r.SourceCode,
			"url":         r.URL,
			"title":       r.Title,
			"content":     r.Content,
			"crawl_time":  r.CrawlTime,
			"updated_at":  r.UpdatedAt,
		}
		if r.PublishTime.Valid {
			body["publish_time"] = r.PublishTime.Time
		}
		if err := enc.Encode(body); err != nil {
			return err
		}
	}

	res, err := es.Bulk(bytes.NewReader(buf.Bytes()), es.Bulk.WithContext(ctx))
	if err != nil {
		// 部分环境下可能在读取响应时返回 EOF，这里记录 warning 但不中断主循环。
		if err == io.EOF {
			log.Printf("es bulk warning: EOF while reading response")
			return nil
		}
		return err
	}
	defer res.Body.Close()

	if res.IsError() {
		body, _ := io.ReadAll(res.Body)
		log.Printf("es bulk error: status=%s body=%s", res.Status(), string(body))
	}

	return nil
}
