package main

import (
	"crypto/tls"
	"log"
	"net/http"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/gin-gonic/gin"

	"recommand/internal/config"
	"recommand/internal/crawler"
	"recommand/internal/db"
	chttp "recommand/internal/http"
	"recommand/internal/http/handlers"
	"recommand/internal/kafka"
	"recommand/internal/repository"
)

func main() {
	var cfg config.Config
	if err := config.Load(&cfg); err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	logger := log.Default()

	pgDB, err := db.NewPostgres(cfg.Database)
	if err != nil {
		logger.Fatalf("failed to connect postgres: %v", err)
	}
	defer pgDB.Close()

	kafkaWriter, err := kafka.NewWriter(cfg.Kafka)
	if err != nil {
		logger.Fatalf("failed to create kafka writer: %v", err)
	}
	defer kafkaWriter.Close()

	// ES client for search (dev only: skip TLS verification for local self-signed cert)
	esClient, err := elasticsearch.NewClient(elasticsearch.Config{
		Addresses: []string{cfg.ES.Address},
		Username:  cfg.ES.Username,
		Password:  cfg.ES.Password,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	})
	if err != nil {
		logger.Fatalf("failed to create ES client: %v", err)
	}

	r := gin.Default()

	sourceRepo := repository.NewSourceRepo(pgDB)
	taskRepo := repository.NewTaskRepo(pgDB)
	sourceHandler := handlers.NewSourceHandler(sourceRepo)
	engine := crawler.NewEngine(taskRepo, sourceRepo, kafkaWriter, logger)
	taskHandler := handlers.NewTaskHandler(sourceRepo, taskRepo, engine)
	searchHandler := handlers.NewSearchHandler(esClient, cfg.ES.Index)

	chttp.RegisterRoutes(r, sourceHandler, taskHandler, searchHandler)

	addr := cfg.HTTP.ListenAddr
	logger.Printf("crawler-service listening on %s", addr)
	if err := http.ListenAndServe(addr, r); err != nil {
		logger.Fatalf("http server stopped: %v", err)
	}
}
