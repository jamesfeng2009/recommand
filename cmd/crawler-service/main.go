package main

import (
	"log"
	"net/http"

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

	r := gin.Default()

	sourceRepo := repository.NewSourceRepo(pgDB)
	taskRepo := repository.NewTaskRepo(pgDB)
	sourceHandler := handlers.NewSourceHandler(sourceRepo)
	engine := crawler.NewEngine(taskRepo, sourceRepo, kafkaWriter, logger)
	taskHandler := handlers.NewTaskHandler(sourceRepo, taskRepo, engine)

	chttp.RegisterRoutes(r, sourceHandler, taskHandler)

	addr := cfg.HTTP.ListenAddr
	logger.Printf("crawler-service listening on %s", addr)
	if err := http.ListenAndServe(addr, r); err != nil {
		logger.Fatalf("http server stopped: %v", err)
	}
}
