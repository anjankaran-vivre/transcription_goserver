package models

import (
	"time"

	"gorm.io/gorm"
)

type CallLog struct {
	ID               uint      `gorm:"primaryKey;autoIncrement" json:"-"`
	Timestamp        time.Time `gorm:"not null" json:"timestamp"`
	CallID           string    `gorm:"column:call_id;size:255;not null;uniqueIndex" json:"call_id"`
	WorkerID         int       `gorm:"column:worker_id;not null" json:"worker_id"`
	Status           string    `gorm:"size:50;not null" json:"status"`
	DurationSec      float64   `gorm:"column:duration_sec;not null" json:"duration_sec"`
	WordCount        int       `gorm:"column:word_count;not null" json:"word_count"`
	AudioQuality     string    `gorm:"column:audio_quality;size:50;not null" json:"audio_quality"`
	SummaryGenerated bool      `gorm:"column:summary_generated;not null" json:"summary_generated"`
	APICalls         int       `gorm:"column:api_calls;not null" json:"api_calls"`
	Transcription    string    `gorm:"type:text" json:"transcription"`
	Summary          string    `gorm:"type:text" json:"summary"`
	ErrorMessage     string    `gorm:"column:error_message;type:text" json:"error_message"`
}

func (CallLog) TableName() string {
	return "call_logs"
}

func MigrateCallLog(db *gorm.DB) error {
	return db.AutoMigrate(&CallLog{})
}

type SystemLog struct {
	ID        uint      `gorm:"primaryKey;autoIncrement" json:"-"`
	Timestamp time.Time `gorm:"not null" json:"timestamp"`
	Level     string    `gorm:"size:50;not null" json:"level"`
	ThreadID  string    `gorm:"column:thread_id;size:255" json:"thread_id"`
	Component string    `gorm:"size:255;not null" json:"component"`
	Message   string    `gorm:"type:text;not null" json:"message"`
}

func (SystemLog) TableName() string {
	return "system_logs"
}

func MigrateSystemLog(db *gorm.DB) error {
	return db.AutoMigrate(&SystemLog{})
}
