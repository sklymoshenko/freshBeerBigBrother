package bot

import (
	"context"
	"errors"
	"fmt"
	"log"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"bigbrother/internal/config"
)

func Run(ctx context.Context, cfg config.Config) error {
	if cfg.Mode == config.ModeWebhook {
		return errors.New("webhook mode is not implemented yet; use BOT_MODE=polling")
	}

	api, err := tgbotapi.NewBotAPI(cfg.Token)
	if err != nil {
		return fmt.Errorf("create bot: %w", err)
	}

	api.Debug = false
	log.Printf("Authorized as @%s", api.Self.UserName)

	handler := NewHandler(api, cfg.DataDir)

	updateCfg := tgbotapi.NewUpdate(0)
	updateCfg.Timeout = 30
	updates := api.GetUpdatesChan(updateCfg)

	for {
		select {
		case <-ctx.Done():
			api.StopReceivingUpdates()
			return nil
		case update, ok := <-updates:
			if !ok {
				return nil
			}
			if err := handler.HandleUpdate(ctx, update); err != nil {
				log.Printf("handle update: %v", err)
			}
		}
	}
}
