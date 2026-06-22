package controllers

import (
	"fmt"
	"os"
	"runtime"

	"transcription-goserver/internal/config"
	"transcription-goserver/internal/logging"
	"transcription-goserver/internal/models"
)

type DashboardController struct{}

func (dc *DashboardController) GetServerStatus() StatusResponse {
	callTracker := models.GetCallTracker()
	logStreamer := logging.GetLogStreamer()

	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	memoryMB := float64(m.Alloc) / 1024 / 1024

	queueSize := len(TaskQueue)

	resp := StatusResponse{
		Status:     "running",
		Uptime:     callTracker.GetUptime(),
		Workers:    config.Settings.NumWorkers,
		QueueSize:  queueSize,
		MemoryMB:   memoryMB,
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
	rateLimit := metrics.GetRateLimitStatus()

	return MetricsResponse{
		Realtime:   callStats,
		Historical: dbStats,
		Today:      todayStats,
		RateLimit:  rateLimit,
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
