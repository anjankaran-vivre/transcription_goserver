package services

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"

	"transcription-goserver/internal/config"
	"transcription-goserver/internal/logging"
	"transcription-goserver/internal/models"
)

type OpenRouterService struct{}

const (
	TranscriptionEndpoint = "https://openrouter.ai/api/v1/audio/transcriptions"
	ChatEndpoint          = "https://openrouter.ai/api/v1/chat/completions"
)

func (ors *OpenRouterService) TranscribeAudio(audioFilePath, callID string) (string, string, string, int) {
	logStreamer := logging.GetLogStreamer()
	metrics := models.GetMetricsTracker()
	apiCalls := 0

	fileInfo, err := os.Stat(audioFilePath)
	if err != nil {
		logStreamer.Error("OpenRouterService", fmt.Sprintf("Call %s: Cannot stat file: %v", callID, err))
		return "", "error", "", apiCalls
	}

	fileSize := fileInfo.Size()
	logStreamer.Debug("OpenRouterService", fmt.Sprintf("Call %s: File size %d bytes", callID, fileSize))

	if fileSize < 500 {
		logStreamer.Warning("OpenRouterService", fmt.Sprintf("Call %s: Audio too small (<500 bytes)", callID))
		return "", "no_speech", "", apiCalls
	}

	transcriptionModel := strings.TrimSpace(config.Settings.TranscriptionModel)
	if transcriptionModel == "" {
		logStreamer.Error("OpenRouterService", fmt.Sprintf("Call %s: TRANSCRIPTION_MODEL is not configured", callID))
		metrics.RecordError()
		return "", "error", "", apiCalls
	}

	logStreamer.Info("OpenRouterService", fmt.Sprintf("Call %s: Starting transcription via OpenRouter (%s)", callID, transcriptionModel))

	audioBytes, err := os.ReadFile(audioFilePath)
	if err != nil {
		logStreamer.Error("OpenRouterService", fmt.Sprintf("Call %s: Read error: %v", callID, err))
		return "", "error", "", apiCalls
	}

	fileExt := detectAudioFormat(audioBytes)
	if fileExt == "" {
		headerLen := len(audioBytes)
		if headerLen > 12 {
			headerLen = 12
		}
		logStreamer.Error("OpenRouterService", fmt.Sprintf("Call %s: Unsupported audio format, first bytes: % X", callID, audioBytes[:headerLen]))
		return "", "error", "", apiCalls
	}
	logStreamer.Debug("OpenRouterService", fmt.Sprintf("Call %s: Detected audio format %s", callID, fileExt))

	if fileExt != "wav" {
		logStreamer.Info("OpenRouterService", fmt.Sprintf("Call %s: Converting %s audio to wav before OpenRouter request", callID, fileExt))
		audioBytes, err = convertAudioToWAV(audioFilePath, callID, fileExt)
		if err != nil {
			logStreamer.Error("OpenRouterService", fmt.Sprintf("Call %s: WAV conversion failed: %v", callID, err))
			metrics.RecordError()
			return "", "error", "", apiCalls
		}
		fileExt = "wav"
		logStreamer.Info("OpenRouterService", fmt.Sprintf("Call %s: Converted audio to wav (%d bytes)", callID, len(audioBytes)))
	}

	audioData := base64.StdEncoding.EncodeToString(audioBytes)
	attempts := []transcriptionAttempt{
		transcriptionAttempt{
			name:    "verbose with Azure options",
			payload: buildTranscriptionPayload(transcriptionModel, audioData, fileExt, true, true),
		},
		transcriptionAttempt{
			name:    "verbose without provider options",
			payload: buildTranscriptionPayload(transcriptionModel, audioData, fileExt, true, false),
		},
		transcriptionAttempt{
			name:    "basic json",
			payload: buildTranscriptionPayload(transcriptionModel, audioData, fileExt, false, false),
		},
	}
	transcriptionFallbackModel := strings.TrimSpace(config.Settings.TranscriptionFallbackModel)
	if transcriptionFallbackModel != "" && transcriptionFallbackModel != transcriptionModel {
		attempts = append(attempts, transcriptionAttempt{
			name:    fmt.Sprintf("fallback model %s basic json", transcriptionFallbackModel),
			payload: buildTranscriptionPayload(transcriptionFallbackModel, audioData, fileExt, false, false),
		})
	}

	var respBody []byte
	for i, attempt := range attempts {
		if i > 0 {
			logStreamer.Warning("OpenRouterService", fmt.Sprintf("Call %s: Retrying transcription with %s", callID, attempt.name))
		}

		respBody, err = ors.postTranscription(attempt.payload, callID, attempt.name)
		if err == nil {
			break
		}

		logStreamer.Error("OpenRouterService", fmt.Sprintf("Call %s: Transcription attempt failed (%s): %v", callID, attempt.name, err))
		if i == len(attempts)-1 {
			metrics.RecordError()
			return "", "error", "", apiCalls
		}
	}

	apiCalls = 1
	metrics.RecordAPICall(0, "transcription")

	var rawTranscript string
	rawTranscript, err = extractTextFromResponse(respBody)
	if err != nil {
		logStreamer.Error("OpenRouterService", fmt.Sprintf("Call %s: Parse error: %v", callID, err))
		return "", "error", "", apiCalls
	}

	wordCount := len(strings.Fields(rawTranscript))
	logStreamer.Info("OpenRouterService", fmt.Sprintf("Call %s: Raw transcript (%d words)", callID, wordCount))

	if rawTranscript == "" {
		return "", "no_speech", "", apiCalls
	}

	var qc QualityChecker
	isClear, reason := qc.CheckAudioQuality(rawTranscript, callID)
	if !isClear {
		logStreamer.Warning("OpenRouterService", fmt.Sprintf("Call %s: Unclear audio -> %s", callID, reason))
		return "", "unclear_audio", rawTranscript, apiCalls
	}

	cleanTranscript := qc.CleanTranscript(rawTranscript)
	cleanWordCount := len(strings.Fields(cleanTranscript))
	logStreamer.Info("OpenRouterService", fmt.Sprintf("Call %s: Transcription SUCCESS (%d words)", callID, cleanWordCount))

	return cleanTranscript, "success", rawTranscript, apiCalls
}

func (ors *OpenRouterService) postTranscription(payload map[string]interface{}, callID, attemptName string) ([]byte, error) {
	logStreamer := logging.GetLogStreamer()

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshal error: %w", err)
	}

	req, err := http.NewRequest("POST", TranscriptionEndpoint, bytes.NewReader(payloadBytes))
	if err != nil {
		return nil, fmt.Errorf("request error: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+config.Settings.OpenRouterAPIKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 180 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		return nil, fmt.Errorf("read response error: %w", readErr)
	}

	generationID := resp.Header.Get("X-Generation-Id")
	if resp.StatusCode != http.StatusOK {
		errText := string(respBody)
		if len(errText) > 500 {
			errText = errText[:500]
		}
		if generationID != "" {
			return nil, fmt.Errorf("api error %d on %s (generation %s): %s", resp.StatusCode, attemptName, generationID, errText)
		}
		return nil, fmt.Errorf("api error %d on %s: %s", resp.StatusCode, attemptName, errText)
	}

	if generationID != "" {
		logStreamer.Debug("OpenRouterService", fmt.Sprintf("Call %s: OpenRouter generation id %s", callID, generationID))
	}
	return respBody, nil
}

type transcriptionAttempt struct {
	name    string
	payload map[string]interface{}
}

func buildTranscriptionPayload(model, audioData, fileExt string, verbose, withProviderOptions bool) map[string]interface{} {
	payload := map[string]interface{}{
		"model": model,
		"input_audio": map[string]string{
			"data":   audioData,
			"format": fileExt,
		},
	}

	if verbose {
		payload["response_format"] = "verbose_json"
		payload["timestamp_granularities"] = []string{"segment", "word"}
	} else {
		payload["response_format"] = "json"
	}

	if withProviderOptions {
		payload["provider"] = microsoftTranscribeProviderOptions()
	}

	return payload
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

func detectAudioFormat(audioBytes []byte) string {
	if len(audioBytes) >= 12 && string(audioBytes[0:4]) == "RIFF" && string(audioBytes[8:12]) == "WAVE" {
		return "wav"
	}
	if len(audioBytes) >= 4 && string(audioBytes[0:4]) == "fLaC" {
		return "flac"
	}
	if len(audioBytes) >= 4 && string(audioBytes[0:4]) == "OggS" {
		return "ogg"
	}
	if len(audioBytes) >= 4 && bytes.Equal(audioBytes[0:4], []byte{0x1A, 0x45, 0xDF, 0xA3}) {
		return "webm"
	}
	if len(audioBytes) >= 3 && string(audioBytes[0:3]) == "ID3" {
		return "mp3"
	}
	if len(audioBytes) >= 2 {
		if audioBytes[0] == 0xFF && (audioBytes[1]&0xE0) == 0xE0 {
			if (audioBytes[1] & 0x06) == 0x00 {
				return "aac"
			}
			return "mp3"
		}
	}
	if len(audioBytes) >= 12 && bytes.Equal(audioBytes[4:8], []byte("ftyp")) {
		return "m4a"
	}
	return ""
}

func convertAudioToWAV(audioFilePath, callID, fileExt string) ([]byte, error) {
	if fileExt == "m4a" || fileExt == "mp4" {
		wavBytes, err := convertM4AToWAVInGo(audioFilePath)
		if err == nil {
			return wavBytes, nil
		}

		externalBytes, externalErr := convertAudioToWAVWithCommand(audioFilePath, callID)
		if externalErr == nil {
			return externalBytes, nil
		}
		return nil, fmt.Errorf("go m4a decoder failed: %v; external converter failed: %v", err, externalErr)
	}

	return convertAudioToWAVWithCommand(audioFilePath, callID)
}

func convertAudioToWAVWithCommand(audioFilePath, callID string) ([]byte, error) {
	tmpFile, err := os.CreateTemp("", callID+"-*.wav")
	if err != nil {
		return nil, err
	}
	outputPath := tmpFile.Name()
	tmpFile.Close()
	defer os.Remove(outputPath)

	ctx, cancel := context.WithTimeout(context.Background(), 180*time.Second)
	defer cancel()

	cmd := exec.CommandContext(
		ctx,
		config.Settings.AudioConverterPath,
		"-y",
		"-hide_banner",
		"-loglevel", "error",
		"-i", audioFilePath,
		"-vn",
		"-ac", "1",
		"-ar", "16000",
		"-acodec", "pcm_s16le",
		outputPath,
	)

	output, err := cmd.CombinedOutput()
	if ctx.Err() == context.DeadlineExceeded {
		return nil, fmt.Errorf("%s timed out", config.Settings.AudioConverterPath)
	}
	if err != nil {
		errText := strings.TrimSpace(string(output))
		if len(errText) > 500 {
			errText = errText[:500]
		}
		if errText == "" {
			errText = err.Error()
		}
		return nil, fmt.Errorf("%s failed: %s", config.Settings.AudioConverterPath, errText)
	}

	wavBytes, err := os.ReadFile(outputPath)
	if err != nil {
		return nil, err
	}
	if len(wavBytes) < 500 {
		return nil, fmt.Errorf("converter produced too-small wav (%d bytes)", len(wavBytes))
	}
	if detectAudioFormat(wavBytes) != "wav" {
		return nil, fmt.Errorf("converter output was not a wav file")
	}
	return wavBytes, nil
}

func microsoftTranscribeProviderOptions() map[string]interface{} {
	azureOptions := map[string]interface{}{
		"diarization": map[string]bool{
			"enabled": true,
		},
		"enhancedMode": map[string]interface{}{
			"modelOptions": map[string]string{
				"transcribeStyle": "clean",
			},
		},
	}

	if len(config.Settings.TranscriptionPhrases) > 0 {
		azureOptions["phraseList"] = map[string][]string{
			"phrases": config.Settings.TranscriptionPhrases,
		}
	}

	return map[string]interface{}{
		"options": map[string]interface{}{
			"azure": azureOptions,
		},
	}
}

func (ors *OpenRouterService) TranslateAndSummarizeTranscript(transcript, callID string) (string, string, bool) {
	logStreamer := logging.GetLogStreamer()
	metrics := models.GetMetricsTracker()

	transcript = strings.TrimSpace(transcript)
	if transcript == "" {
		return "", "", false
	}

	timeout := time.Duration(config.Settings.TranslationTimeoutSeconds) * time.Second
	if timeout <= 0 {
		timeout = 240 * time.Second
	}

	model := strings.TrimSpace(config.Settings.SummaryModel)
	if model == "" {
		logStreamer.Error("OpenRouterService", fmt.Sprintf("Call %s: SUMMARY_MODEL is not configured", callID))
		metrics.RecordError()
		return "", "", false
	}

	if len(transcript) > 18000 {
		transcript = transcript[:18000] + "... [truncated]"
	}

	logStreamer.Info("OpenRouterService", fmt.Sprintf("Call %s: Translating transcript to English and generating summary in one OpenRouter chat call (%s, timeout %ds)", callID, model, int(timeout.Seconds())))

	prompt := fmt.Sprintf(`You are a CRM call post-processor.
Translate the transcript into clear English. If the transcript is already English, keep it in English and only clean obvious punctuation.
Then summarize only what is explicitly present in the transcript.

Rules:
- Return valid JSON only. Do not wrap it in markdown.
- Do not invent names, products, dates, decisions, or next actions.
- Preserve CRM details such as names, phone numbers, products, prices, dates, and commitments.
- The summary must be in English and use exactly these two labels: PURPOSE and OUTCOME.
- If a detail is unclear or absent, write "Not clearly mentioned."

JSON shape:
{
  "english_transcript": "Full English transcript here",
  "summary": "PURPOSE: ...\nOUTCOME: ..."
}

Transcript:
%s`, transcript)

	payload := map[string]interface{}{
		"model": model,
		"messages": []map[string]string{
			{"role": "user", "content": prompt},
		},
		"max_tokens":  6000,
		"temperature": 0.1,
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		logStreamer.Error("OpenRouterService", fmt.Sprintf("Call %s: Post-processing marshal error: %v", callID, err))
		metrics.RecordError()
		return "", "", false
	}

	req, err := http.NewRequest("POST", ChatEndpoint, bytes.NewReader(payloadBytes))
	if err != nil {
		logStreamer.Error("OpenRouterService", fmt.Sprintf("Call %s: Post-processing request error: %v", callID, err))
		metrics.RecordError()
		return "", "", false
	}
	req.Header.Set("Authorization", "Bearer "+config.Settings.OpenRouterAPIKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: timeout}
	resp, err := client.Do(req)
	if err != nil {
		logStreamer.Error("OpenRouterService", fmt.Sprintf("Call %s: Post-processing API request failed: %v", callID, err))
		metrics.RecordError()
		return "", "", false
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		logStreamer.Error("OpenRouterService", fmt.Sprintf("Call %s: Read post-processing response error: %v", callID, err))
		metrics.RecordError()
		return "", "", false
	}

	if resp.StatusCode != http.StatusOK {
		logStreamer.Error("OpenRouterService", fmt.Sprintf("Call %s: Post-processing API error %d: %s", callID, resp.StatusCode, truncateForLog(string(respBody), 700)))
		metrics.RecordError()
		return "", "", false
	}

	content, err := extractFirstChatText(respBody)
	if err != nil {
		logStreamer.Error("OpenRouterService", fmt.Sprintf("Call %s: Parse post-processing response error: %v", callID, err))
		metrics.RecordError()
		return "", "", false
	}

	englishTranscript, summary, err := parsePostProcessingJSON(content)
	if err != nil {
		logStreamer.Warning("OpenRouterService", fmt.Sprintf("Call %s: Invalid post-processing JSON: %v; content=%s", callID, err, truncateForLog(content, 700)))
		metrics.RecordError()
		return "", "", false
	}

	metrics.RecordAPICall(0, "translate_summary")
	logStreamer.Info("OpenRouterService", fmt.Sprintf("Call %s: English transcript and summary generated (%d words)", callID, len(strings.Fields(englishTranscript))))
	return englishTranscript, summary, true
}

func (ors *OpenRouterService) TranslateTranscriptToEnglish(transcript, callID string) (string, bool) {
	logStreamer := logging.GetLogStreamer()
	metrics := models.GetMetricsTracker()

	if strings.TrimSpace(transcript) == "" {
		return transcript, false
	}

	timeout := time.Duration(config.Settings.TranslationTimeoutSeconds) * time.Second
	if timeout <= 0 {
		timeout = 240 * time.Second
	}

	chunks := splitTranscriptChunks(transcript, 3500)
	logStreamer.Info("OpenRouterService", fmt.Sprintf("Call %s: Translating transcript to English (%s, %d chunk(s), timeout %ds)", callID, config.Settings.TranslationModel, len(chunks), int(timeout.Seconds())))

	var translatedParts []string
	for i, chunk := range chunks {
		prompt := fmt.Sprintf(`Translate this transcript chunk into clear English.
If it is already English, return it unchanged except for obvious transcription punctuation cleanup.
Do not summarize, shorten, add labels, add explanations, or invent missing words.
Preserve names, phone numbers, product names, prices, dates, and CRM-relevant details.
Return only the English transcript for this chunk.

Chunk %d of %d:
%s`, i+1, len(chunks), chunk)

		translated, err := ors.requestTranslation(prompt, config.Settings.TranslationModel, timeout, callID)
		if err != nil {
			logStreamer.Error("OpenRouterService", fmt.Sprintf("Call %s: Translation chunk %d/%d failed: %v", callID, i+1, len(chunks), err))
			metrics.RecordError()
			return transcript, false
		}
		translatedParts = append(translatedParts, strings.TrimSpace(translated))
		metrics.RecordAPICall(0, "translation")
	}

	translated := strings.TrimSpace(strings.Join(translatedParts, "\n"))
	logStreamer.Info("OpenRouterService", fmt.Sprintf("Call %s: Transcript translated to English (%d words)", callID, len(strings.Fields(translated))))
	return translated, true
}

func (ors *OpenRouterService) requestTranslation(prompt, model string, timeout time.Duration, callID string) (string, error) {
	payload := map[string]interface{}{
		"model": model,
		"messages": []map[string]string{
			{"role": "user", "content": prompt},
		},
		"max_tokens":  4000,
		"temperature": 0.1,
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("marshal error: %w", err)
	}

	req, err := http.NewRequest("POST", ChatEndpoint, bytes.NewReader(payloadBytes))
	if err != nil {
		return "", fmt.Errorf("request error: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+config.Settings.OpenRouterAPIKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: timeout}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("api request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read response error: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		errText := string(respBody)
		if len(errText) > 500 {
			errText = errText[:500]
		}
		return "", fmt.Errorf("api error %d: %s", resp.StatusCode, errText)
	}

	var chatResp struct {
		Choices []struct {
			Message struct {
				Content json.RawMessage `json:"content"`
			} `json:"message"`
			Text         string `json:"text"`
			FinishReason string `json:"finish_reason"`
		} `json:"choices"`
		Text   string `json:"text"`
		Output string `json:"output"`
	}

	if err := json.Unmarshal(respBody, &chatResp); err != nil || len(chatResp.Choices) == 0 {
		if err != nil {
			return "", fmt.Errorf("parse response error: %w", err)
		}
		return "", fmt.Errorf("parse response error: no choices")
	}

	translated := strings.TrimSpace(extractChatContent(chatResp.Choices[0].Message.Content))
	if translated == "" {
		translated = strings.TrimSpace(chatResp.Choices[0].Text)
	}
	if translated == "" {
		translated = strings.TrimSpace(chatResp.Text)
	}
	if translated == "" {
		translated = strings.TrimSpace(chatResp.Output)
	}
	if translated == "" {
		return "", fmt.Errorf("empty response content, finish_reason=%q, body=%s", chatResp.Choices[0].FinishReason, truncateForLog(string(respBody), 700))
	}
	return translated, nil
}

func extractFirstChatText(respBody []byte) (string, error) {
	var chatResp struct {
		Choices []struct {
			Message struct {
				Content json.RawMessage `json:"content"`
			} `json:"message"`
			Text         string `json:"text"`
			FinishReason string `json:"finish_reason"`
		} `json:"choices"`
		Text   string `json:"text"`
		Output string `json:"output"`
	}

	if err := json.Unmarshal(respBody, &chatResp); err != nil {
		return "", fmt.Errorf("parse response error: %w", err)
	}

	if len(chatResp.Choices) > 0 {
		content := strings.TrimSpace(extractChatContent(chatResp.Choices[0].Message.Content))
		if content == "" {
			content = strings.TrimSpace(chatResp.Choices[0].Text)
		}
		if content != "" {
			return content, nil
		}
		return "", fmt.Errorf("empty response content, finish_reason=%q, body=%s", chatResp.Choices[0].FinishReason, truncateForLog(string(respBody), 700))
	}

	if content := strings.TrimSpace(chatResp.Text); content != "" {
		return content, nil
	}
	if content := strings.TrimSpace(chatResp.Output); content != "" {
		return content, nil
	}
	return "", fmt.Errorf("parse response error: no choices")
}

func parsePostProcessingJSON(content string) (string, string, error) {
	cleaned := stripJSONFence(content)
	if start := strings.Index(cleaned, "{"); start >= 0 {
		if end := strings.LastIndex(cleaned, "}"); end > start {
			cleaned = cleaned[start : end+1]
		}
	}

	var result struct {
		EnglishTranscript string `json:"english_transcript"`
		Summary           string `json:"summary"`
		Purpose           string `json:"purpose"`
		Outcome           string `json:"outcome"`
	}
	if err := json.Unmarshal([]byte(cleaned), &result); err != nil {
		return "", "", fmt.Errorf("json parse error: %w", err)
	}

	englishTranscript := strings.TrimSpace(result.EnglishTranscript)
	summary := strings.TrimSpace(result.Summary)
	if summary == "" && (strings.TrimSpace(result.Purpose) != "" || strings.TrimSpace(result.Outcome) != "") {
		purpose := strings.TrimSpace(result.Purpose)
		outcome := strings.TrimSpace(result.Outcome)
		if purpose == "" {
			purpose = "Not clearly mentioned."
		}
		if outcome == "" {
			outcome = "Not clearly mentioned."
		}
		summary = fmt.Sprintf("PURPOSE: %s\nOUTCOME: %s", purpose, outcome)
	}

	if englishTranscript == "" {
		return "", "", fmt.Errorf("english_transcript is empty")
	}
	if summary == "" {
		return "", "", fmt.Errorf("summary is empty")
	}

	upperSummary := strings.ToUpper(summary)
	if !strings.Contains(upperSummary, "PURPOSE:") || !strings.Contains(upperSummary, "OUTCOME:") {
		summary = "PURPOSE: Not clearly mentioned.\nOUTCOME: " + summary
	}

	return englishTranscript, summary, nil
}

func stripJSONFence(content string) string {
	cleaned := strings.TrimSpace(content)
	if !strings.HasPrefix(cleaned, "```") {
		return cleaned
	}

	lines := strings.Split(cleaned, "\n")
	if len(lines) == 0 {
		return cleaned
	}
	lines = lines[1:]
	if len(lines) > 0 && strings.HasPrefix(strings.TrimSpace(lines[len(lines)-1]), "```") {
		lines = lines[:len(lines)-1]
	}
	return strings.TrimSpace(strings.Join(lines, "\n"))
}

func extractChatContent(raw json.RawMessage) string {
	if len(raw) == 0 || string(raw) == "null" {
		return ""
	}

	var text string
	if err := json.Unmarshal(raw, &text); err == nil {
		return text
	}

	var parts []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	}
	if err := json.Unmarshal(raw, &parts); err == nil {
		var out []string
		for _, part := range parts {
			if part.Text != "" {
				out = append(out, part.Text)
			}
		}
		return strings.Join(out, "\n")
	}

	return ""
}

func splitTranscriptChunks(transcript string, maxChars int) []string {
	words := strings.Fields(transcript)
	if len(words) == 0 {
		return nil
	}

	var chunks []string
	var current strings.Builder
	for _, word := range words {
		if current.Len() > 0 && current.Len()+1+len(word) > maxChars {
			chunks = append(chunks, current.String())
			current.Reset()
		}
		if current.Len() > 0 {
			current.WriteString(" ")
		}
		current.WriteString(word)
	}
	if current.Len() > 0 {
		chunks = append(chunks, current.String())
	}
	return chunks
}

func truncateForLog(value string, maxLen int) string {
	if len(value) > maxLen {
		return value[:maxLen]
	}
	return value
}

func (ors *OpenRouterService) GenerateSummary(transcript, callID string) (string, bool) {
	logStreamer := logging.GetLogStreamer()
	metrics := models.GetMetricsTracker()

	logStreamer.Info("OpenRouterService", fmt.Sprintf("Call %s: Generating summary", callID))

	if len(transcript) > 10000 {
		transcript = transcript[:10000] + "... [truncated]"
	}

	prompt := fmt.Sprintf(`You are a CRM call summarizer.
Summarize ONLY based on the transcript below — do not invent or assume any context.
Provide two concise sections: PURPOSE and OUTCOME.

Guidelines:
- Write the summary in English.
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
		"model": config.Settings.SummaryModel,
		"messages": []map[string]string{
			{"role": "user", "content": prompt},
		},
		"max_tokens":  300,
		"temperature": 0.4,
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		logStreamer.Error("OpenRouterService", fmt.Sprintf("Call %s: Marshal error: %v", callID, err))
		metrics.RecordError()
		return "PURPOSE: Failed to generate\nOUTCOME: See transcript", false
	}

	req, err := http.NewRequest("POST", ChatEndpoint, bytes.NewReader(payloadBytes))
	if err != nil {
		logStreamer.Error("OpenRouterService", fmt.Sprintf("Call %s: Request error: %v", callID, err))
		metrics.RecordError()
		return "PURPOSE: Not available\nOUTCOME: See transcript", false
	}
	req.Header.Set("Authorization", "Bearer "+config.Settings.OpenRouterAPIKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		logStreamer.Error("OpenRouterService", fmt.Sprintf("Call %s: Summary API request failed: %v", callID, err))
		metrics.RecordError()
		return "PURPOSE: Not available\nOUTCOME: See transcript", false
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		logStreamer.Error("OpenRouterService", fmt.Sprintf("Call %s: Summary API error %d", callID, resp.StatusCode))
		metrics.RecordError()
		return "PURPOSE: Failed to generate\nOUTCOME: See transcript", false
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		logStreamer.Error("OpenRouterService", fmt.Sprintf("Call %s: Read response error: %v", callID, err))
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
		logStreamer.Error("OpenRouterService", fmt.Sprintf("Call %s: Parse summary response error: %v", callID, err))
		metrics.RecordError()
		return "PURPOSE: Not available\nOUTCOME: See transcript", false
	}

	summary := strings.TrimSpace(chatResp.Choices[0].Message.Content)
	metrics.RecordAPICall(0, "summary")
	logStreamer.Info("OpenRouterService", fmt.Sprintf("Call %s: Summary generated", callID))

	return summary, true
}
