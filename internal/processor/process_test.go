package processor

import (
	_ "embed"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/xuri/excelize/v2"
)

//go:embed testData_mismatch.csv
var mismatchCSV []byte

func TestProcessXLSX_EmptySheet(t *testing.T) {
	path := writeXLSX(t, nil, nil)
	_, err := ProcessXLSX(path)
	if err == nil || !strings.Contains(err.Error(), "empty sheet") {
		t.Fatalf("expected empty sheet error, got: %v", err)
	}
}

func TestProcessXLSX_MissingHeaders(t *testing.T) {
	headers := []string{
		headerReceipt,
		headerProduct,
		headerIssuedAt,
		headerQuantity,
	}
	path := writeXLSX(t, headers, [][]string{
		{"R1", "Beer", "2026-02-06 10:00:00", "1"},
	})
	_, err := ProcessXLSX(path)
	if err == nil || !strings.Contains(err.Error(), "missing required columns") {
		t.Fatalf("expected missing columns error, got: %v", err)
	}
}

func TestProcessXLSX_HeaderBOM(t *testing.T) {
	headers := []string{
		"\uFEFF" + headerReceipt,
		headerCategory,
		headerProduct,
		headerIssuedAt,
		headerQuantity,
	}
	rows := [][]string{
		{"R1", "Pivovar Test", "Beer", "2026-02-06 10:00:00", "1"},
		{"R1", "PET láhve", "Láhev 1 l", "2026-02-06 10:00:00", "1"},
	}
	path := writeXLSX(t, headers, rows)
	report, err := ProcessXLSX(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if report.TotalReceipts != 1 || report.MismatchCount != 0 {
		t.Fatalf("unexpected report: %+v", report)
	}
}

func TestProcessXLSX_MatchBeerAndBottles(t *testing.T) {
	headers := []string{
		headerCategory,
		headerReceipt,
		headerProduct,
		headerIssuedAt,
		headerQuantity,
	}
	rows := [][]string{
		{"Pivovar Premium", "R1", "Beer A", "2026-02-06 10:00:00", "2"},
		{"PET láhve", "R1", "Láhev 1 l", "2026-02-06 10:00:00", "2"},
		{"PET láhve", "R1", "Taška s uchem", "2026-02-06 10:00:00", "1"},
	}
	path := writeXLSX(t, headers, rows)
	report, err := ProcessXLSX(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if report.TotalReceipts != 1 || report.MismatchCount != 0 {
		t.Fatalf("expected 1 receipt with no mismatch, got: %+v", report)
	}
	rec := report.Receipts[0]
	if rec.BeerML != 2000 || rec.BottleTotalML != 2000 {
		t.Fatalf("unexpected totals: beer=%d, bottles=%d", rec.BeerML, rec.BottleTotalML)
	}
	if rec.BottleByML[1000] != 2 {
		t.Fatalf("expected 1.00L x2 bottles, got: %+v", rec.BottleByML)
	}
}

func TestProcessXLSX_MismatchDiff(t *testing.T) {
	headers := []string{
		headerReceipt,
		headerCategory,
		headerProduct,
		headerIssuedAt,
		headerQuantity,
	}
	rows := [][]string{
		{"R1", "Pivovar Test", "Beer", "2026-02-06 10:00:00", "1"},
		{"R1", "PET láhve", "Láhev 0,5 l", "2026-02-06 10:00:00", "1"},
	}
	path := writeXLSX(t, headers, rows)
	report, err := ProcessXLSX(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if report.MismatchCount != 1 {
		t.Fatalf("expected mismatch count 1, got: %d", report.MismatchCount)
	}
	rec := report.Receipts[0]
	if rec.DiffML != -500 {
		t.Fatalf("expected diff -500ml, got: %d", rec.DiffML)
	}
}

func TestProcessXLSX_BottleQuantityMustBeWhole(t *testing.T) {
	headers := []string{
		headerReceipt,
		headerCategory,
		headerProduct,
		headerIssuedAt,
		headerQuantity,
	}
	rows := [][]string{
		{"R1", "PET láhve", "Láhev 1 l", "2026-02-06 10:00:00", "1,5"},
	}
	path := writeXLSX(t, headers, rows)
	_, err := ProcessXLSX(path)
	if err == nil || !strings.Contains(err.Error(), "expected whole number") {
		t.Fatalf("expected whole number error, got: %v", err)
	}
}

func TestProcessXLSX_ParseCommaLiters(t *testing.T) {
	headers := []string{
		headerReceipt,
		headerCategory,
		headerProduct,
		headerIssuedAt,
		headerQuantity,
	}
	rows := [][]string{
		{"R1", "Pivovar Test", "Beer", "2026-02-06 10:00:00", "1,0"},
		{"R1", "PET láhve", "Láhev 0,5 l", "2026-02-06 10:00:00", "2"},
	}
	path := writeXLSX(t, headers, rows)
	report, err := ProcessXLSX(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	rec := report.Receipts[0]
	if rec.BeerML != 1000 || rec.BottleTotalML != 1000 {
		t.Fatalf("unexpected totals: beer=%d, bottles=%d", rec.BeerML, rec.BottleTotalML)
	}
}

func TestProcessXLSX_PivoNaCepuCategory(t *testing.T) {
	headers := []string{
		headerReceipt,
		headerCategory,
		headerProduct,
		headerIssuedAt,
		headerQuantity,
	}
	rows := [][]string{
		{"R1", "Pivo na čepu světlé", "Beer", "2026-02-06 10:00:00", "1,5"},
		{"R1", "PET láhve", "Láhev 0,5 l", "2026-02-06 10:00:00", "3"},
	}
	path := writeXLSX(t, headers, rows)
	report, err := ProcessXLSX(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if report.MismatchCount != 0 {
		t.Fatalf("expected match, got: %+v", report)
	}
	rec := report.Receipts[0]
	if rec.BeerML != 1500 || rec.BottleTotalML != 1500 {
		t.Fatalf("unexpected totals: beer=%d, bottles=%d", rec.BeerML, rec.BottleTotalML)
	}
}

func TestProcessXLSX_BottleProductAccentStrict(t *testing.T) {
	headers := []string{
		headerReceipt,
		headerCategory,
		headerProduct,
		headerIssuedAt,
		headerQuantity,
	}
	rows := [][]string{
		{"R1", "Pivovar Test", "Beer", "2026-02-06 10:00:00", "1"},
		{"R1", "PET láhve", "Lahev 1 l", "2026-02-06 10:00:00", "1"},
	}
	path := writeXLSX(t, headers, rows)
	report, err := ProcessXLSX(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	rec := report.Receipts[0]
	if rec.BottleTotalML != 0 {
		t.Fatalf("expected no bottles counted, got: %d", rec.BottleTotalML)
	}
	if rec.Match {
		t.Fatalf("expected mismatch when bottles are ignored")
	}
}

func TestProcessFile_Unsupported(t *testing.T) {
	path := filepath.Join(t.TempDir(), "test.txt")
	if err := os.WriteFile(path, []byte("x"), 0o644); err != nil {
		t.Fatalf("write file: %v", err)
	}
	_, err := ProcessFile(path)
	if err == nil || !strings.Contains(err.Error(), "unsupported file type") {
		t.Fatalf("expected unsupported file type error, got: %v", err)
	}
}

func TestProcessCSV_MismatchFile(t *testing.T) {
	if len(mismatchCSV) == 0 {
		t.Fatal("embedded mismatch csv is empty")
	}

	path := filepath.Join(t.TempDir(), "testData_mismatch.csv")
	if err := os.WriteFile(path, mismatchCSV, 0o644); err != nil {
		t.Fatalf("write mismatch csv: %v", err)
	}

	report, err := ProcessCSV(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := map[string]struct {
		beerML   int64
		bottleML int64
		diffML   int64
		match    bool
	}{
		"3001": {beerML: 2000, bottleML: 1500, diffML: -500, match: false},
		"3002": {beerML: 1000, bottleML: 2000, diffML: 1000, match: false},
		"3003": {beerML: 1500, bottleML: 1000, diffML: -500, match: false},
		"3004": {beerML: 2000, bottleML: 2000, diffML: 0, match: true},
		"3005": {beerML: 1000, bottleML: 1000, diffML: 0, match: true},
		"3006": {beerML: 0, bottleML: 1000, diffML: 1000, match: false},
		"3007": {beerML: 500, bottleML: 0, diffML: -500, match: false},
		"3008": {beerML: 2500, bottleML: 2500, diffML: 0, match: true},
		"3009": {beerML: 1000, bottleML: 0, diffML: -1000, match: false},
		"3010": {beerML: 3000, bottleML: 2500, diffML: -500, match: false},
	}

	if report.TotalReceipts != len(expected) {
		t.Fatalf("expected %d receipts, got: %d", len(expected), report.TotalReceipts)
	}

	mismatchCount := 0
	for _, rec := range report.Receipts {
		exp, ok := expected[rec.ReceiptNo]
		if !ok {
			t.Fatalf("unexpected receipt in report: %s", rec.ReceiptNo)
		}
		if rec.BeerML != exp.beerML || rec.BottleTotalML != exp.bottleML || rec.DiffML != exp.diffML {
			t.Fatalf("receipt %s: expected beer=%d bottles=%d diff=%d, got beer=%d bottles=%d diff=%d",
				rec.ReceiptNo, exp.beerML, exp.bottleML, exp.diffML, rec.BeerML, rec.BottleTotalML, rec.DiffML)
		}
		if rec.Match != exp.match {
			t.Fatalf("receipt %s: expected match=%v, got %v", rec.ReceiptNo, exp.match, rec.Match)
		}
		if !rec.Match {
			mismatchCount++
		}
	}
	if mismatchCount != 7 {
		t.Fatalf("expected 7 mismatches, got: %d", mismatchCount)
	}
}

func BenchmarkProcessFile_XLSX_Synthetic(b *testing.B) {
	const receiptCount = 2000
	const issuedAt = "2026-02-06 10:00:00"

	headers := []string{
		headerReceipt,
		headerCategory,
		headerProduct,
		headerIssuedAt,
		headerQuantity,
	}

	rows := make([][]string, 0, receiptCount*3)
	for i := 0; i < receiptCount; i++ {
		receipt := fmt.Sprintf("R%06d", i)
		rows = append(rows, []string{receipt, "Pivovar Premium", "Beer", issuedAt, "2"})
		rows = append(rows, []string{receipt, "PET láhve", "Láhev 1 l", issuedAt, "2"})
		rows = append(rows, []string{receipt, "PET láhve", "Taška s uchem", issuedAt, "1"})
	}

	path := writeXLSX(b, headers, rows)
	if info, err := os.Stat(path); err == nil {
		b.SetBytes(info.Size())
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := ProcessFile(path); err != nil {
			b.Fatalf("process: %v", err)
		}
	}
}

func BenchmarkProcessFile_CSV_RealFile(b *testing.B) {
	path := filepath.Join("testData.csv")
	info, err := os.Stat(path)
	if err != nil {
		b.Fatalf("expected test file at %s", path)
	}
	b.SetBytes(info.Size())
	b.ReportAllocs()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := ProcessFile(path); err != nil {
			b.Fatalf("process: %v", err)
		}
	}
}

func writeXLSX(tb testing.TB, headers []string, rows [][]string) string {
	tb.Helper()

	f := excelize.NewFile()
	sheet := f.GetSheetName(f.GetActiveSheetIndex())

	if len(headers) > 0 {
		for col, header := range headers {
			cell, _ := excelize.CoordinatesToCellName(col+1, 1)
			if err := f.SetCellValue(sheet, cell, header); err != nil {
				tb.Fatalf("set header cell: %v", err)
			}
		}
	}

	for r, row := range rows {
		for c, val := range row {
			cell, _ := excelize.CoordinatesToCellName(c+1, r+2)
			if err := f.SetCellValue(sheet, cell, val); err != nil {
				tb.Fatalf("set row cell: %v", err)
			}
		}
	}

	path := filepath.Join(tb.TempDir(), "test.xlsx")
	if err := f.SaveAs(path); err != nil {
		tb.Fatalf("save xlsx: %v", err)
	}
	if err := f.Close(); err != nil {
		tb.Fatalf("close xlsx: %v", err)
	}
	return path
}
