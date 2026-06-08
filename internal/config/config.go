package config

import (
	"bufio"
	"errors"
	"os"
	"path/filepath"
	"strings"
)

type Config struct {
	TelegramBotToken      string
	GoogleCredentialsPath string
	GoogleSpreadsheetID   string
}

func LoadConfig() (*Config, error) {
	if err := loadDotEnv(); err != nil {
		return nil, err
	}

	cfg := &Config{
		TelegramBotToken:      strings.TrimSpace(os.Getenv("TELEGRAM_BOT_TOKEN")),
		GoogleCredentialsPath: strings.TrimSpace(os.Getenv("GOOGLE_CREDENTIALS_PATH")),
		GoogleSpreadsheetID:   strings.TrimSpace(os.Getenv("GOOGLE_SPREADSHEET_ID")),
	}

	if cfg.TelegramBotToken == "" {
		return nil, errors.New("не задано TELEGRAM_BOT_TOKEN")
	}
	if cfg.GoogleCredentialsPath == "" {
		return nil, errors.New("не задано GOOGLE_CREDENTIALS_PATH")
	}
	if cfg.GoogleSpreadsheetID == "" {
		return nil, errors.New("не задано GOOGLE_SPREADSHEET_ID")
	}

	if !filepath.IsAbs(cfg.GoogleCredentialsPath) {
		cfg.GoogleCredentialsPath = filepath.Clean(cfg.GoogleCredentialsPath)
	}

	return cfg, nil
}

func loadDotEnv() error {
	file, err := os.Open(".env")
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])
		if key == "" {
			continue
		}

		if _, exists := os.LookupEnv(key); !exists {
			os.Setenv(key, strings.Trim(value, `"'`))
		}
	}

	return scanner.Err()
}
