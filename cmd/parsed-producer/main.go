package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"log"
	"time"

	"github.com/segmentio/kafka-go"

	"recommand/internal/config"
	"recommand/internal/content"
	ikafka "recommand/internal/kafka"
)

// RawMessage is the payload written by crawler Engine into news.raw.
type RawMessage struct {
	TaskID      string `json:"task_id"`
	SourceID    int64  `json:"source_id"`
	SourceCode  string `json:"source_code"`
	URL         string `json:"url"`
	StatusCode  int    `json:"status_code"`
	BodySnippet string `json:"body_snippet"`
}

// ParsedNews is the structured news we will write into news.parsed.
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

	brokers := cfg.Kafka.Brokers
	if len(brokers) == 0 {
		brokers = []string{"localhost:9092"}
	}

	// reader: consume from news.raw（简单分区读取，从最早 offset 开始，方便本地调试）
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:     brokers,
		Topic:       cfg.Kafka.TopicRaw,
		Partition:   0,
		StartOffset: kafka.FirstOffset,
		MinBytes:    1,
		MaxBytes:    10e6,
	})
	defer reader.Close()

	// writer: produce to news.parsed
	writer, err := ikafka.NewWriter(cfg.Kafka)
	if err != nil {
		log.Fatalf("failed to create kafka writer: %v", err)
	}
	defer writer.Close()

	log.Printf("parsed-producer consuming from %s and producing to %s", cfg.Kafka.TopicRaw, cfg.Kafka.TopicParsed)

	ctx := context.Background()
	for {
		m, err := reader.ReadMessage(ctx)
		if err != nil {
			log.Printf("read error from raw: %v", err)
			time.Sleep(time.Second)
			continue
		}

		var raw RawMessage
		if err := json.Unmarshal(m.Value, &raw); err != nil {
			log.Printf("decode raw error at offset=%d: %v", m.Offset, err)
			continue
		}

		article, err := content.Parse(raw.SourceCode, raw.BodySnippet)
		if err != nil {
			log.Printf("parse source=%s failed at offset=%d: %v", raw.SourceCode, m.Offset, err)
			continue
		}

		h := sha256.New()
		h.Write([]byte(raw.URL))
		h.Write([]byte(article.Title))
		h.Write([]byte(article.PublishTime.Format(time.RFC3339)))
		hash := hex.EncodeToString(h.Sum(nil))

		parsed := ParsedNews{
			ID:          raw.TaskID + "::" + raw.URL, // 简单 ID，后续可换成 uuid/hash
			TaskID:      raw.TaskID,
			SourceID:    raw.SourceID,
			SourceCode:  raw.SourceCode,
			URL:         raw.URL,
			Title:       article.Title,
			Content:     article.Content,
			PublishTime: article.PublishTime,
			CrawlTime:   time.Now().UTC(),
			Hash:        hash,
		}

		b, err := json.Marshal(parsed)
		if err != nil {
			log.Printf("marshal parsed error at offset=%d: %v", m.Offset, err)
			continue
		}

		if err := writer.WriteParsed(ctx, b); err != nil {
			log.Printf("write parsed error for task=%s: %v", raw.TaskID, err)
			continue
		}

		log.Printf("parsed-producer: produced parsed news for task=%s url=%s", raw.TaskID, raw.URL)
	}
}
