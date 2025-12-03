package config

import "github.com/kelseyhightower/envconfig"

type Config struct {
	HTTP     HTTPConfig
	Database DatabaseConfig
	Kafka    KafkaConfig
}

type HTTPConfig struct {
	ListenAddr string `envconfig:"HTTP_LISTEN_ADDR" default:":8080"`
}

type DatabaseConfig struct {
	DSN string `envconfig:"DB_DSN" default:"postgres://user:password@localhost:5432/crawler?sslmode=disable"`
}

type KafkaConfig struct {
	Brokers     []string `envconfig:"KAFKA_BROKERS" default:"localhost:9092"`
	TopicRaw    string   `envconfig:"KAFKA_TOPIC_RAW" default:"news.raw"`
	TopicParsed string   `envconfig:"KAFKA_TOPIC_PARSED" default:"news.parsed"`
}

func Load(cfg *Config) error {
	if err := envconfig.Process("", cfg); err != nil {
		return err
	}
	return nil
}
