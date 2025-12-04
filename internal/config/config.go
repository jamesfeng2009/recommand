package config

import "github.com/kelseyhightower/envconfig"

type Config struct {
	HTTP     HTTPConfig
	Database DatabaseConfig
	Kafka    KafkaConfig
	ES       ESConfig
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

type ESConfig struct {
	Address  string `envconfig:"ES_ADDRESS" default:"http://localhost:9200"`
	Username string `envconfig:"ES_USERNAME" default:"elastic"`
	Password string `envconfig:"ES_PASSWORD" default:""`
	Index    string `envconfig:"ES_INDEX" default:"news"`
}

func Load(cfg *Config) error {
	if err := envconfig.Process("", cfg); err != nil {
		return err
	}
	return nil
}
