package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"log"
	"time"

	_ "github.com/lib/pq"
	"github.com/segmentio/kafka-go"

	"recommand/internal/config"
	"recommand/internal/db"
)

// ParsedNews mirrors the message structure in news.parsed.
type ParsedNews struct {
	ID          string    `json:"id"`
	TaskID      string    `json:"task_id"`
	SourceID    int64     `json:"source_id"`
	SourceCode  string    `json:"source_code"`
	URL         string    `json:"url"`
	Title       string    `json:"title"`
	Content     string    `json:"content"`
	PublishTime time.Time `json:"publish_time"`
	CrawlTime   time.Time `json:"crawl_time"`
	Hash        string    `json:"hash"`
}

func main() {
	var cfg config.Config
	if err := config.Load(&cfg); err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	// DB connection
	sqldb, err := db.NewPostgres(cfg.Database)
	if err != nil {
		log.Fatalf("failed to connect postgres: %v", err)
	}
	defer sqldb.Close()

	// Kafka reader on news.parsed（简单分区读取，从最早 offset 开始，方便本地调试）
	brokers := cfg.Kafka.Brokers
	if len(brokers) == 0 {
		brokers = []string{"localhost:9092"}
	}

	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:     brokers,
		Topic:       cfg.Kafka.TopicParsed,
		Partition:   0,
		StartOffset: kafka.FirstOffset,
		MinBytes:    1,
		MaxBytes:    10e6,
	})
	defer reader.Close()

	log.Printf("news-sink consuming from %s and writing to Postgres", cfg.Kafka.TopicParsed)

	ctx := context.Background()
	for {
		m, err := reader.ReadMessage(ctx)
		if err != nil {
			log.Printf("read error from parsed: %v", err)
			time.Sleep(time.Second)
			continue
		}

		var n ParsedNews
		if err := json.Unmarshal(m.Value, &n); err != nil {
			log.Printf("decode parsed error at offset=%d: %v", m.Offset, err)
			continue
		}

		if err := upsertNews(ctx, sqldb, &n); err != nil {
			log.Printf("upsert news error id=%s url=%s: %v", n.ID, n.URL, err)
			continue
		}

		log.Printf("news-sink: upserted news id=%s url=%s", n.ID, n.URL)
	}
}

func upsertNews(ctx context.Context, db *sql.DB, n *ParsedNews) error {
	// 使用 hash 作为幂等键进行 UPSERT，按 hash 去重。
	const q = `
INSERT INTO news (
	id, hash, task_id, source_id, source_code, url, title, content, publish_time, crawl_time, created_at, updated_at
) VALUES (
	$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, now(), now()
) ON CONFLICT (hash) DO UPDATE SET
	title = EXCLUDED.title,
	content = EXCLUDED.content,
	publish_time = EXCLUDED.publish_time,
	crawl_time = EXCLUDED.crawl_time,
	updated_at = now();
`
	_, err := db.ExecContext(ctx, q,
		n.ID,
		n.Hash,
		n.TaskID,
		n.SourceID,
		n.SourceCode,
		n.URL,
		n.Title,
		n.Content,
		n.PublishTime,
		n.CrawlTime,
	)
	return err
}
