package config

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	ZohoClientID     string
	ZohoClientSecret string
	ZohoRedirectURI  string
	TokenFile        string

	AudioUsername string
	AudioPassword string

	OpenRouterAPIKey           string
	TranscriptionModel         string
	TranscriptionFallbackModel string
	TranscriptionPhrases       []string
	AudioConverterPath         string
	SummaryModel               string
	TranslationModel           string
	TranslationFallbackModel   string
	TranslationTimeoutSeconds  int

	ProcessDelay       int
	MaxDownloadRetries int
	RetryInterval      int
	NumWorkers         int

	PIDFile string

	EmailEnabled    bool
	SMTPServer      string
	SMTPPort        int
	EmailSender     string
	EmailPassword   string
	EmailRecipients []string

	Host  string
	Port  int
	Debug bool

	DBHost     string
	DBServer   string
	DBName     string
	DBUser     string
	DBPassword string
	DBPort     int

	LogsDir string
	DataDir string
}

var Settings *Config

func Load() *Config {
	godotenv.Load()

	baseDir := defaultBaseDir()

	logsDir := getEnvStr("LOGS_DIR", filepath.Join(baseDir, "logs"))
	dataDir := getEnvStr("DATA_DIR", filepath.Join(baseDir, "data"))
	tokenFile := getEnvStr("ZOHO_TOKEN_FILE", filepath.Join(dataDir, "zoho_tokens.json"))
	os.MkdirAll(logsDir, 0755)
	os.MkdirAll(dataDir, 0755)

	cfg := &Config{
		ZohoClientID:               os.Getenv("ZOHO_CLIENT_ID"),
		ZohoClientSecret:           os.Getenv("ZOHO_CLIENT_SECRET"),
		ZohoRedirectURI:            os.Getenv("ZOHO_REDIRECT_URI"),
		TokenFile:                  tokenFile,
		AudioUsername:              os.Getenv("AUDIO_USERNAME"),
		AudioPassword:              os.Getenv("AUDIO_PASSWORD"),
		OpenRouterAPIKey:           os.Getenv("OPEN_ROUTER_API_KEY"),
		TranscriptionModel:         os.Getenv("TRANSCRIPTION_MODEL"),
		TranscriptionFallbackModel: os.Getenv("TRANSCRIPTION_FALLBACK_MODEL"),
		TranscriptionPhrases:       getEnvList("TRANSCRIPTION_PHRASES"),
		AudioConverterPath:         getEnvStr("AUDIO_CONVERTER_PATH", "ffmpeg"),
		SummaryModel:               os.Getenv("SUMMARY_MODEL"),
		TranslationModel:           os.Getenv("TRANSLATION_MODEL"),
		TranslationFallbackModel:   os.Getenv("TRANSLATION_FALLBACK_MODEL"),
		TranslationTimeoutSeconds:  getEnvInt("TRANSLATION_TIMEOUT_SECONDS", 240),
		ProcessDelay:               getEnvInt("PROCESS_DELAY", 20),
		MaxDownloadRetries:         getEnvInt("MAX_DOWNLOAD_RETRIES", 2),
		RetryInterval:              getEnvInt("RETRY_INTERVAL", 10),
		NumWorkers:                 getEnvInt("NUM_WORKERS", 2),
		PIDFile:                    filepath.Join(dataDir, "server.pid"),
		Host:                       getEnvStr("HOST", "127.0.0.1"),
		Port:                       getEnvInt("PORT", 5050),
		DBHost:                     os.Getenv("DB_HOST"),
		DBServer:                   os.Getenv("DB_SERVER"),
		DBName:                     os.Getenv("DB_NAME"),
		DBUser:                     os.Getenv("DB_USER"),
		DBPassword:                 os.Getenv("DB_PASSWORD"),
		DBPort:                     getEnvInt("DB_PORT", 1433),
		LogsDir:                    logsDir,
		DataDir:                    dataDir,
	}

	emailEnabled := os.Getenv("EMAIL_ENABLED")
	if emailEnabled == "" {
		cfg.EmailEnabled = false
	} else {
		cfg.EmailEnabled = strings.ToLower(emailEnabled) == "true"
	}

	cfg.SMTPServer = os.Getenv("SMTP_SERVER")
	cfg.SMTPPort = getEnvInt("SMTP_PORT", 587)
	cfg.EmailSender = os.Getenv("EMAIL_SENDER")
	cfg.EmailPassword = os.Getenv("EMAIL_PASSWORD")

	recipients := os.Getenv("EMAIL_RECIPIENTS")
	if recipients != "" {
		cfg.EmailRecipients = strings.Split(recipients, ",")
	}

	debug := os.Getenv("DEBUG")
	cfg.Debug = strings.ToLower(debug) == "true"

	Settings = cfg
	return cfg
}

func defaultBaseDir() string {
	execBaseDir := executableBaseDir()
	if fileExists(filepath.Join(execBaseDir, "data", "zoho_tokens.json")) {
		return execBaseDir
	}

	if cwd, err := os.Getwd(); err == nil {
		if _, err := os.Stat(filepath.Join(cwd, "go.mod")); err == nil {
			return cwd
		}
		if _, err := os.Stat(filepath.Join(cwd, "backend", "go.mod")); err == nil {
			return cwd
		}
	}

	return execBaseDir
}

func executableBaseDir() string {
	execPath, _ := os.Executable()
	return filepath.Dir(filepath.Dir(execPath))
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func getEnvInt(key string, defaultVal int) int {
	val := os.Getenv(key)
	if val == "" {
		return defaultVal
	}
	n, err := strconv.Atoi(val)
	if err != nil {
		return defaultVal
	}
	return n
}

func getEnvStr(key string, defaultVal string) string {
	val := os.Getenv(key)
	if val == "" {
		return defaultVal
	}
	return val
}

func getEnvList(key string) []string {
	val := os.Getenv(key)
	if val == "" {
		return nil
	}

	items := strings.Split(val, ",")
	result := make([]string, 0, len(items))
	for _, item := range items {
		item = strings.TrimSpace(item)
		if item != "" {
			result = append(result, item)
		}
	}
	return result
}
