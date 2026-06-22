package routes

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"transcription-goserver/internal/controllers"
)

func SetupDashboardRoutes(r *gin.RouterGroup) {
	dc := &controllers.DashboardController{}

	r.GET("/status", func(c *gin.Context) {
		status := dc.GetServerStatus()
		c.JSON(http.StatusOK, status)
	})

	r.GET("/metrics", func(c *gin.Context) {
		metrics := dc.GetMetrics()
		c.JSON(http.StatusOK, metrics)
	})

	r.GET("/calls", func(c *gin.Context) {
		limitStr := c.DefaultQuery("limit", "50")
		limit, err := strconv.Atoi(limitStr)
		if err != nil || limit < 1 {
			limit = 50
		}
		if limit > 1000 {
			limit = 1000
		}

		calls, err := dc.GetCalls(limit)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"detail": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"calls": calls, "total": len(calls)})
	})

	r.GET("/logs", func(c *gin.Context) {
		limitStr := c.DefaultQuery("limit", "100")
		limit, err := strconv.Atoi(limitStr)
		if err != nil || limit < 1 {
			limit = 100
		}
		if limit > 1000 {
			limit = 1000
		}

		logs, err := dc.GetLogs(limit)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"detail": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"logs": logs})
	})

	r.GET("/processing", func(c *gin.Context) {
		stats := dc.GetProcessingStats()
		c.JSON(http.StatusOK, gin.H{
			"processing": stats["processing_ids"],
			"count":      stats["processing"],
		})
	})
}
