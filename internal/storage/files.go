package storage

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type SaveInput struct {
	FileURL      string
	DataDir      string
	ChatID       int64
	OriginalName string
	MaxBytes     int64
}

func SaveIncomingFile(ctx context.Context, in SaveInput) (string, error) {
	if in.FileURL == "" {
		return "", fmt.Errorf("file URL is empty")
	}
	if in.DataDir == "" {
		return "", fmt.Errorf("data dir is empty")
	}

	dir := filepath.Join(in.DataDir, "incoming")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", fmt.Errorf("mkdir: %w", err)
	}

	safeName := sanitizeFilename(in.OriginalName)
	ext := strings.ToLower(filepath.Ext(safeName))
	if ext != ".xlsx" && ext != ".csv" {
		safeName = safeName + ".xlsx"
	}

	timestamp := time.Now().UTC().Format("20060102T150405Z")
	filename := fmt.Sprintf("%s_%d_%s", timestamp, in.ChatID, safeName)
	dstPath := filepath.Join(dir, filename)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, in.FileURL, nil)
	if err != nil {
		return "", fmt.Errorf("create request: %w", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("download file: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("download failed: status %s", resp.Status)
	}

	if in.MaxBytes > 0 && resp.ContentLength > 0 && resp.ContentLength > in.MaxBytes {
		return "", fmt.Errorf("file too large: %d bytes (max %d)", resp.ContentLength, in.MaxBytes)
	}

	out, err := os.Create(dstPath)
	if err != nil {
		return "", fmt.Errorf("create file: %w", err)
	}
	defer out.Close()

	reader := io.Reader(resp.Body)
	if in.MaxBytes > 0 {
		reader = io.LimitReader(resp.Body, in.MaxBytes+1)
	}

	written, err := io.Copy(out, reader)
	if err != nil {
		_ = os.Remove(dstPath)
		return "", fmt.Errorf("write file: %w", err)
	}
	if in.MaxBytes > 0 && written > in.MaxBytes {
		_ = os.Remove(dstPath)
		return "", fmt.Errorf("file too large: wrote %d bytes (max %d)", written, in.MaxBytes)
	}

	return dstPath, nil
}

func sanitizeFilename(name string) string {
	name = strings.TrimSpace(name)
	if name == "" {
		return "upload.xlsx"
	}

	var b strings.Builder
	b.Grow(len(name))
	for _, r := range name {
		switch {
		case r >= 'a' && r <= 'z':
			b.WriteRune(r)
		case r >= 'A' && r <= 'Z':
			b.WriteRune(r)
		case r >= '0' && r <= '9':
			b.WriteRune(r)
		case r == '.' || r == '-' || r == '_':
			b.WriteRune(r)
		default:
			b.WriteRune('_')
		}
	}

	cleaned := b.String()
	if cleaned == "" {
		return "upload.xlsx"
	}

	return cleaned
}
