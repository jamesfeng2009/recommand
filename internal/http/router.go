package http

import "github.com/gin-gonic/gin"

func RegisterRoutes(r *gin.Engine) {
	api := r.Group("/api/v1")
	{
		crawler := api.Group("/crawler")
		{
			crawler.GET("/health", func(c *gin.Context) {
				c.JSON(200, gin.H{"status": "ok"})
			})
			// TODO: sources and tasks routes will be added here
		}
	}
}
