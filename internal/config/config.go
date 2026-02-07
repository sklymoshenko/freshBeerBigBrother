package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"
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

	MaxFileBytes         int64
	MaxDocsPerMinuteChat int
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

	maxFileBytes := int64(25 * 1024 * 1024) // 25 MiB default
	if raw := strings.TrimSpace(os.Getenv("MAX_FILE_BYTES")); raw != "" {
		n, err := strconv.ParseInt(raw, 10, 64)
		if err != nil || n <= 0 {
			return Config{}, fmt.Errorf("invalid MAX_FILE_BYTES: %s", raw)
		}
		maxFileBytes = n
	}

	maxDocsPerMinuteChat := 6
	if raw := strings.TrimSpace(os.Getenv("MAX_DOCS_PER_MINUTE_CHAT")); raw != "" {
		n, err := strconv.Atoi(raw)
		if err != nil || n <= 0 {
			return Config{}, fmt.Errorf("invalid MAX_DOCS_PER_MINUTE_CHAT: %s", raw)
		}
		maxDocsPerMinuteChat = n
	}

	return Config{
		Token:                token,
		DataDir:              dataDir,
		Mode:                 mode,
		PublicURL:            publicURL,
		MaxFileBytes:         maxFileBytes,
		MaxDocsPerMinuteChat: maxDocsPerMinuteChat,
	}, nil
}
