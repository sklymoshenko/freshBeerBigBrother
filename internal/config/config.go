package config

import (
	"errors"
	"fmt"
	"os"
	"strings"
)

type Mode string

const (
	ModePolling Mode = "polling"
	ModeWebhook Mode = "webhook"
)

type Config struct {
	Token     string
	DataDir   string
	Mode      Mode
	PublicURL string
}

func Load() (Config, error) {
	if err := loadEnvFiles(); err != nil {
		return Config{}, err
	}

	token := strings.TrimSpace(os.Getenv("TELEGRAM_BOT_TOKEN"))
	if token == "" {
		return Config{}, errors.New("TELEGRAM_BOT_TOKEN is required")
	}

	dataDir := strings.TrimSpace(os.Getenv("DATA_DIR"))
	if dataDir == "" {
		dataDir = "./data"
	}

	modeRaw := strings.TrimSpace(os.Getenv("BOT_MODE"))
	if modeRaw == "" {
		modeRaw = string(ModePolling)
	}

	mode := Mode(strings.ToLower(modeRaw))
	switch mode {
	case ModePolling, ModeWebhook:
		// ok
	default:
		return Config{}, fmt.Errorf("unsupported BOT_MODE: %s", modeRaw)
	}

	publicURL := strings.TrimSpace(os.Getenv("BOT_PUBLIC_URL"))
	if mode == ModeWebhook && publicURL == "" {
		return Config{}, errors.New("BOT_PUBLIC_URL is required for webhook mode")
	}

	return Config{
		Token:     token,
		DataDir:   dataDir,
		Mode:      mode,
		PublicURL: publicURL,
	}, nil
}
