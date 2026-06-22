package routes

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"transcription-goserver/internal/controllers"
)

func SetupTranscriptionRoutes(r *gin.RouterGroup) {
	tc := &controllers.TranscriptionController{}

	r.GET("/server_transcription", func(c *gin.Context) {
		code := c.Query("code")
		result, err := tc.HandleOAuthCallback(code)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"detail": err.Error()})
			return
		}
		c.JSON(http.StatusOK, result)
	})

	r.POST("/server_transcription", func(c *gin.Context) {
		var req controllers.TranscriptionRequest
		if err := c.ShouldBind(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"detail": err.Error()})
			return
		}

		resp, err := tc.SubmitTranscription(req)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"detail": err.Error()})
			return
		}
		c.JSON(http.StatusOK, resp)
	})

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "healthy"})
	})
}
