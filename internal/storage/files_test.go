package storage

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestSaveIncomingFile_MaxBytes(t *testing.T) {
	payload := strings.Repeat("a", 1024)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/octet-stream")
		w.Header().Set("Content-Length", "1024")
		_, _ = w.Write([]byte(payload))
	}))
	defer srv.Close()

	_, err := SaveIncomingFile(context.Background(), SaveInput{
		FileURL:      srv.URL,
		DataDir:      t.TempDir(),
		ChatID:       1,
		OriginalName: "file.csv",
		MaxBytes:     100,
	})
	if err == nil || !strings.Contains(err.Error(), "file too large") {
		t.Fatalf("expected file too large error, got: %v", err)
	}
}
