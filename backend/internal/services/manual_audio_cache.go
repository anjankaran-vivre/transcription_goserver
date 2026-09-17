package services

import (
	"errors"
	"os"
	"sync"
)

var manualAudioCache = struct {
	sync.Mutex
	files map[string]string
}{
	files: make(map[string]string),
}

func DownloadManualAudio(callID, recURL string) (string, error) {
	CleanupManualAudio(callID)

	var as AudioService
	audioFile, success, errMsg := as.DownloadAudio(recURL, callID, 0)
	if !success {
		return "", errors.New(errMsg)
	}

	manualAudioCache.Lock()
	manualAudioCache.files[callID] = audioFile
	manualAudioCache.Unlock()

	return audioFile, nil
}

func GetManualAudioPath(callID string) (string, bool) {
	manualAudioCache.Lock()
	defer manualAudioCache.Unlock()

	audioFile, ok := manualAudioCache.files[callID]
	if !ok || audioFile == "" {
		return "", false
	}
	if _, err := os.Stat(audioFile); err != nil {
		delete(manualAudioCache.files, callID)
		return "", false
	}
	return audioFile, true
}

func CleanupManualAudio(callID string) bool {
	manualAudioCache.Lock()
	audioFile := manualAudioCache.files[callID]
	delete(manualAudioCache.files, callID)
	manualAudioCache.Unlock()

	var as AudioService
	return as.CleanupAudio(audioFile)
}
