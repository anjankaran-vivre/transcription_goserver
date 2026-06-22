package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"strings"
	"time"

	"transcription-goserver/internal/config"
	"transcription-goserver/internal/logging"
	"transcription-goserver/internal/models"
)

type GroqService struct{}

const (
	WhisperEndpoint = "https://api.groq.com/openai/v1/audio/translations"
	ChatEndpoint    = "https://api.groq.com/openai/v1/chat/completions"
)

func (gs *GroqService) TranscribeAudio(audioFilePath, callID string) (string, string, string, int) {
	logStreamer := logging.GetLogStreamer()
	metrics := models.GetMetricsTracker()
	apiCalls := 0

	fileInfo, err := os.Stat(audioFilePath)
	if err != nil {
		logStreamer.Error("GroqService", fmt.Sprintf("Call %s: Cannot stat file: %v", callID, err))
		return "", "error", "", apiCalls
	}

	fileSize := fileInfo.Size()
	logStreamer.Debug("GroqService", fmt.Sprintf("Call %s: File size %d bytes", callID, fileSize))

	if fileSize < 500 {
		logStreamer.Warning("GroqService", fmt.Sprintf("Call %s: Audio too small (<500 bytes)", callID))
		return "", "no_speech", "", apiCalls
	}

	logStreamer.Info("GroqService", fmt.Sprintf("Call %s: Starting transcription via Groq Whisper", callID))

	audioBytes, err := os.ReadFile(audioFilePath)
	if err != nil {
		logStreamer.Error("GroqService", fmt.Sprintf("Call %s: Read error: %v", callID, err))
		return "", "error", "", apiCalls
	}

	fileExt := "mp3"
	if len(audioBytes) >= 3 {
		if (audioBytes[0] == 0xFF && audioBytes[1] == 0xFB) || (audioBytes[0] == 0xFF && audioBytes[1] == 0xF3) || (len(audioBytes) >= 3 && audioBytes[0] == 'I' && audioBytes[1] == 'D' && audioBytes[2] == '3') {
			fileExt = "mp3"
		} else if len(audioBytes) >= 4 && bytes.Contains(audioBytes[:20], []byte("ftyp")) {
			fileExt = "m4a"
		} else if len(audioBytes) >= 12 && string(audioBytes[0:4]) == "RIFF" && string(audioBytes[8:12]) == "WAVE" {
			fileExt = "wav"
		}
	}

	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	part, err := writer.CreateFormFile("file", "recording."+fileExt)
	if err != nil {
		logStreamer.Error("GroqService", fmt.Sprintf("Call %s: Form file error: %v", callID, err))
		return "", "error", "", apiCalls
	}
	part.Write(audioBytes)

	writer.WriteField("model", "whisper-large-v3")
	writer.WriteField("response_format", "verbose_json")
	writer.Close()

	req, err := http.NewRequest("POST", WhisperEndpoint, &buf)
	if err != nil {
		logStreamer.Error("GroqService", fmt.Sprintf("Call %s: Request error: %v", callID, err))
		return "", "error", "", apiCalls
	}
	req.Header.Set("Authorization", "Bearer "+config.Settings.GroqAPIKey)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	client := &http.Client{Timeout: 180 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		logStreamer.Error("GroqService", fmt.Sprintf("Call %s: Whisper API request failed: %v", callID, err))
		metrics.RecordError()
		return "", "error", "", apiCalls
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		errText := string(respBody)
		if len(errText) > 200 {
			errText = errText[:200]
		}
		logStreamer.Error("GroqService", fmt.Sprintf("Call %s: Whisper API error %d: %s", callID, resp.StatusCode, errText))
		metrics.RecordError()
		return "", "error", "", apiCalls
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		logStreamer.Error("GroqService", fmt.Sprintf("Call %s: Read response error: %v", callID, err))
		return "", "error", "", apiCalls
	}

	apiCalls = 1
	metrics.RecordAPICall(0, "transcription")

	var rawTranscript string
	rawTranscript, err = extractTextFromResponse(respBody)
	if err != nil {
		logStreamer.Error("GroqService", fmt.Sprintf("Call %s: Parse error: %v", callID, err))
		return "", "error", "", apiCalls
	}

	wordCount := len(strings.Fields(rawTranscript))
	logStreamer.Info("GroqService", fmt.Sprintf("Call %s: Raw transcript (%d words)", callID, wordCount))

	if rawTranscript == "" {
		return "", "no_speech", "", apiCalls
	}

	var qc QualityChecker
	isClear, reason := qc.CheckAudioQuality(rawTranscript, callID)
	if !isClear {
		logStreamer.Warning("GroqService", fmt.Sprintf("Call %s: Unclear audio -> %s", callID, reason))
		return "", "unclear_audio", rawTranscript, apiCalls
	}

	cleanTranscript := qc.CleanTranscript(rawTranscript)
	cleanWordCount := len(strings.Fields(cleanTranscript))
	logStreamer.Info("GroqService", fmt.Sprintf("Call %s: Transcription SUCCESS (%d words)", callID, cleanWordCount))

	return cleanTranscript, "success", rawTranscript, apiCalls
}

func extractTextFromResponse(body []byte) (string, error) {
	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return "", err
	}

	if text, ok := result["text"].(string); ok && text != "" {
		return strings.TrimSpace(text), nil
	}

	if segments, ok := result["segments"].([]interface{}); ok {
		var parts []string
		for _, seg := range segments {
			if segMap, ok := seg.(map[string]interface{}); ok {
				if text, ok := segMap["text"].(string); ok {
					parts = append(parts, strings.TrimSpace(text))
				}
			}
		}
		if len(parts) > 0 {
			return strings.Join(parts, " "), nil
		}
	}

	return "", nil
}

func (gs *GroqService) GenerateSummary(transcript, callID string) (string, bool) {
	logStreamer := logging.GetLogStreamer()
	metrics := models.GetMetricsTracker()

	logStreamer.Info("GroqService", fmt.Sprintf("Call %s: Generating summary", callID))

	if len(transcript) > 10000 {
		transcript = transcript[:10000] + "... [truncated]"
	}

	prompt := fmt.Sprintf(`You are a CRM call summarizer.
Summarize ONLY based on the transcript below — do not invent or assume any context.
Provide two concise sections: PURPOSE and OUTCOME.

Guidelines:
- Be specific. Include product/service names if mentioned.
- PURPOSE: Main issue, request, or discussion topic.
- OUTCOME: Decisions made, resolutions, or next actions.
- If unclear, say "Not clearly mentioned."
- Each section: 1–2 concise sentences.

Format:
PURPOSE: ...
OUTCOME: ...

Transcript:
%s`, transcript)

	payload := map[string]interface{}{
		"model": "llama-3.1-8b-instant",
		"messages": []map[string]string{
			{"role": "user", "content": prompt},
		},
		"max_tokens":  300,
		"temperature": 0.4,
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		logStreamer.Error("GroqService", fmt.Sprintf("Call %s: Marshal error: %v", callID, err))
		metrics.RecordError()
		return "PURPOSE: Failed to generate\nOUTCOME: See transcript", false
	}

	req, err := http.NewRequest("POST", ChatEndpoint, bytes.NewReader(payloadBytes))
	if err != nil {
		logStreamer.Error("GroqService", fmt.Sprintf("Call %s: Request error: %v", callID, err))
		metrics.RecordError()
		return "PURPOSE: Not available\nOUTCOME: See transcript", false
	}
	req.Header.Set("Authorization", "Bearer "+config.Settings.GroqAPIKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		logStreamer.Error("GroqService", fmt.Sprintf("Call %s: Summary API request failed: %v", callID, err))
		metrics.RecordError()
		return "PURPOSE: Not available\nOUTCOME: See transcript", false
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		logStreamer.Error("GroqService", fmt.Sprintf("Call %s: Summary API error %d", callID, resp.StatusCode))
		metrics.RecordError()
		return "PURPOSE: Failed to generate\nOUTCOME: See transcript", false
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		logStreamer.Error("GroqService", fmt.Sprintf("Call %s: Read response error: %v", callID, err))
		metrics.RecordError()
		return "PURPOSE: Not available\nOUTCOME: See transcript", false
	}

	var chatResp struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}

	if err := json.Unmarshal(respBody, &chatResp); err != nil || len(chatResp.Choices) == 0 {
		logStreamer.Error("GroqService", fmt.Sprintf("Call %s: Parse summary response error: %v", callID, err))
		metrics.RecordError()
		return "PURPOSE: Not available\nOUTCOME: See transcript", false
	}

	summary := strings.TrimSpace(chatResp.Choices[0].Message.Content)
	metrics.RecordAPICall(0, "summary")
	logStreamer.Info("GroqService", fmt.Sprintf("Call %s: Summary generated", callID))

	return summary, true
}
