package main

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/segmentio/kafka-go"

	"recommand/internal/config"
	"recommand/internal/content"
)

func main() {
	var cfg config.Config
	if err := config.Load(&cfg); err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	brokers := cfg.Kafka.Brokers
	if len(brokers) == 0 {
		brokers = []string{"localhost:9092"}
	}

	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:   brokers,
		Topic:     cfg.Kafka.TopicRaw,
		Partition: 0,
		// 从最早的 offset 开始读取，便于本地调试时看到所有历史消息
		StartOffset: kafka.FirstOffset,
		MinBytes:    1,
		MaxBytes:    10e6,
	})
	defer reader.Close()

	log.Printf("raw-consumer listening on topic %s", cfg.Kafka.TopicRaw)

	ctx := context.Background()
	for {
		m, err := reader.ReadMessage(ctx)
		if err != nil {
			log.Printf("read error: %v", err)
			time.Sleep(time.Second)
			continue
		}

		// 反序列化 Engine 写入的原始 JSON
		var raw struct {
			TaskID      string `json:"task_id"`
			SourceID    int64  `json:"source_id"`
			SourceCode  string `json:"source_code"`
			URL         string `json:"url"`
			StatusCode  int    `json:"status_code"`
			BodySnippet string `json:"body_snippet"`
		}
		if err := json.Unmarshal(m.Value, &raw); err != nil {
			log.Printf("decode error at offset=%d: %v, raw=%s", m.Offset, err, string(m.Value))
			continue
		}

		// 目前只处理 people_military，其它来源先打印原始 JSON
		if raw.SourceCode == "people_military" {
			article, err := content.ParsePeopleMilitary(raw.BodySnippet)
			if err != nil {
				log.Printf("parse people_military failed at offset=%d: %v", m.Offset, err)
				continue
			}
			log.Printf("parsed article: task_id=%s title=%q publish_time=%s url=%s", raw.TaskID, article.Title, article.PublishTime.Format(time.RFC3339), raw.URL)
		} else {
			log.Printf("%s: partition=%d offset=%d source=%s url=%s status=%d", cfg.Kafka.TopicRaw, m.Partition, m.Offset, raw.SourceCode, raw.URL, raw.StatusCode)
		}
	}
}
