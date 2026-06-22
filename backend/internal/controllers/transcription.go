package controllers

import (
	"fmt"

	"transcription-goserver/internal/logging"
	"transcription-goserver/internal/models"
	"transcription-goserver/internal/services"
)

var TaskQueue chan Job

type Job struct {
	CallID string
	RecURL string
}

func InitTaskQueue(size int) {
	TaskQueue = make(chan Job, size)
}

type TranscriptionController struct{}

func (tc *TranscriptionController) HandleOAuthCallback(code string) (map[string]string, error) {
	logStreamer := logging.GetLogStreamer()

	if code != "" {
		var zs services.ZohoService
		err := zs.GenerateAccessToken(code)
		if err != nil {
			logStreamer.Error("OAuth", fmt.Sprintf("OAuth error: %v", err))
			return nil, err
		}
		logStreamer.Info("OAuth", "Tokens saved successfully")
		return map[string]string{"message": "Tokens saved successfully!"}, nil
	}

	return map[string]string{"message": "Send POST with callId and recUrl"}, nil
}

func (tc *TranscriptionController) SubmitTranscription(req TranscriptionRequest) (TranscriptionResponse, error) {
	callTracker := models.GetCallTracker()
	logStreamer := logging.GetLogStreamer()

	if callTracker.IsProcessing(req.CallID) {
		logStreamer.Info("API", fmt.Sprintf("Call %s already processing", req.CallID))
		return TranscriptionResponse{Status: "already_processing", CallID: req.CallID}, nil
	}

	if callTracker.IsCompleted(req.CallID) {
		logStreamer.Info("API", fmt.Sprintf("Call %s already completed", req.CallID))
		return TranscriptionResponse{Status: "already_completed", CallID: req.CallID}, nil
	}

	TaskQueue <- Job{CallID: req.CallID, RecURL: req.RecURL}
	logStreamer.Info("API", fmt.Sprintf("Call %s queued for processing", req.CallID))

	queuePos := len(TaskQueue)
	return TranscriptionResponse{
		Status:        "queued",
		CallID:        req.CallID,
		QueuePosition: &queuePos,
	}, nil
}
