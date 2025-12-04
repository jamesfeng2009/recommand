package http

import (
	"github.com/gin-gonic/gin"

	"recommand/internal/http/handlers"
)

func RegisterRoutes(r *gin.Engine, sh *handlers.SourceHandler, th *handlers.TaskHandler) {
	api := r.Group("/api/v1")
	{
		crawler := api.Group("/crawler")
		{
			crawler.GET("/health", func(c *gin.Context) {
				c.JSON(200, gin.H{"status": "ok"})
			})

			// sources
			crawler.GET("/sources", sh.ListSources)
			crawler.POST("/sources", sh.CreateSource)
			crawler.PUT("/sources/:id", sh.UpdateSource)
			crawler.PUT("/sources/:id/status", sh.UpdateSourceStatus)

			// tasks
			crawler.POST("/tasks", th.CreateTask)
			crawler.GET("/tasks", th.ListTasks)
			crawler.GET("/tasks/:task_id", th.GetTask)
			crawler.POST("/tasks/:task_id/stop", th.StopTask)
		}
	}
}
