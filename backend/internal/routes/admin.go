package routes

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"transcription-goserver/internal/controllers"
)

func SetupAdminRoutes(r *gin.RouterGroup) {
	ac := &controllers.AdminController{}

	r.POST("/restart", func(c *gin.Context) {
		result := ac.RestartServer()
		c.JSON(http.StatusOK, result)
	})

	r.POST("/stop", func(c *gin.Context) {
		result := ac.StopServer()
		c.JSON(http.StatusOK, result)
	})

	r.POST("/clear-queue", func(c *gin.Context) {
		result := ac.ClearQueue()
		c.JSON(http.StatusOK, result)
	})

	r.POST("/test-email", func(c *gin.Context) {
		result := ac.TestEmail()
		c.JSON(http.StatusOK, result)
	})
}
