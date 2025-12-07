package main

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"log"
	"time"

	"recommand/internal/config"
	"recommand/internal/db"
)

type row struct {
	ID          string
	URL         string
	Title       string
	PublishTime sql.NullTime
}

func main() {
	var cfg config.Config
	if err := config.Load(&cfg); err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	sqldb, err := db.NewPostgres(cfg.Database)
	if err != nil {
		log.Fatalf("failed to connect postgres: %v", err)
	}
	defer sqldb.Close()

	ctx := context.Background()

	for {
		rows, err := fetchBatch(ctx, sqldb)
		if err != nil {
			log.Fatalf("fetch batch error: %v", err)
		}
		if len(rows) == 0 {
			log.Println("backfill-hash done: no more rows")
			return
		}

		for _, r := range rows {
			var pt time.Time
			if r.PublishTime.Valid {
				pt = r.PublishTime.Time
			}
			h := sha256.New()
			h.Write([]byte(r.URL))
			h.Write([]byte(r.Title))
			h.Write([]byte(pt.Format(time.RFC3339)))
			hash := hex.EncodeToString(h.Sum(nil))

			if err := updateHash(ctx, sqldb, r.ID, hash); err != nil {
				log.Fatalf("update hash error for id=%s: %v", r.ID, err)
			}
		}

		log.Printf("backfill-hash: processed %d rows", len(rows))
	}
}

func fetchBatch(ctx context.Context, db *sql.DB) ([]row, error) {
	const q = `
SELECT id, url, title, publish_time
FROM news
WHERE hash IS NULL
ORDER BY created_at ASC
LIMIT 500
`
	rows, err := db.QueryContext(ctx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []row
	for rows.Next() {
		var r row
		if err := rows.Scan(&r.ID, &r.URL, &r.Title, &r.PublishTime); err != nil {
			return nil, err
		}
		result = append(result, r)
	}
	return result, rows.Err()
}

func updateHash(ctx context.Context, db *sql.DB, id, hash string) error {
	const q = `UPDATE news SET hash = $1, updated_at = now() WHERE id = $2`
	_, err := db.ExecContext(ctx, q, hash, id)
	return err
}
