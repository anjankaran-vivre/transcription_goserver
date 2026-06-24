package logging

import (
	"log"
	"strings"
	"sync"
	"time"

	"transcription-goserver/internal/database"
	"transcription-goserver/internal/models"
)

type DBLogger struct{}

var dbLoggerInstance *DBLogger
var dbLoggerOnce sync.Once

func GetDBLogger() *DBLogger {
	dbLoggerOnce.Do(func() {
		dbLoggerInstance = &DBLogger{}
	})
	return dbLoggerInstance
}

func (d *DBLogger) LogCall(callID string, workerID int, status string, duration float64, wordCount int,
	audioQuality string, summaryGenerated bool, apiCalls int, transcription string, summary string, errorMsg string) {

	entry := models.CallLog{
		Timestamp:        time.Now(),
		CallID:           callID,
		WorkerID:         workerID,
		Status:           status,
		DurationSec:      duration,
		WordCount:        wordCount,
		AudioQuality:     audioQuality,
		SummaryGenerated: summaryGenerated,
		APICalls:         apiCalls,
		Transcription:    transcription,
		Summary:          summary,
		ErrorMessage:     errorMsg,
	}

	db := database.DB
	var existing models.CallLog
	result := db.Where("call_id = ?", callID).First(&existing)

	if result.Error == nil {
		existing.Timestamp = time.Now()
		existing.WorkerID = workerID
		existing.Status = status
		existing.DurationSec = duration
		existing.WordCount = wordCount
		existing.AudioQuality = audioQuality
		existing.SummaryGenerated = summaryGenerated
		existing.APICalls = apiCalls
		existing.Transcription = transcription
		existing.Summary = summary
		existing.ErrorMessage = errorMsg
		if err := db.Save(&existing).Error; err != nil {
			log.Printf("Error updating call log: %v", err)
		} else {
			log.Printf("✓ Updated call log for %s", callID)
		}
	} else {
		if err := db.Create(&entry).Error; err != nil {
			log.Printf("Error creating call log: %v", err)
			if strings.Contains(err.Error(), "UNIQUE") || strings.Contains(err.Error(), "unique") {
				db.Where("call_id = ?", callID).Assign(entry).FirstOrCreate(&entry)
			}
		} else {
			log.Printf("✓ Created new call log for %s", callID)
		}
	}
}

func (d *DBLogger) LogSystem(level, component, message string, threadID string) {
	entry := models.SystemLog{
		Timestamp: time.Now(),
		Level:     level,
		ThreadID:  threadID,
		Component: component,
		Message:   strings.NewReplacer("\n", " ", "\r", "").Replace(message),
	}

	db := database.DB
	if err := db.Create(&entry).Error; err != nil {
		log.Printf("Error logging system: %v", err)
	}
}

type CallStatsData struct {
	TotalCalls      int     `json:"total_calls"`
	SuccessfulCalls int     `json:"successful_calls"`
	FailedCalls     int     `json:"failed_calls"`
	TodayCalls      int     `json:"today_calls"`
	AvgDurationSec  float64 `json:"avg_duration_sec"`
	TotalWords      int     `json:"total_words"`
	RecentCalls24h  int     `json:"recent_calls_24h"`
}

func (d *DBLogger) GetStatsFromLogs() CallStatsData {
	db := database.DB
	var stats CallStatsData

	db.Model(&models.CallLog{}).Select("COUNT(*)").Scan(&stats.TotalCalls)
	db.Model(&models.CallLog{}).Where("status = ?", "success").Select("COUNT(*)").Scan(&stats.SuccessfulCalls)
	db.Model(&models.CallLog{}).Where("status IN ?", []string{"error", "download_failed"}).Select("COUNT(*)").Scan(&stats.FailedCalls)

	today := time.Now().Truncate(24 * time.Hour)
	db.Model(&models.CallLog{}).Where("timestamp >= ?", today).Select("COUNT(*)").Scan(&stats.TodayCalls)

	db.Model(&models.CallLog{}).Where("duration_sec > 0").Select("AVG(duration_sec)").Scan(&stats.AvgDurationSec)

	db.Model(&models.CallLog{}).Select("COALESCE(SUM(word_count), 0)").Scan(&stats.TotalWords)

	yesterday := time.Now().Add(-24 * time.Hour)
	db.Model(&models.CallLog{}).Where("timestamp >= ?", yesterday).Select("COUNT(*)").Scan(&stats.RecentCalls24h)

	return stats
}

type CallLogRecord struct {
	Timestamp        string `json:"timestamp"`
	CallID           string `json:"call_id"`
	WorkerID         int    `json:"worker_id"`
	Status           string `json:"status"`
	DurationSec      float64 `json:"duration_sec"`
	WordCount        int    `json:"word_count"`
	AudioQuality     string `json:"audio_quality"`
	SummaryGenerated bool   `json:"summary_generated"`
	APICalls         int    `json:"api_calls"`
	Transcription    string `json:"transcription"`
	Summary          string `json:"summary"`
	ErrorMessage     string `json:"error_message"`
}

func (d *DBLogger) GetRecentCalls(limit int) []CallLogRecord {
	db := database.DB
	var calls []models.CallLog
	db.Order("timestamp DESC").Limit(limit).Find(&calls)

	result := make([]CallLogRecord, 0, len(calls))
	for _, c := range calls {
		transcription := c.Transcription
		if len(transcription) > 200 {
			transcription = transcription[:200]
		}
		summary := c.Summary
		if len(summary) > 200 {
			summary = summary[:200]
		}

		result = append(result, CallLogRecord{
			Timestamp:        c.Timestamp.Format(time.RFC3339),
			CallID:           c.CallID,
			WorkerID:         c.WorkerID,
			Status:           c.Status,
			DurationSec:      c.DurationSec,
			WordCount:        c.WordCount,
			AudioQuality:     c.AudioQuality,
			SummaryGenerated: c.SummaryGenerated,
			APICalls:         c.APICalls,
			Transcription:    transcription,
			Summary:          summary,
			ErrorMessage:     c.ErrorMessage,
		})
	}
	return result
}

type SystemLogRecord struct {
	Timestamp string `json:"timestamp"`
	Level     string `json:"level"`
	Component string `json:"component"`
	Message   string `json:"message"`
	ThreadID  string `json:"thread_id"`
}

func (d *DBLogger) GetRecentLogs(limit int) []SystemLogRecord {
	db := database.DB
	var logs []models.SystemLog
	db.Order("timestamp DESC").Limit(limit).Find(&logs)

	result := make([]SystemLogRecord, 0, len(logs))
	for _, l := range logs {
		result = append(result, SystemLogRecord{
			Timestamp: l.Timestamp.Format(time.RFC3339),
			Level:     l.Level,
			Component: l.Component,
			Message:   l.Message,
			ThreadID:  l.ThreadID,
		})
	}
	return result
}

// GetCallByID retrieves full call details by call ID
func (d *DBLogger) GetCallByID(callID string) *CallLogRecord {
	db := database.DB
	var call models.CallLog
	result := db.Where("call_id = ?", callID).First(&call)
	
	if result.Error != nil {
		return nil
	}

	return &CallLogRecord{
		Timestamp:        call.Timestamp.Format(time.RFC3339),
		CallID:           call.CallID,
		WorkerID:         call.WorkerID,
		Status:           call.Status,
		DurationSec:      call.DurationSec,
		WordCount:        call.WordCount,
		AudioQuality:     call.AudioQuality,
		SummaryGenerated: call.SummaryGenerated,
		APICalls:         call.APICalls,
		Transcription:    call.Transcription,
		Summary:          call.Summary,
		ErrorMessage:     call.ErrorMessage,
	}
}

// UpdateCallManual updates transcript and summary for a specific call
func (d *DBLogger) UpdateCallManual(callID string, transcription string, summary string) error {
	db := database.DB
	return db.Model(&models.CallLog{}).
		Where("call_id = ?", callID).
		Updates(map[string]interface{}{
			"transcription": transcription,
			"summary":       summary,
			"timestamp":     time.Now(),
		}).Error
}
