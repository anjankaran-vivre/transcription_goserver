package controllers

import (
	"fmt"
	"os"
	"strings"

	"transcription-goserver/internal/config"
	"transcription-goserver/internal/logging"
	"transcription-goserver/internal/models"
	"transcription-goserver/internal/services"
)

type DashboardController struct{}

func (dc *DashboardController) GetServerStatus() StatusResponse {
	callTracker := models.GetCallTracker()
	logStreamer := logging.GetLogStreamer()

	queueSize := len(TaskQueue)

	resp := StatusResponse{
		Status:     "running",
		Uptime:     callTracker.GetUptime(),
		Workers:    config.Settings.NumWorkers,
		QueueSize:  queueSize,
		MemoryMB:   0,
		CPUPercent: 0,
		PID:        os.Getpid(),
	}

	logStreamer.Debug("Dashboard", "Status fetched")
	return resp
}

func (dc *DashboardController) GetMetrics() MetricsResponse {
	callTracker := models.GetCallTracker()
	metrics := models.GetMetricsTracker()
	dbLogger := logging.GetDBLogger()

	callStats := callTracker.GetStats()
	dbStats := dbLogger.GetStatsFromLogs()
	todayStats := metrics.GetTodayStats()

	return MetricsResponse{
		Realtime:   callStats,
		Historical: dbStats,
		Today:      todayStats,
	}
}

func (dc *DashboardController) GetCalls(limit int) ([]logging.CallLogRecord, error) {
	logStreamer := logging.GetLogStreamer()
	dbLogger := logging.GetDBLogger()

	calls := dbLogger.GetRecentCalls(limit)
	if calls == nil {
		calls = make([]logging.CallLogRecord, 0)
	}
	logStreamer.Debug("Dashboard", fmt.Sprintf("Fetched %d calls", len(calls)))
	return calls, nil
}

func (dc *DashboardController) GetLogs(limit int) ([]logging.SystemLogRecord, error) {
	logStreamer := logging.GetLogStreamer()
	dbLogger := logging.GetDBLogger()

	logs := dbLogger.GetRecentLogs(limit)
	if logs == nil {
		logs = make([]logging.SystemLogRecord, 0)
	}
	logStreamer.Debug("Dashboard", fmt.Sprintf("Fetched %d logs", len(logs)))
	return logs, nil
}

func (dc *DashboardController) GetProcessingStats() map[string]interface{} {
	callTracker := models.GetCallTracker()
	logStreamer := logging.GetLogStreamer()

	stats := callTracker.GetStats()
	result := map[string]interface{}{
		"processing_ids": stats.ProcessingIDs,
		"processing":     stats.Processing,
	}
	logStreamer.Debug("Dashboard", fmt.Sprintf("Processing stats fetched"))
	return result
}

// GetCallByID retrieves a specific call by its ID with full details
func (dc *DashboardController) GetCallByID(callID string) (map[string]interface{}, error) {
	logStreamer := logging.GetLogStreamer()
	dbLogger := logging.GetDBLogger()

	call := dbLogger.GetCallByID(callID)
	if call == nil {
		logStreamer.Warning("Dashboard", fmt.Sprintf("Call not found: %s", callID))
		return map[string]interface{}{"error": "Call not found"}, fmt.Errorf("call not found")
	}

	logStreamer.Debug("Dashboard", fmt.Sprintf("Fetched call details: %s", callID))
	return map[string]interface{}{
		"call": call,
	}, nil
}

// UpdateCallManually updates transcript and summary for a call
func (dc *DashboardController) UpdateCallManually(callID, transcription, summary string) (map[string]interface{}, error) {
	logStreamer := logging.GetLogStreamer()
	dbLogger := logging.GetDBLogger()

	// Validate input
	if callID == "" {
		return map[string]interface{}{"error": "Call ID is required"}, fmt.Errorf("call ID required")
	}

	// Update in database
	err := dbLogger.UpdateCallManual(callID, transcription, summary)
	if err != nil {
		logStreamer.Error("Dashboard", fmt.Sprintf("Failed to update call %s: %v", callID, err))
		return map[string]interface{}{"error": "Failed to update call"}, err
	}

	logStreamer.Info("Dashboard", fmt.Sprintf("Call %s updated manually - transcript and summary changed", callID))
	return map[string]interface{}{
		"message": "Call updated successfully",
		"call_id": callID,
	}, nil
}

// PostToZohoManually sends manually edited call data to Zoho
func (dc *DashboardController) PostToZohoManually(callID, transcription, summary string) (map[string]interface{}, error) {
	logStreamer := logging.GetLogStreamer()
	var zs services.ZohoService

	logStreamer.Info("Dashboard", fmt.Sprintf("Manual Zoho update initiated for call %s", callID))

	// Update Zoho
	success, errMsg := zs.UpdateCall(callID, transcription, summary)
	if !success {
		logStreamer.Error("Dashboard", fmt.Sprintf("Manual Zoho update failed for %s: %s", callID, errMsg))
		return map[string]interface{}{
			"success": false,
			"call_id": callID,
			"error":   errMsg,
		}, fmt.Errorf("zoho update failed: %s", errMsg)
	}

	services.CleanupManualAudio(callID)
	logStreamer.Info("Dashboard", fmt.Sprintf("Manual Zoho update successful for call %s", callID))
	return map[string]interface{}{
		"success": true,
		"call_id": callID,
		"message": "Successfully updated in Zoho CRM",
	}, nil
}

// FetchCallFromZoho fetches call details from Zoho CRM and processes it
func (dc *DashboardController) FetchCallFromZoho(callID string) (map[string]interface{}, error) {
	logStreamer := logging.GetLogStreamer()
	var zs services.ZohoService

	if callID == "" {
		return map[string]interface{}{"error": "Call ID is required"}, fmt.Errorf("call ID required")
	}

	logStreamer.Info("Dashboard", fmt.Sprintf("Fetching call %s from Zoho CRM", callID))

	// Fetch from Zoho
	callData, err := zs.FetchCallFromZoho(callID)
	if err != nil {
		errMsg := err.Error()
		logStreamer.Error("Dashboard", fmt.Sprintf("Failed to fetch call %s from Zoho: %s", callID, errMsg))
		return map[string]interface{}{
			"success": false,
			"call_id": callID,
			"error":   errMsg,
		}, err
	}

	// Extract relevant fields from Zoho response
	zohoCall := map[string]interface{}{
		"call_id":       callID,
		"phone_number":  "",
		"duration":      0,
		"call_url":      "",
		"transcription": "",
		"summary":       "",
		"status":        "fetched_from_zoho",
		"zoho_raw":      callData,
	}

	// Map Zoho fields to our structure
	if phone, ok := callData["Phone"].(string); ok {
		zohoCall["phone_number"] = phone
	}
	if duration, ok := callData["Call_Duration_c"].(float64); ok {
		zohoCall["duration"] = int(duration)
	}
	if recordingURL := extractZohoRecordingURL(callData); recordingURL != "" {
		zohoCall["call_url"] = recordingURL
	}
	if transcript, ok := callData["Transcription_c"].(string); ok {
		zohoCall["transcription"] = transcript
	}
	if summary, ok := callData["Summary_c"].(string); ok {
		zohoCall["summary"] = summary
	}

	logStreamer.Info("Dashboard", fmt.Sprintf("Successfully fetched call %s from Zoho", callID))
	return zohoCall, nil
}

func extractZohoRecordingURL(callData map[string]interface{}) string {
	preferredFields := []string{
		"Call_URL",
		"Call_Recording_URL_c",
		"Call_Recording_URL",
		"Recording_URL_c",
		"Recording_URL",
		"Call_Recording",
		"Voice_Recording__s",
		"$recording_url",
		"recUrl",
		"rec_url",
		"Audio_URL",
	}

	for _, field := range preferredFields {
		if recordingURL := zohoURLValue(callData[field]); recordingURL != "" {
			return recordingURL
		}
	}

	for field, value := range callData {
		normalizedField := strings.ToLower(strings.NewReplacer("_", "", "-", "", " ", "").Replace(field))
		if strings.Contains(normalizedField, "recording") {
			if recordingURL := zohoURLValue(value); recordingURL != "" {
				return recordingURL
			}
		}
	}

	return ""
}

func zohoURLValue(value interface{}) string {
	switch typed := value.(type) {
	case string:
		candidate := strings.TrimSpace(typed)
		if strings.HasPrefix(candidate, "http://") || strings.HasPrefix(candidate, "https://") {
			return candidate
		}
	case map[string]interface{}:
		for _, key := range []string{"url", "href", "value", "link"} {
			if candidate := zohoURLValue(typed[key]); candidate != "" {
				return candidate
			}
		}
	}

	return ""
}

func (dc *DashboardController) DownloadCallAudio(callID, callURL string) (map[string]interface{}, error) {
	logStreamer := logging.GetLogStreamer()

	if callID == "" {
		return map[string]interface{}{"error": "Call ID is required"}, fmt.Errorf("call ID required")
	}
	if callURL == "" {
		return map[string]interface{}{"error": "Call recording URL is required"}, fmt.Errorf("call recording URL required")
	}

	_, err := services.DownloadManualAudio(callID, callURL)
	if err != nil {
		logStreamer.Error("Dashboard", fmt.Sprintf("Failed to download manual audio for %s: %v", callID, err))
		return map[string]interface{}{"error": err.Error()}, err
	}

	logStreamer.Info("Dashboard", fmt.Sprintf("Manual audio cached for call %s", callID))
	return map[string]interface{}{
		"success":   true,
		"call_id":   callID,
		"audio_url": fmt.Sprintf("/api/call/%s/audio", callID),
	}, nil
}

func (dc *DashboardController) CleanupCallAudioCache(callID string) map[string]interface{} {
	services.CleanupManualAudio(callID)
	logging.GetLogStreamer().Info("Dashboard", fmt.Sprintf("Manual audio cache cleared for call %s", callID))
	return map[string]interface{}{
		"success": true,
		"call_id": callID,
	}
}

func (dc *DashboardController) ResendCallForTranscription(callID, callURL string) (map[string]interface{}, error) {
	logStreamer := logging.GetLogStreamer()

	if callID == "" {
		return map[string]interface{}{"error": "Call ID is required"}, fmt.Errorf("call ID required")
	}
	if callURL == "" {
		return map[string]interface{}{"error": "Call recording URL is required"}, fmt.Errorf("call recording URL required")
	}

	audioFile, cached := services.GetManualAudioPath(callID)
	if !cached {
		var err error
		audioFile, err = services.DownloadManualAudio(callID, callURL)
		if err != nil {
			logStreamer.Error("Dashboard", fmt.Sprintf("Manual audio download failed for %s: %v", callID, err))
			return map[string]interface{}{"error": "Failed to download call audio"}, err
		}
	}

	var ors services.OpenRouterService
	transcript, status, rawTranscript, _ := ors.TranscribeAudio(audioFile, callID+"-manual")
	if status != "success" {
		return map[string]interface{}{
			"error":  fmt.Sprintf("Transcription failed with status: %s", status),
			"status": status,
		}, fmt.Errorf("manual transcription failed: %s", status)
	}

	englishTranscript, summary, generated := ors.TranslateAndSummarizeTranscript(transcript, callID+"-manual")
	if !generated {
		return map[string]interface{}{"error": "English translation and summary generation failed"}, fmt.Errorf("manual post-processing failed")
	}
	if rawTranscript == "" {
		rawTranscript = transcript
	}

	if err := logging.GetDBLogger().UpdateManualAIResult(callID, rawTranscript, englishTranscript, summary); err != nil {
		logStreamer.Error("Dashboard", fmt.Sprintf("Failed to save manual AI result for %s: %v", callID, err))
		return map[string]interface{}{"error": "Transcription completed but could not be saved"}, err
	}

	logStreamer.Info("Dashboard", fmt.Sprintf("Manual AI transcription completed for call %s", callID))
	return map[string]interface{}{
		"success":           true,
		"call_id":           callID,
		"raw_transcription": rawTranscript,
		"transcription":     englishTranscript,
		"summary":           summary,
		"message":           "AI transcription refreshed",
	}, nil
}
