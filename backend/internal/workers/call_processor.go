package workers

import (
	"fmt"
	"strings"
	"time"

	"transcription-goserver/internal/config"
	"transcription-goserver/internal/controllers"
	"transcription-goserver/internal/logging"
	"transcription-goserver/internal/models"
	"transcription-goserver/internal/services"
)

func StartWorkers() {
	cfg := config.Settings
	logStreamer := logging.GetLogStreamer()

	fmt.Printf("\n%s\n", strings.Repeat("=", 60))
	fmt.Printf("STARTING %d WORKER GOROUTINES\n", cfg.NumWorkers)
	fmt.Printf("%s\n\n", strings.Repeat("=", 60))

	for i := 0; i < cfg.NumWorkers; i++ {
		go worker(i + 1)
		fmt.Printf("[MAIN] ✓ Started worker goroutine %d\n", i+1)
		logStreamer.Info("Main", fmt.Sprintf("Started worker goroutine %d", i+1))
	}

	fmt.Printf("\n%s\n", strings.Repeat("=", 60))
	fmt.Printf("ALL %d WORKERS STARTED - READY FOR JOBS\n", cfg.NumWorkers)
	fmt.Printf("%s\n\n", strings.Repeat("=", 60))
}

func worker(id int) {
	logStreamer := logging.GetLogStreamer()
	logStreamer.Info("Worker", fmt.Sprintf("Worker %d started", id))

	checkCount := 0
	for job := range controllers.TaskQueue {
		checkCount++

		if checkCount%12 == 0 {
			logStreamer.Info("Worker", fmt.Sprintf("Worker %d: Still alive, queue size: %d", id, len(controllers.TaskQueue)))
		}

		logStreamer.Info("Worker", fmt.Sprintf("Worker %d: GOT JOB: %s", id, job.CallID))
		logStreamer.Info("Worker", fmt.Sprintf("Worker %d: PROCESSING %s", id, job.CallID))

		processCall(job, id)

		logStreamer.Info("Worker", fmt.Sprintf("Worker %d: COMPLETED %s", id, job.CallID))
	}
}

func processCall(job controllers.Job, workerID int) {
	callID := job.CallID
	recURL := job.RecURL
	startTime := time.Now()

	callTracker := models.GetCallTracker()
	logStreamer := logging.GetLogStreamer()

	if !callTracker.TryAcquire(callID) {
		logStreamer.Info("Worker", fmt.Sprintf("Worker %d: Call %s already processing/skipped", workerID, callID))
		return
	}

	logStreamer.Info("Worker", fmt.Sprintf("Worker %d: STARTED -> Call %s", workerID, callID))
	time.Sleep(time.Duration(config.Settings.ProcessDelay) * time.Second)

	var audioFile string
	var totalAPICalls int

	defer func() {
		if audioFile != "" {
			var as services.AudioService
			as.CleanupAudio(audioFile)
		}
	}()

	// 1. Download Audio
	var as services.AudioService
	var success bool
	var errorMsg string
	audioFile, success, errorMsg = as.DownloadAudio(recURL, callID, workerID)
	if !success {
		logging.GetDBLogger().LogCall(callID, workerID, "download_failed", 0, 0, "", false, 0, "", "", errorMsg)
		var es services.EmailService
		es.SendFailureAlert(callID, errorMsg, "Download Failed")
		callTracker.MarkCompleted(callID, false)
		return
	}

	// 2. Transcribe with Groq
	var gs services.GroqService
	transcript, status, _, apiCalls := gs.TranscribeAudio(audioFile, callID)
	totalAPICalls += apiCalls

	// 3. Default values
	audioQuality := "good"
	summaryGenerated := false
	summary := ""

	// 4. Generate Summary based on status
	if status == "success" && len(strings.Fields(transcript)) >= 10 {
		summary, summaryGenerated = gs.GenerateSummary(transcript, callID)
		if summaryGenerated {
			totalAPICalls++
		}
	} else if status == "success" {
		summary = "PURPOSE: Brief conversation\nOUTCOME: Limited content for summary"
	} else if status == "no_speech" {
		transcript = "No clear speech detected in recording"
		summary = "PURPOSE: No speech detected\nOUTCOME: Recording appears silent or empty"
		audioQuality = "silent"
	} else if status == "unclear_audio" {
		transcript = "No clear speech detected in recording"
		summary = "PURPOSE: Audio quality too poor\nOUTCOME: Transcription not possible"
		audioQuality = "unclear"
	} else {
		transcript = "Transcription unavailable due to technical issue"
		summary = "PURPOSE: Processing failed\nOUTCOME: Please check manually"
		audioQuality = "error"
		var es services.EmailService
		es.SendFailureAlert(callID, fmt.Sprintf("Transcription failed: %s", status), "Processing Error")
	}

	// 5. Update Zoho CRM
	var zs services.ZohoService
	zohoSuccess, zohoErr := zs.UpdateCall(callID, transcript, summary)
	if !zohoSuccess {
		logStreamer.Warning("Worker", fmt.Sprintf("Call %s: Zoho update failed (but transcription ok): %s", callID, zohoErr))
	}

	// 6. Log to DB
	duration := time.Since(startTime).Seconds()
	wordCount := 0
	if !strings.Contains(transcript, "No clear speech") && !strings.Contains(strings.ToLower(transcript), "unavailable") {
		wordCount = len(strings.Fields(transcript))
	}

	logging.GetDBLogger().LogCall(
		callID, workerID, status, duration, wordCount,
		audioQuality, summaryGenerated, totalAPICalls,
		truncateString(transcript, 1000), summary, "",
	)

	callTracker.MarkCompleted(callID, true)
	logStreamer.Info("Worker", fmt.Sprintf("Worker %d: SUCCESS -> Call %s | %.1fs | %d words", workerID, callID, duration, wordCount))
}

func truncateString(s string, maxLen int) string {
	if len(s) > maxLen {
		return s[:maxLen]
	}
	return s
}
