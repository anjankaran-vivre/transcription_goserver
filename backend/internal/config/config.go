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

		GroqAPIKey string

		ProcessDelay       int
		MaxDownloadRetries int
		RetryInterval      int
		NumWorkers         int

		PIDFile string

		EmailEnabled   bool
		SMTPServer     string
		SMTPPort       int
		EmailSender    string
		EmailPassword  string
		EmailRecipients []string

		GroqDailyLimit  int
		GroqMinuteLimit int

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

		baseDir, _ := os.Getwd()
		logsDir := filepath.Join(baseDir, "logs")
		dataDir := filepath.Join(baseDir, "data")
		tokenFile := filepath.Join(dataDir, "zoho_tokens.json")
		os.MkdirAll(logsDir, 0755)
		os.MkdirAll(dataDir, 0755)

		cfg := &Config{
			ZohoClientID:       os.Getenv("ZOHO_CLIENT_ID"),
			ZohoClientSecret:   os.Getenv("ZOHO_CLIENT_SECRET"),
			ZohoRedirectURI:    os.Getenv("ZOHO_REDIRECT_URI"),
			TokenFile:          tokenFile,
			AudioUsername:      os.Getenv("AUDIO_USERNAME"),
			AudioPassword:      os.Getenv("AUDIO_PASSWORD"),
			GroqAPIKey:         os.Getenv("GROQ_API_KEY"),
			ProcessDelay:       getEnvInt("PROCESS_DELAY", 20),
			MaxDownloadRetries: getEnvInt("MAX_DOWNLOAD_RETRIES", 2),
			RetryInterval:      getEnvInt("RETRY_INTERVAL", 10),
			NumWorkers:         getEnvInt("NUM_WORKERS", 2),
			PIDFile:            filepath.Join(dataDir, "server.pid"),
			GroqDailyLimit:     getEnvInt("GROQ_DAILY_LIMIT", 14400),
			GroqMinuteLimit:    getEnvInt("GROQ_MINUTE_LIMIT", 30),
			Host:               getEnvStr("HOST", "127.0.0.1"),
			Port:               getEnvInt("PORT", 5050),
			DBHost:             os.Getenv("DB_HOST"),
			DBServer:           os.Getenv("DB_SERVER"),
			DBName:             os.Getenv("DB_NAME"),
			DBUser:             os.Getenv("DB_USER"),
			DBPassword:         os.Getenv("DB_PASSWORD"),
			DBPort:             getEnvInt("DB_PORT", 1433),
			LogsDir:            logsDir,
			DataDir:            dataDir,
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
