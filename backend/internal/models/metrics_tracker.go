package models

import (
	"sync"
	"time"

	"transcription-goserver/internal/config"
)

type MetricsTracker struct {
	mu                sync.Mutex
	dailyAPICalls     map[string]int
	dailyTokens       map[string]int
	minuteCalls       []time.Time
	totalTranscriptions int
	totalSummaries    int
	totalErrors       int
}

var globalMetrics *MetricsTracker
var metricsOnce sync.Once

func GetMetricsTracker() *MetricsTracker {
	metricsOnce.Do(func() {
		globalMetrics = &MetricsTracker{
			dailyAPICalls: make(map[string]int),
			dailyTokens:   make(map[string]int),
		}
	})
	return globalMetrics
}

func (mt *MetricsTracker) RecordAPICall(tokensUsed int, callType string) {
	mt.mu.Lock()
	defer mt.mu.Unlock()

	today := time.Now().Format("2006-01-02")
	now := time.Now()

	mt.dailyAPICalls[today]++
	mt.dailyTokens[today] += tokensUsed
	mt.minuteCalls = append(mt.minuteCalls, now)

	cutoff := time.Now().Add(-60 * time.Second)
	filtered := make([]time.Time, 0)
	for _, t := range mt.minuteCalls {
		if t.After(cutoff) {
			filtered = append(filtered, t)
		}
	}
	mt.minuteCalls = filtered

	switch callType {
	case "transcription":
		mt.totalTranscriptions++
	case "summary":
		mt.totalSummaries++
	}
}

func (mt *MetricsTracker) RecordError() {
	mt.mu.Lock()
	defer mt.mu.Unlock()
	mt.totalErrors++
}

type TodayStats struct {
	APICallsToday      int `json:"api_calls_today"`
	TokensToday        int `json:"tokens_today"`
	CallsThisMinute    int `json:"calls_this_minute"`
	TotalTranscriptions int `json:"total_transcriptions"`
	TotalSummaries     int `json:"total_summaries"`
	TotalErrors        int `json:"total_errors"`
}

func (mt *MetricsTracker) GetTodayStats() TodayStats {
	mt.mu.Lock()
	defer mt.mu.Unlock()

	today := time.Now().Format("2006-01-02")
	return TodayStats{
		APICallsToday:      mt.dailyAPICalls[today],
		TokensToday:        mt.dailyTokens[today],
		CallsThisMinute:    len(mt.minuteCalls),
		TotalTranscriptions: mt.totalTranscriptions,
		TotalSummaries:     mt.totalSummaries,
		TotalErrors:        mt.totalErrors,
	}
}

type RateLimitStatus struct {
	DailyUsed       int     `json:"daily_used"`
	DailyLimit      int     `json:"daily_limit"`
	DailyPercentage float64 `json:"daily_percentage"`
	MinuteUsed      int     `json:"minute_used"`
	MinuteLimit     int     `json:"minute_limit"`
	MinutePercentage float64 `json:"minute_percentage"`
}

func (mt *MetricsTracker) GetRateLimitStatus() RateLimitStatus {
	mt.mu.Lock()
	defer mt.mu.Unlock()

	today := time.Now().Format("2006-01-02")
	dailyUsed := mt.dailyAPICalls[today]
	minuteUsed := len(mt.minuteCalls)

	dailyLimit := config.Settings.GroqDailyLimit
	minuteLimit := config.Settings.GroqMinuteLimit

	if dailyLimit <= 0 {
		dailyLimit = 1
	}
	if minuteLimit <= 0 {
		minuteLimit = 1
	}

	return RateLimitStatus{
		DailyUsed:        dailyUsed,
		DailyLimit:       config.Settings.GroqDailyLimit,
		DailyPercentage:  float64(dailyUsed) / float64(dailyLimit) * 100,
		MinuteUsed:       minuteUsed,
		MinuteLimit:      config.Settings.GroqMinuteLimit,
		MinutePercentage: float64(minuteUsed) / float64(minuteLimit) * 100,
	}
}
