package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"transcription-goserver/internal/config"
	"transcription-goserver/internal/logging"
)

type ZohoService struct{}

type ZohoTokens struct {
	AccessToken  string  `json:"access_token"`
	Scope        string  `json:"scope"`
	APIDomain    string  `json:"api_domain"`
	TokenType    string  `json:"token_type"`
	ExpiresIn    int     `json:"expires_in"`
	RefreshToken string  `json:"refresh_token"`
	CreatedAt    float64 `json:"created_at"`
}

func (zs *ZohoService) SaveTokens(tokens *ZohoTokens) {
	data, _ := json.Marshal(tokens)
	os.WriteFile(config.Settings.TokenFile, data, 0644)
}

func (zs *ZohoService) LoadTokens() (*ZohoTokens, error) {
	data, err := os.ReadFile(config.Settings.TokenFile)
	if err != nil {
		return nil, fmt.Errorf("no tokens file found")
	}
	var tokens ZohoTokens
	if err := json.Unmarshal(data, &tokens); err != nil {
		return nil, fmt.Errorf("invalid tokens file")
	}
	return &tokens, nil
}

func (zs *ZohoService) GenerateAccessToken(grantCode string) error {
	logStreamer := logging.GetLogStreamer()
	url := "https://accounts.zoho.in/oauth/v2/token"

	params := map[string]string{
		"grant_type":    "authorization_code",
		"client_id":     config.Settings.ZohoClientID,
		"client_secret": config.Settings.ZohoClientSecret,
		"redirect_uri":  config.Settings.ZohoRedirectURI,
		"code":          grantCode,
	}

	body, err := zs.postForm(url, params)
	if err != nil {
		return err
	}

	var tokens ZohoTokens
	if err := json.Unmarshal(body, &tokens); err != nil {
		return fmt.Errorf("failed to parse token response: %v", err)
	}
	tokens.CreatedAt = float64(time.Now().Unix())

	zs.SaveTokens(&tokens)
	logStreamer.Info("ZohoService", "Access token saved")
	return nil
}

func (zs *ZohoService) RefreshAccessToken() (*ZohoTokens, error) {
	logStreamer := logging.GetLogStreamer()
	tokens, err := zs.LoadTokens()
	if err != nil {
		return nil, fmt.Errorf("no refresh token found")
	}

	url := "https://accounts.zoho.in/oauth/v2/token"
	params := map[string]string{
		"refresh_token": tokens.RefreshToken,
		"client_id":     config.Settings.ZohoClientID,
		"client_secret": config.Settings.ZohoClientSecret,
		"grant_type":    "refresh_token",
	}

	body, err := zs.postForm(url, params)
	if err != nil {
		return nil, err
	}

	var newTokens ZohoTokens
	if err := json.Unmarshal(body, &newTokens); err != nil {
		return nil, fmt.Errorf("failed to parse refresh response: %v", err)
	}
	newTokens.RefreshToken = tokens.RefreshToken
	newTokens.CreatedAt = float64(time.Now().Unix())

	zs.SaveTokens(&newTokens)
	logStreamer.Info("ZohoService", "Access token refreshed")
	return &newTokens, nil
}

func (zs *ZohoService) GetAccessToken() (string, error) {
	tokens, err := zs.LoadTokens()
	if err != nil {
		return "", fmt.Errorf("no tokens found")
	}

	if float64(time.Now().Unix())-tokens.CreatedAt > 3500 {
		newTokens, err := zs.RefreshAccessToken()
		if err != nil {
			return "", err
		}
		return newTokens.AccessToken, nil
	}

	return tokens.AccessToken, nil
}

func (zs *ZohoService) UpdateCall(callID, transcription, summary string) (bool, string) {
	logStreamer := logging.GetLogStreamer()

	token, err := zs.GetAccessToken()
	if err != nil {
		logStreamer.Error("ZohoService", fmt.Sprintf("Call %s: Token error: %v", callID, err))
		return false, err.Error()
	}

	transcriptionText := transcription
	if transcriptionText == "" {
		transcriptionText = "No clear speech detected"
	}
	if len(transcriptionText) > 2000 {
		transcriptionText = transcriptionText[:2000]
	}

	updateData := map[string]interface{}{
		"data": []map[string]interface{}{
			{
				"id":              callID,
				"Transcription_c": transcriptionText,
				"Summary_c":       summary,
			},
		},
	}

	jsonData, _ := json.Marshal(updateData)
	req, err := http.NewRequest("PUT", "https://www.zohoapis.in/crm/v2/Calls", bytes.NewReader(jsonData))
	if err != nil {
		logStreamer.Error("ZohoService", fmt.Sprintf("Call %s: Request error: %v", callID, err))
		return false, err.Error()
	}

	req.Header.Set("Authorization", "Zoho-oauthtoken "+token)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		logStreamer.Error("ZohoService", fmt.Sprintf("Call %s: Zoho error: %v", callID, err))
		return false, err.Error()
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		logStreamer.Info("ZohoService", fmt.Sprintf("Call %s: Updated Zoho successfully", callID))
		return true, ""
	}

	respBody, _ := io.ReadAll(resp.Body)
	errorMsg := fmt.Sprintf("Status %d: %s", resp.StatusCode, string(respBody))
	logStreamer.Error("ZohoService", fmt.Sprintf("Call %s: Zoho update failed - %s", callID, errorMsg))
	return false, errorMsg
}

func (zs *ZohoService) postForm(url string, params map[string]string) ([]byte, error) {
	formData := bytes.Buffer{}
	for k, v := range params {
		if formData.Len() > 0 {
			formData.WriteString("&")
		}
		formData.WriteString(k)
		formData.WriteString("=")
		formData.WriteString(v)
	}

	resp, err := http.Post(url, "application/x-www-form-urlencoded", &formData)
	if err != nil {
		return nil, fmt.Errorf("request failed: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read failed: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API error %d: %s", resp.StatusCode, string(body))
	}

	return body, nil
}
