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

	// Manual call editing endpoints
	r.GET("/call/:call_id", func(c *gin.Context) {
		callID := c.Param("call_id")
		result, err := dc.GetCallByID(callID)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Call not found"})
			return
		}
		c.JSON(http.StatusOK, result)
	})

	r.POST("/call/:call_id/update-manual", func(c *gin.Context) {
		callID := c.Param("call_id")
		var req struct {
			Transcription string `json:"transcription" binding:"required"`
			Summary       string `json:"summary" binding:"required"`
			PostToZoho    bool   `json:"post_to_zoho"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Update database
		result, err := dc.UpdateCallManually(callID, req.Transcription, req.Summary)
		if err != nil {
			c.JSON(http.StatusInternalServerError, result)
			return
		}

		// Optionally post to Zoho
		if req.PostToZoho {
			zohoResult, zohoErr := dc.PostToZohoManually(callID, req.Transcription, req.Summary)
			if zohoErr != nil {
				c.JSON(http.StatusOK, gin.H{
					"db_updated": true,
					"zoho_updated": false,
					"zoho_error": zohoResult["error"],
					"result": result,
				})
				return
			}
			c.JSON(http.StatusOK, gin.H{
				"db_updated": true,
				"zoho_updated": true,
				"result": result,
				"zoho": zohoResult,
			})
			return
		}

		c.JSON(http.StatusOK, result)
	})

	r.POST("/call/:call_id/post-to-zoho", func(c *gin.Context) {
		callID := c.Param("call_id")
		var req struct {
			Transcription string `json:"transcription" binding:"required"`
			Summary       string `json:"summary" binding:"required"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		result, err := dc.PostToZohoManually(callID, req.Transcription, req.Summary)
		if err != nil {
			c.JSON(http.StatusInternalServerError, result)
			return
		}
		c.JSON(http.StatusOK, result)
	})

	r.POST("/call/:call_id/fetch-from-zoho", func(c *gin.Context) {
		callID := c.Param("call_id")
		
		result, err := dc.FetchCallFromZoho(callID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, result)
			return
		}
		c.JSON(http.StatusOK, result)
	})
}
