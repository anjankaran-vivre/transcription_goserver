package controllers

import (
	"fmt"
	"os"
	"syscall"
	"time"

	"transcription-goserver/internal/logging"
	"transcription-goserver/internal/services"
)

type AdminController struct{}

func (ac *AdminController) RestartServer() AdminResponse {
	logStreamer := logging.GetLogStreamer()
	logStreamer.Info("Admin", "Server restart requested")

	go func() {
		time.Sleep(2 * time.Second)
		proc, _ := os.FindProcess(os.Getpid())
		proc.Signal(syscall.SIGTERM)
	}()

	return AdminResponse{
		Status:  "restarting",
		Message: "Server will restart in 2 seconds",
	}
}

func (ac *AdminController) StopServer() AdminResponse {
	logStreamer := logging.GetLogStreamer()
	logStreamer.Info("Admin", "Server stop requested")

	go func() {
		time.Sleep(1 * time.Second)
		proc, _ := os.FindProcess(os.Getpid())
		proc.Signal(syscall.SIGTERM)
	}()

	return AdminResponse{
		Status:  "stopping",
		Message: "Server will stop in 1 second",
	}
}

func (ac *AdminController) ClearQueue() AdminResponse {
	logStreamer := logging.GetLogStreamer()
	cleared := 0

	for {
		select {
		case <-TaskQueue:
			cleared++
		default:
			logStreamer.Info("Admin", fmt.Sprintf("Queue cleared: %d items removed", cleared))
			return AdminResponse{
				Status:  "success",
				Message: fmt.Sprintf("Cleared %d items from queue", cleared),
				Cleared: &cleared,
			}
		}
	}
}

func (ac *AdminController) TestEmail() AdminResponse {
	logStreamer := logging.GetLogStreamer()
	var es services.EmailService
	es.SendAlert("Test Alert",
		"This is a test email from the Transcription Server dashboard.", false)

	logStreamer.Info("Admin", "Test email sent")
	return AdminResponse{
		Status:  "success",
		Message: "Test email sent",
	}
}
