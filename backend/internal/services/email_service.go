package services

import (
	"fmt"
	"net/smtp"
	"strings"

	"transcription-goserver/internal/config"
	"transcription-goserver/internal/logging"
)

type EmailService struct{}

func (es *EmailService) SendAlert(subject, body string, isHTML bool) {
	if !config.Settings.EmailEnabled {
		return
	}

	go func() {
		logStreamer := logging.GetLogStreamer()

		headers := make([]string, 0)
		headers = append(headers, fmt.Sprintf("Subject: [Transcription Server] %s", subject))
		headers = append(headers, fmt.Sprintf("From: %s", config.Settings.EmailSender))
		headers = append(headers, fmt.Sprintf("To: %s", strings.Join(config.Settings.EmailRecipients, ", ")))
		headers = append(headers, "MIME-Version: 1.0")

		contentType := "text/plain; charset=\"UTF-8\""
		if isHTML {
			contentType = "text/html; charset=\"UTF-8\""
		}
		headers = append(headers, fmt.Sprintf("Content-Type: %s", contentType))

		msg := strings.Join(headers, "\r\n") + "\r\n\r\n" + body

		addr := fmt.Sprintf("%s:%d", config.Settings.SMTPServer, config.Settings.SMTPPort)
		auth := smtp.PlainAuth("", config.Settings.EmailSender, config.Settings.EmailPassword, config.Settings.SMTPServer)

		err := smtp.SendMail(addr, auth, config.Settings.EmailSender, config.Settings.EmailRecipients, []byte(msg))
		if err != nil {
			logStreamer.Error("EmailService", fmt.Sprintf("Failed to send email: %v", err))
			return
		}

		logStreamer.Info("EmailService", fmt.Sprintf("Alert sent: %s", subject))
	}()
}

func (es *EmailService) SendFailureAlert(callID, errorMessage, errorType string) {
	subject := fmt.Sprintf("Call Processing Failed - %s", callID)

	body := fmt.Sprintf(`<html>
<body style="font-family: Arial, sans-serif;">
    <h2 style="color: #e74c3c;">⚠️ Call Processing Failed</h2>
    <table style="border-collapse: collapse; width: 100%%;">
        <tr>
            <td style="padding: 10px; border: 1px solid #ddd;"><strong>Call ID:</strong></td>
            <td style="padding: 10px; border: 1px solid #ddd;">%s</td>
        </tr>
        <tr>
            <td style="padding: 10px; border: 1px solid #ddd;"><strong>Error Type:</strong></td>
            <td style="padding: 10px; border: 1px solid #ddd;">%s</td>
        </tr>
        <tr>
            <td style="padding: 10px; border: 1px solid #ddd;"><strong>Error Message:</strong></td>
            <td style="padding: 10px; border: 1px solid #ddd;">%s</td>
        </tr>
    </table>
    <p style="color: #666; margin-top: 20px;">
        Please check the server dashboard for more details.
    </p>
</body>
</html>`, callID, errorType, errorMessage)

	es.SendAlert(subject, body, true)
}

func (es *EmailService) SendDailySummary(stats map[string]interface{}) {
	subject := "Daily Transcription Summary"

	todayCalls := 0
	if v, ok := stats["today_calls"]; ok {
		todayCalls, _ = v.(int)
	}
	todaySuccess := 0
	if v, ok := stats["today_success"]; ok {
		todaySuccess, _ = v.(int)
	}
	failed := 0
	if v, ok := stats["failed"]; ok {
		failed, _ = v.(int)
	}
	apiCalls := 0
	if v, ok := stats["api_calls_today"]; ok {
		apiCalls, _ = v.(int)
	}

	body := fmt.Sprintf(`<html>
<body style="font-family: Arial, sans-serif;">
    <h2 style="color: #3498db;">📊 Daily Transcription Summary</h2>
    <table style="border-collapse: collapse; width: 100%%;">
        <tr style="background-color: #f2f2f2;">
            <td style="padding: 10px; border: 1px solid #ddd;"><strong>Total Calls</strong></td>
            <td style="padding: 10px; border: 1px solid #ddd;">%d</td>
        </tr>
        <tr>
            <td style="padding: 10px; border: 1px solid #ddd;"><strong>Successful</strong></td>
            <td style="padding: 10px; border: 1px solid #ddd; color: #27ae60;">%d</td>
        </tr>
        <tr style="background-color: #f2f2f2;">
            <td style="padding: 10px; border: 1px solid #ddd;"><strong>Failed</strong></td>
            <td style="padding: 10px; border: 1px solid #ddd; color: #e74c3c;">%d</td>
        </tr>
        <tr>
            <td style="padding: 10px; border: 1px solid #ddd;"><strong>API Calls</strong></td>
            <td style="padding: 10px; border: 1px solid #ddd;">%d</td>
        </tr>
    </table>
</body>
</html>`, todayCalls, todaySuccess, failed, apiCalls)

	es.SendAlert(subject, body, true)
}
