package kafka

import (
	"context"
	"recommand/internal/config"

	"github.com/segmentio/kafka-go"
)

type Writer struct {
	Raw    *kafka.Writer
	Parsed *kafka.Writer
}

func NewWriter(cfg config.KafkaConfig) (*Writer, error) {
	wRaw := &kafka.Writer{
		Addr:     kafka.TCP(cfg.Brokers...),
		Topic:    cfg.TopicRaw,
		Balancer: &kafka.LeastBytes{},
	}
	wParsed := &kafka.Writer{
		Addr:     kafka.TCP(cfg.Brokers...),
		Topic:    cfg.TopicParsed,
		Balancer: &kafka.LeastBytes{},
	}
	return &Writer{Raw: wRaw, Parsed: wParsed}, nil
}

func (w *Writer) Close() error {
	if w == nil {
		return nil
	}
	if err := w.Raw.Close(); err != nil {
		return err
	}
	if err := w.Parsed.Close(); err != nil {
		return err
	}
	return nil
}

// WriteRaw writes a single message value to the raw topic.
func (w *Writer) WriteRaw(ctx context.Context, value []byte) error {
	if w == nil || w.Raw == nil {
		return nil
	}
	return w.Raw.WriteMessages(ctx, kafka.Message{Value: value})
}

// WriteParsed writes a single message value to the parsed topic.
func (w *Writer) WriteParsed(ctx context.Context, value []byte) error {
	if w == nil || w.Parsed == nil {
		return nil
	}
	return w.Parsed.WriteMessages(ctx, kafka.Message{Value: value})
}
