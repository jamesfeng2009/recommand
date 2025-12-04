package crawler

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"time"

	"recommand/internal/domain"
	"recommand/internal/kafka"
	"recommand/internal/repository"
)

// Engine is a very simple fake crawler engine that simulates task progress.
type Engine struct {
	taskRepo   *repository.TaskRepo
	sourceRepo *repository.SourceRepo
	writer     *kafka.Writer
	logger     *log.Logger
}

func NewEngine(taskRepo *repository.TaskRepo, sourceRepo *repository.SourceRepo, writer *kafka.Writer, logger *log.Logger) *Engine {
	return &Engine{taskRepo: taskRepo, sourceRepo: sourceRepo, writer: writer, logger: logger}
}

// StartFakeTask runs a fake crawl in background, updating status/progress in DB.
// It uses a background context so it is not tied to any single HTTP request lifecycle.
func (e *Engine) StartFakeTask(taskID string) {
	go func() {
		ctx := context.Background()

		if e.logger != nil {
			e.logger.Printf("StartFakeTask begin, task_id=%s", taskID)
		}

		// load task and source for potential real fetch
		task, err := e.taskRepo.GetByID(ctx, taskID)
		if err != nil {
			if e.logger != nil {
				e.logger.Printf("StartFakeTask: failed to load task %s: %v", taskID, err)
			}
			return
		}
		if task == nil {
			if e.logger != nil {
				e.logger.Printf("StartFakeTask: task %s not found", taskID)
			}
			return
		}
		source, err := e.sourceRepo.GetByID(ctx, task.SourceID)
		if err != nil {
			if e.logger != nil {
				e.logger.Printf("StartFakeTask: failed to load source %d: %v", task.SourceID, err)
			}
		} else if source != nil && source.Enabled && source.Code == "people_military" {
			if e.logger != nil {
				e.logger.Printf("StartFakeTask: fetching people_military, url=%s", source.BaseURL)
			}
			// perform a minimal real HTTP GET once for people_military
			resp, err := http.Get(source.BaseURL)
			if err == nil {
				defer resp.Body.Close()
				// read a limited snippet to avoid huge payloads
				body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
				payload := map[string]any{
					"task_id":      task.TaskID,
					"source_id":    source.ID,
					"source_code":  source.Code,
					"url":          source.BaseURL,
					"status_code":  resp.StatusCode,
					"body_snippet": string(body),
				}
				if b, err := json.Marshal(payload); err == nil {
					if e.logger != nil {
						e.logger.Printf("StartFakeTask: writing to Kafka news.raw, bytes=%d", len(b))
					}
					if err := e.writer.WriteRaw(ctx, b); err != nil && e.logger != nil {
						e.logger.Printf("StartFakeTask: kafka write error: %v", err)
					}
				}
			} else if e.logger != nil {
				e.logger.Printf("StartFakeTask: http get failed for %s: %v", source.BaseURL, err)
			}
		} else if e.logger != nil {
			e.logger.Printf("StartFakeTask: source condition not matched, source=%+v", source)
		}

		// initial: set to running
		if err := e.taskRepo.UpdateStatusAndProgress(ctx, taskID, domain.StatusRunning, 0, 0); err != nil {
			if e.logger != nil {
				e.logger.Printf("StartFakeTask: failed to set task %s running: %v", taskID, err)
			}
			return
		}

		progress := 0.0
		pages := 0

		for i := 0; i < 5; i++ {
			// simulate work
			time.Sleep(1 * time.Second)
			progress += 20
			pages++

			if err := e.taskRepo.UpdateStatusAndProgress(ctx, taskID, domain.StatusRunning, progress, pages); err != nil {
				if e.logger != nil {
					e.logger.Printf("failed to update task %s progress: %v", taskID, err)
				}
				return
			}
		}

		// final: completed
		if err := e.taskRepo.UpdateStatusAndProgress(ctx, taskID, domain.StatusCompleted, 100, pages); err != nil {
			if e.logger != nil {
				e.logger.Printf("failed to complete task %s: %v", taskID, err)
			}
		}
	}()
}
