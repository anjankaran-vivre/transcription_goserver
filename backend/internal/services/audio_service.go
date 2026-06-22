package services

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"transcription-goserver/internal/config"
	"transcription-goserver/internal/logging"
)

type AudioService struct{}

func (as *AudioService) DownloadAudio(recURL, callID string, workerID int) (string, bool, string) {
	logStreamer := logging.GetLogStreamer()
	cfg := config.Settings
	var audioFile string
	errorMsg := "Max retries exceeded"

	maxAttempts := cfg.MaxDownloadRetries + 1

	for attempt := 1; attempt <= maxAttempts; attempt++ {
		logStreamer.Info("AudioService",
			fmt.Sprintf("Worker %d: Downloading call %s (Attempt %d)", workerID, callID, attempt))

		req, err := http.NewRequest("GET", recURL, nil)
		if err != nil {
			errorMsg = err.Error()
			logStreamer.Error("AudioService", fmt.Sprintf("Worker %d: Attempt %d failed: %s", workerID, attempt, errorMsg))
			continue
		}

		req.SetBasicAuth(cfg.AudioUsername, cfg.AudioPassword)

		client := &http.Client{Timeout: 120 * time.Second}
		resp, err := client.Do(req)
		if err != nil {
			errorMsg = err.Error()
			logStreamer.Error("AudioService", fmt.Sprintf("Worker %d: Attempt %d failed: %s", workerID, attempt, errorMsg))
			if attempt < maxAttempts {
				time.Sleep(time.Duration(cfg.RetryInterval) * time.Second)
			}
			continue
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			errorMsg = fmt.Sprintf("HTTP %d", resp.StatusCode)
			logStreamer.Error("AudioService", fmt.Sprintf("Worker %d: Attempt %d failed: %s", workerID, attempt, errorMsg))
			if attempt < maxAttempts {
				time.Sleep(time.Duration(cfg.RetryInterval) * time.Second)
			}
			continue
		}

		tmpFile, err := os.CreateTemp("", "*.m4a")
		if err != nil {
			errorMsg = err.Error()
			return "", false, errorMsg
		}

		written, err := io.Copy(tmpFile, resp.Body)
		tmpFile.Close()
		if err != nil {
			os.Remove(tmpFile.Name())
			errorMsg = err.Error()
			logStreamer.Error("AudioService", fmt.Sprintf("Worker %d: Attempt %d failed: %s", workerID, attempt, errorMsg))
			if attempt < maxAttempts {
				time.Sleep(time.Duration(cfg.RetryInterval) * time.Second)
			}
			continue
		}

		audioFile = tmpFile.Name()
		logStreamer.Info("AudioService",
			fmt.Sprintf("Worker %d: Download successful (%d bytes)", workerID, written))
		return audioFile, true, ""
	}

	return "", false, errorMsg
}

func (as *AudioService) CleanupAudio(audioFile string) bool {
	if audioFile == "" {
		return true
	}
	if _, err := os.Stat(audioFile); os.IsNotExist(err) {
		return true
	}
	if err := os.Remove(audioFile); err != nil {
		logging.GetLogStreamer().Warning("AudioService", fmt.Sprintf("Cleanup failed: %v", err))
		return false
	}
	return true
}
