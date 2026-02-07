package bot

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"bigbrother/internal/processor"
	"bigbrother/internal/storage"
)

type Handler struct {
	api     *tgbotapi.BotAPI
	dataDir string
}

func NewHandler(api *tgbotapi.BotAPI, dataDir string) *Handler {
	return &Handler{
		api:     api,
		dataDir: dataDir,
	}
}

func (h *Handler) HandleUpdate(ctx context.Context, update tgbotapi.Update) error {
	if update.Message == nil {
		return nil
	}

	msg := update.Message

	if msg.IsCommand() {
		return h.handleCommand(msg)
	}

	if msg.Document != nil {
		return h.handleDocument(ctx, msg)
	}

	return nil
}

func (h *Handler) handleCommand(msg *tgbotapi.Message) error {
	switch msg.Command() {
	case "start":
		return h.replyText(msg.Chat.ID, "Send me an .xlsx or .csv file and I will process it.")
	case "help":
		return h.replyText(msg.Chat.ID, "Upload an .xlsx or .csv document. I will download and process it.")
	default:
		return h.replyText(msg.Chat.ID, "Unknown command. Use /help.")
	}
}

func (h *Handler) handleDocument(ctx context.Context, msg *tgbotapi.Message) error {
	doc := msg.Document
	if doc == nil {
		return nil
	}

	name := strings.TrimSpace(doc.FileName)
	if name == "" {
		name = "upload.xlsx"
	}

	ext := strings.ToLower(filepath.Ext(name))
	if ext != ".xlsx" && ext != ".csv" {
		return h.replyText(msg.Chat.ID, "Please upload a .xlsx or .csv file.")
	}

	file, err := h.api.GetFile(tgbotapi.FileConfig{FileID: doc.FileID})
	if err != nil {
		return fmt.Errorf("get file: %w", err)
	}

	fileURL := file.Link(h.api.Token)
	if fileURL == "" {
		return errors.New("empty file download URL")
	}

	savedPath, err := storage.SaveIncomingFile(ctx, storage.SaveInput{
		FileURL:      fileURL,
		DataDir:      h.dataDir,
		ChatID:       msg.Chat.ID,
		OriginalName: filepath.Base(name),
	})
	if err != nil {
		_ = h.replyText(msg.Chat.ID, "Failed to download the file.")
		return fmt.Errorf("save incoming file: %w", err)
	}
	defer func() { _ = os.Remove(savedPath) }()

	report, err := processor.ProcessFile(savedPath)
	if err != nil {
		_ = h.replyText(msg.Chat.ID, "Failed to process the file.")
		return fmt.Errorf("process xlsx: %w", err)
	}

	if err := h.replyText(msg.Chat.ID, report.FormatText()); err != nil {
		return err
	}

	if snark := report.MismatchSnarkText(); snark != "" {
		_ = h.replyText(msg.Chat.ID, snark)
	}

	return nil
}

func (h *Handler) replyText(chatID int64, text string) error {
	msg := tgbotapi.NewMessage(chatID, text)
	_, err := h.api.Send(msg)
	return err
}
