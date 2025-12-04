package handlers

import (
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"recommand/internal/crawler"
	"recommand/internal/domain"
	"recommand/internal/repository"
)

type TaskHandler struct {
	sourceRepo *repository.SourceRepo
	taskRepo   *repository.TaskRepo
	engine     *crawler.Engine
}

func NewTaskHandler(sr *repository.SourceRepo, tr *repository.TaskRepo, eng *crawler.Engine) *TaskHandler {
	return &TaskHandler{sourceRepo: sr, taskRepo: tr, engine: eng}
}

type CreateTaskRequest struct {
	SourceID int64  `json:"source_id" binding:"required"`
	Mode     string `json:"mode" binding:"required,oneof=full incremental"`
	Since    string `json:"since"`     // RFC3339，可选
	MaxPages *int   `json:"max_pages"` // 可选
}

// CreateTask POST /api/v1/crawler/tasks
func (h *TaskHandler) CreateTask(c *gin.Context) {
	var req CreateTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	source, err := h.sourceRepo.GetByID(c.Request.Context(), req.SourceID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error"})
		return
	}
	if source == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "source_not_found"})
		return
	}

	var sincePtr *time.Time
	if req.Since != "" {
		parsed, err := time.Parse(time.RFC3339, req.Since)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid since format, expect RFC3339"})
			return
		}
		sincePtr = &parsed
	}

	mode := domain.CrawlMode(req.Mode)
	if mode != domain.CrawlModeFull && mode != domain.CrawlModeIncremental {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid mode"})
		return
	}

	task := &domain.CrawlTask{
		TaskID:     uuid.NewString(),
		SourceID:   source.ID,
		SourceName: source.Name,
		Mode:       mode,
		Since:      sincePtr,
		MaxPages:   req.MaxPages,
		Status:     domain.StatusPending,
		Progress:   0,
		CreatedAt:  time.Now().UTC(),
		UpdatedAt:  time.Now().UTC(),
	}

	if err := h.taskRepo.Create(c.Request.Context(), task); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error"})
		return
	}

	// 启动一个假的爬虫执行流程（后台 goroutine 模拟进度）
	h.engine.StartFakeTask(task.TaskID)

	c.JSON(http.StatusAccepted, task)
}

// GetTask GET /api/v1/crawler/tasks/:task_id
func (h *TaskHandler) GetTask(c *gin.Context) {
	id := c.Param("task_id")
	task, err := h.taskRepo.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error"})
		return
	}
	if task == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not_found"})
		return
	}
	c.JSON(http.StatusOK, task)
}

// ListTasks GET /api/v1/crawler/tasks
func (h *TaskHandler) ListTasks(c *gin.Context) {
	var (
		statusPtr   *domain.CrawlStatus
		sourceIDPtr *int64
	)

	if s := c.Query("status"); s != "" {
		st := domain.CrawlStatus(s)
		statusPtr = &st
	}
	if sid := c.Query("source_id"); sid != "" {
		if parsed, err := strconv.ParseInt(sid, 10, 64); err == nil {
			sourceIDPtr = &parsed
		}
	}

	tasks, err := h.taskRepo.List(c.Request.Context(), sourceIDPtr, statusPtr)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": tasks, "total": len(tasks)})
}

type StopTaskRequest struct {
	Reason string `json:"reason"`
}

// StopTask POST /api/v1/crawler/tasks/:task_id/stop
func (h *TaskHandler) StopTask(c *gin.Context) {
	id := c.Param("task_id")
	var req StopTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil && err != io.EOF {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 简单地把状态标记为 stopped，后续实际运行中需要通知执行引擎
	if err := h.taskRepo.UpdateStatus(c.Request.Context(), id, domain.StatusStopped); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error"})
		return
	}
	c.Status(http.StatusNoContent)
}
