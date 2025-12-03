package kafka

import (
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
