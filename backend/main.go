package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	"transcription-goserver/internal/config"
	"transcription-goserver/internal/controllers"
	"transcription-goserver/internal/database"
	"transcription-goserver/internal/logging"
	"transcription-goserver/internal/models"
	"transcription-goserver/internal/routes"
	"transcription-goserver/internal/socketio"
	"transcription-goserver/internal/workers"
)

func main() {
	// Load configuration
	cfg := config.Load()

	// Initialize database
	database.InitDB()
	database.AutoMigrate(&models.CallLog{}, &models.SystemLog{})

	// Initialize task queue
	controllers.InitTaskQueue(1000)

	// Initialize Socket.IO
	sio := socketio.InitSocketIO()
	defer sio.Close()

	// Connect LogStreamer to Socket.IO broadcast
	logStreamer := logging.GetLogStreamer()
	logStreamer.SetBroadcastFunc(socketio.BroadcastLog)

	// Start workers
	workers.StartWorkers()

	// Setup Gin router
	gin.SetMode(gin.ReleaseMode)
	if cfg.Debug {
		gin.SetMode(gin.DebugMode)
	}

	r := gin.Default()

	// CORS
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"*"},
		AllowCredentials: true,
	}))

	// REST API routes
	api := r.Group("/api")
	routes.SetupDashboardRoutes(api)

	admin := r.Group("/admin")
	routes.SetupAdminRoutes(admin)

	routes.SetupTranscriptionRoutes(r.Group("/"))

	// Run Socket.IO on separate port 5052 so it never blocks main API
	go func() {
		sioMux := http.NewServeMux()
		sioMux.Handle("/socket.io/", sio)
		sioServer := &http.Server{
			Addr:    fmt.Sprintf("%s:5052", cfg.Host),
			Handler: sioMux,
		}
		fmt.Printf("  Socket.IO: ws://%s:5052/socket.io\n", cfg.Host)
		if err := sioServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("Socket.IO server error: %v", err)
		}
	}()

	// Print banner
	fmt.Println("=" + strings.Repeat("=", 59))
	fmt.Println("  TRANSCRIPTION SERVER (Go)")
	fmt.Println("=" + strings.Repeat("=", 59))
	fmt.Printf("  PID:      %d\n", os.Getpid())
	fmt.Printf("  Host:     %s\n", cfg.Host)
	fmt.Printf("  Port:     %d\n", cfg.Port)
	fmt.Printf("  Workers:  %d\n", cfg.NumWorkers)
	fmt.Printf("  Logs:     %s\n", cfg.LogsDir)
	fmt.Println("=" + strings.Repeat("=", 59))
	fmt.Printf("  API:       http://%s:%d\n", cfg.Host, cfg.Port)
	fmt.Printf("  Dashboard: http://localhost:3000\n")
	fmt.Printf("  Socket.IO: ws://%s:5052/socket.io\n", cfg.Host)
	fmt.Println("=" + strings.Repeat("=", 59))
	fmt.Println("  ✓ Gin with goroutine workers")
	fmt.Println("  ✓ Socket.IO on separate port 5052")
	fmt.Println("  ✓ MSSQL database logging")
	fmt.Println("  Press Ctrl+C to stop")
	fmt.Println("=" + strings.Repeat("=", 59))

	logStreamer.Info("Main", "Server started with Gin + Socket.IO")

	// Create HTTP server
	srv := &http.Server{
		Addr:    fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		Handler: r,
	}

	// Graceful shutdown
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed: %v", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	fmt.Println("\n" + "=" + strings.Repeat("=", 59))
	fmt.Println("Shutting down gracefully...")
	fmt.Println("=" + strings.Repeat("=", 59))

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	fmt.Println("All workers stopped")
	logStreamer.Info("Main", "Server stopped")
}