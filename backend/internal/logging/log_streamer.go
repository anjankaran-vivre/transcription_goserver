package logging

import (
	"fmt"
	"sync"
	"time"
)

type LogEntry struct {
	Timestamp string `json:"timestamp"`
	Level     string `json:"level"`
	Component string `json:"component"`
	Message   string `json:"message"`
}

type LogBroadcastFunc func(level, component, message string)

type LogStreamer struct {
	mu            sync.Mutex
	buffer        []LogEntry
	maxBufferSize int
	broadcastFn   LogBroadcastFunc
}

var globalLogStreamer *LogStreamer
var logStreamerOnce sync.Once

func GetLogStreamer() *LogStreamer {
	logStreamerOnce.Do(func() {
		globalLogStreamer = &LogStreamer{
			buffer:        make([]LogEntry, 0, 500),
			maxBufferSize: 500,
		}
	})
	return globalLogStreamer
}

func (ls *LogStreamer) SetBroadcastFunc(fn LogBroadcastFunc) {
	ls.mu.Lock()
	defer ls.mu.Unlock()
	ls.broadcastFn = fn
}

func (ls *LogStreamer) Log(level, component, message string) {
	entry := LogEntry{
		Timestamp: time.Now().Format("15:04:05"),
		Level:     level,
		Component: component,
		Message:   message,
	}

	ls.mu.Lock()
	ls.buffer = append(ls.buffer, entry)
	if len(ls.buffer) > ls.maxBufferSize {
		ls.buffer = ls.buffer[len(ls.buffer)-ls.maxBufferSize:]
	}
	broadcastFn := ls.broadcastFn
	ls.mu.Unlock()

	fmt.Printf("[%s] [%s] [%s] %s\n", entry.Timestamp, level, component, message)

	if broadcastFn != nil && component != "Dashboard" {
		go broadcastFn(level, component, message)
	}
}

func (ls *LogStreamer) Info(component, message string) {
	ls.Log("INFO", component, message)
}

func (ls *LogStreamer) Error(component, message string) {
	ls.Log("ERROR", component, message)
}

func (ls *LogStreamer) Warning(component, message string) {
	ls.Log("WARNING", component, message)
}

func (ls *LogStreamer) Debug(component, message string) {
	ls.Log("DEBUG", component, message)
}

func (ls *LogStreamer) GetRecentLogs(count int) []LogEntry {
	ls.mu.Lock()
	defer ls.mu.Unlock()

	if count > len(ls.buffer) {
		count = len(ls.buffer)
	}
	result := make([]LogEntry, count)
	copy(result, ls.buffer[len(ls.buffer)-count:])
	return result
}
