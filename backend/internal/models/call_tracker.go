package models

import (
	"fmt"
	"sync"
	"time"
)

type CallTracker struct {
	mu         sync.Mutex
	processing map[string]bool
	completed  map[string]bool
	failed     map[string]bool
	startTime  time.Time
}

var globalCallTracker *CallTracker
var callTrackerOnce sync.Once

func GetCallTracker() *CallTracker {
	callTrackerOnce.Do(func() {
		globalCallTracker = &CallTracker{
			processing: make(map[string]bool),
			completed:  make(map[string]bool),
			failed:     make(map[string]bool),
			startTime:  time.Now(),
		}
	})
	return globalCallTracker
}

func (ct *CallTracker) TryAcquire(callID string) bool {
	ct.mu.Lock()
	defer ct.mu.Unlock()
	if ct.processing[callID] || ct.completed[callID] {
		return false
	}
	ct.processing[callID] = true
	return true
}

func (ct *CallTracker) MarkCompleted(callID string, success bool) {
	ct.mu.Lock()
	defer ct.mu.Unlock()
	delete(ct.processing, callID)
	if success {
		ct.completed[callID] = true
	} else {
		ct.failed[callID] = true
	}
}

func (ct *CallTracker) IsProcessing(callID string) bool {
	ct.mu.Lock()
	defer ct.mu.Unlock()
	return ct.processing[callID]
}

func (ct *CallTracker) IsCompleted(callID string) bool {
	ct.mu.Lock()
	defer ct.mu.Unlock()
	return ct.completed[callID]
}

type CallStats struct {
	Processing     int      `json:"processing"`
	Completed      int      `json:"completed"`
	Failed         int      `json:"failed"`
	ProcessingIDs  []string `json:"processing_ids"`
}

func (ct *CallTracker) GetStats() CallStats {
	ct.mu.Lock()
	defer ct.mu.Unlock()
	ids := make([]string, 0, len(ct.processing))
	for id := range ct.processing {
		ids = append(ids, id)
	}
	return CallStats{
		Processing:    len(ct.processing),
		Completed:     len(ct.completed),
		Failed:        len(ct.failed),
		ProcessingIDs: ids,
	}
}

func (ct *CallTracker) GetUptime() string {
	delta := time.Since(ct.startTime)
	hours := int(delta.Hours())
	minutes := int(delta.Minutes()) % 60
	seconds := int(delta.Seconds()) % 60
	return fmt.Sprintf("%02d:%02d:%02d", hours, minutes, seconds)
}
