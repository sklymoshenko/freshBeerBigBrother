//go:build realfile

package processor

import (
	_ "embed"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

//go:embed testData.xlsx
var realFileXLSX []byte

func TestProcessXLSX_RealFileSmoke(t *testing.T) {
	if len(realFileXLSX) == 0 {
		t.Fatal("embedded test file is empty")
	}

	path := filepath.Join(t.TempDir(), "real.xlsx")
	if err := os.WriteFile(path, realFileXLSX, 0o644); err != nil {
		t.Fatalf("write embedded file: %v", err)
	}

	report, err := ProcessXLSX(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if report.TotalReceipts == 0 || len(report.Receipts) == 0 {
		t.Fatalf("expected non-empty report")
	}
	if strings.TrimSpace(report.FormatText()) == "" {
		t.Fatalf("expected formatted output")
	}
}

func BenchmarkProcessFile_XLSX_RealFile(b *testing.B) {
	if len(realFileXLSX) == 0 {
		b.Fatal("embedded test file is empty")
	}

	path := filepath.Join(b.TempDir(), "real.xlsx")
	if err := os.WriteFile(path, realFileXLSX, 0o644); err != nil {
		b.Fatalf("write embedded file: %v", err)
	}
	b.SetBytes(int64(len(realFileXLSX)))
	b.ReportAllocs()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := ProcessFile(path); err != nil {
			b.Fatalf("process: %v", err)
		}
	}
}
