package processor

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/xuri/excelize/v2"
)

const (
	headerReceipt  = "Číslo daňového dokladu"
	headerCategory = "Kategorie"
	headerProduct  = "Produkt"
	headerIssuedAt = "Datum vystavení"
	headerQuantity = "Prodané množství"
)

type Report struct {
	Receipts      []ReceiptReport
	TotalReceipts int
	MismatchCount int
}

type ReceiptReport struct {
	ReceiptNo     string
	IssuedAt      string
	BeerML        int64
	BottleByML    map[int64]int64
	BottleOrder   []int64
	BottleTotalML int64
	DiffML        int64
	Match         bool
}

type columnIndex struct {
	receipt  int
	category int
	product  int
	issuedAt int
	quantity int
}

type receiptAgg struct {
	receiptNo     string
	issuedAt      string
	beerML        int64
	bottleByML    map[int64]int64
	bottleOrder   []int64
	bottleTotalML int64
}

func ProcessFile(path string) (Report, error) {
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".xlsx":
		return ProcessXLSX(path)
	case ".csv":
		return ProcessCSV(path)
	default:
		return Report{}, fmt.Errorf("unsupported file type: %s", ext)
	}
}

func ProcessXLSX(path string) (Report, error) {
	f, err := excelize.OpenFile(path)
	if err != nil {
		return Report{}, fmt.Errorf("open file: %w", err)
	}
	defer func() { _ = f.Close() }()

	sheets := f.GetSheetList()
	if len(sheets) == 0 {
		return Report{}, fmt.Errorf("no sheets found")
	}

	rows, err := f.Rows(sheets[0])
	if err != nil {
		return Report{}, fmt.Errorf("open rows: %w", err)
	}
	defer rows.Close()

	if !rows.Next() {
		return Report{}, fmt.Errorf("empty sheet")
	}
	headerRow, err := rows.Columns()
	if err != nil {
		return Report{}, fmt.Errorf("read header: %w", err)
	}
	idx, err := mapHeaders(headerRow)
	if err != nil {
		return Report{}, err
	}

	receipts := make(map[string]*receiptAgg)
	order := make([]string, 0, 256)
	rowNum := 1
	for rows.Next() {
		rowNum++
		row, err := rows.Columns()
		if err != nil {
			return Report{}, fmt.Errorf("read row %d: %w", rowNum, err)
		}
		if err := accumulateRow(row, rowNum, idx, receipts, &order); err != nil {
			return Report{}, err
		}
	}
	if err := rows.Error(); err != nil {
		return Report{}, fmt.Errorf("rows error: %w", err)
	}

	report := buildReport(receipts, order)
	return report, nil
}

func ProcessCSV(path string) (Report, error) {
	f, err := os.Open(path)
	if err != nil {
		return Report{}, fmt.Errorf("open file: %w", err)
	}
	defer f.Close()

	delimiter, err := detectCSVDelimiter(f)
	if err != nil {
		return Report{}, fmt.Errorf("detect delimiter: %w", err)
	}

	reader := csv.NewReader(f)
	reader.Comma = delimiter
	reader.FieldsPerRecord = -1

	headerRow, err := reader.Read()
	if err == io.EOF {
		return Report{}, fmt.Errorf("empty sheet")
	}
	if err != nil {
		return Report{}, fmt.Errorf("read header: %w", err)
	}

	idx, err := mapHeaders(headerRow)
	if err != nil {
		return Report{}, err
	}

	receipts := make(map[string]*receiptAgg)
	order := make([]string, 0, 256)
	rowNum := 1
	for {
		row, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return Report{}, fmt.Errorf("read row %d: %w", rowNum+1, err)
		}
		rowNum++
		if err := accumulateRow(row, rowNum, idx, receipts, &order); err != nil {
			return Report{}, err
		}
	}

	report := buildReport(receipts, order)
	return report, nil
}

func mapHeaders(headerRow []string) (columnIndex, error) {
	idx := columnIndex{
		receipt:  -1,
		category: -1,
		product:  -1,
		issuedAt: -1,
		quantity: -1,
	}

	for i, raw := range headerRow {
		header := normalizeHeader(raw)
		switch {
		case strings.EqualFold(header, headerReceipt):
			idx.receipt = i
		case strings.EqualFold(header, headerCategory):
			idx.category = i
		case strings.EqualFold(header, headerProduct):
			idx.product = i
		case strings.EqualFold(header, headerIssuedAt):
			idx.issuedAt = i
		case strings.EqualFold(header, headerQuantity):
			idx.quantity = i
		}
	}

	var missing []string
	if idx.receipt < 0 {
		missing = append(missing, headerReceipt)
	}
	if idx.category < 0 {
		missing = append(missing, headerCategory)
	}
	if idx.product < 0 {
		missing = append(missing, headerProduct)
	}
	if idx.issuedAt < 0 {
		missing = append(missing, headerIssuedAt)
	}
	if idx.quantity < 0 {
		missing = append(missing, headerQuantity)
	}
	if len(missing) > 0 {
		return idx, fmt.Errorf("missing required columns: %s", strings.Join(missing, ", "))
	}

	return idx, nil
}

func normalizeHeader(raw string) string {
	header := strings.TrimSpace(raw)
	header = strings.TrimPrefix(header, "\uFEFF")
	return header
}

func getCell(row []string, idx int) string {
	if idx < 0 || idx >= len(row) {
		return ""
	}
	return row[idx]
}

func accumulateRow(row []string, rowNum int, idx columnIndex, receipts map[string]*receiptAgg, order *[]string) error {
	receiptNo := strings.TrimSpace(getCell(row, idx.receipt))
	if receiptNo == "" {
		return nil
	}

	agg := receipts[receiptNo]
	if agg == nil {
		agg = &receiptAgg{
			receiptNo:  receiptNo,
			bottleByML: make(map[int64]int64),
		}
		receipts[receiptNo] = agg
		*order = append(*order, receiptNo)
	}

	category := strings.TrimSpace(getCell(row, idx.category))
	product := strings.TrimSpace(getCell(row, idx.product))
	issuedAt := strings.TrimSpace(getCell(row, idx.issuedAt))
	quantity := strings.TrimSpace(getCell(row, idx.quantity))
	if agg.issuedAt == "" && issuedAt != "" {
		agg.issuedAt = issuedAt
	}

	if isPivovarCategory(category) {
		beerML, err := parseLitersToML(quantity)
		if err != nil {
			return fmt.Errorf("row %d: invalid beer quantity: %w", rowNum, err)
		}
		agg.beerML += beerML
	}

	if isPETCategory(category) && isBottleProduct(product) {
		bottleML, err := parseBottleLitersML(product)
		if err != nil {
			return fmt.Errorf("row %d: invalid bottle size: %w", rowNum, err)
		}
		count, err := parseWholeCount(quantity)
		if err != nil {
			return fmt.Errorf("row %d: invalid bottle quantity: %w", rowNum, err)
		}
		if _, ok := agg.bottleByML[bottleML]; !ok {
			agg.bottleOrder = append(agg.bottleOrder, bottleML)
		}
		agg.bottleByML[bottleML] += count
		agg.bottleTotalML += bottleML * count
	}

	return nil
}

func isPivovarCategory(category string) bool {
	return hasPrefixFold(category, "Pivovar") || hasPrefixFold(category, "Pivo na čepu")
}

func isPETCategory(category string) bool {
	return strings.EqualFold(category, "PET láhve")
}

func isBottleProduct(product string) bool {
	return hasPrefixFold(product, "Láhev")
}

func parseLitersToML(raw string) (int64, error) {
	return parseDecimalToMilli(raw)
}

func parseWholeCount(raw string) (int64, error) {
	intPart, ok, err := parseWholeNumber(raw)
	if err != nil {
		return 0, err
	}
	if !ok {
		return 0, fmt.Errorf("expected whole number, got %s", raw)
	}
	return intPart, nil
}

func parseBottleLitersML(product string) (int64, error) {
	ml, ok, err := parseLitersFromProduct(product)
	if err != nil {
		return 0, err
	}
	if !ok {
		return 0, fmt.Errorf("no liters pattern")
	}
	if ml <= 0 {
		return 0, fmt.Errorf("invalid liters value")
	}
	return ml, nil
}

func buildReport(receipts map[string]*receiptAgg, order []string) Report {
	result := Report{}
	if len(receipts) == 0 {
		return result
	}

	list := make([]ReceiptReport, 0, len(receipts))
	for _, receiptNo := range order {
		agg := receipts[receiptNo]
		if agg == nil {
			continue
		}
		if agg.beerML == 0 && agg.bottleTotalML == 0 && len(agg.bottleByML) == 0 {
			continue
		}
		diff := agg.bottleTotalML - agg.beerML
		match := diff == 0
		if !match {
			result.MismatchCount++
		}
		list = append(list, ReceiptReport{
			ReceiptNo:     agg.receiptNo,
			IssuedAt:      agg.issuedAt,
			BeerML:        agg.beerML,
			BottleByML:    agg.bottleByML,
			BottleOrder:   agg.bottleOrder,
			BottleTotalML: agg.bottleTotalML,
			DiffML:        diff,
			Match:         match,
		})
	}

	result.Receipts = list
	result.TotalReceipts = len(list)
	return result
}

func (r Report) FormatText() string {
	if len(r.Receipts) == 0 {
		return "No matching beer/PET rows found."
	}

	if r.MismatchCount == 0 {
		return fmt.Sprintf("%s\nChecked %d receipts. All beer vs bottles match.",
			randomMatchMessage(),
			r.TotalReceipts,
		)
	}

	var b strings.Builder
	b.WriteString(fmt.Sprintf("Checked %d receipts. Found %d mismatches.\n", r.TotalReceipts, r.MismatchCount))

	limit := 3900
	for _, rec := range r.Receipts {
		if rec.Match {
			continue
		}
		card := formatMismatchCard(rec)
		if b.Len()+len(card) > limit {
			b.WriteString("...truncated")
			break
		}
		b.WriteString(card)
	}

	return strings.TrimSpace(b.String())
}

func (r Report) MismatchSnarkText() string {
	if r.MismatchCount == 0 {
		return ""
	}
	return randomMismatchMessage() + "\n" + randomSnark(false)
}

func formatMismatchCard(rec ReceiptReport) string {
	timePart := "-"
	if rec.IssuedAt != "" {
		timePart = rec.IssuedAt
	}

	return fmt.Sprintf(
		"===== Receipt %s =====\nTime: %s\nTotal beer: %s\nTotal bottles: %s\nDifference: %s\nBottles: %s\n\n",
		rec.ReceiptNo,
		timePart,
		formatLiters(rec.BeerML),
		formatLiters(rec.BottleTotalML),
		formatDiff(rec.DiffML),
		formatBottleList(rec.BottleByML, rec.BottleOrder),
	)
}

func formatBottleList(byML map[int64]int64, order []int64) string {
	if len(byML) == 0 {
		return "-"
	}

	parts := make([]string, 0, len(byML))
	for _, ml := range order {
		count := byML[ml]
		parts = append(parts, fmt.Sprintf("%s x%d", formatLiters(ml), count))
	}

	return strings.Join(parts, ", ")
}

func formatLiters(ml int64) string {
	liters := float64(ml) / 1000.0
	return fmt.Sprintf("%.2fL", liters)
}

func formatDiff(diffML int64) string {
	sign := "+"
	if diffML < 0 {
		sign = "-"
		diffML = -diffML
	}
	return sign + formatLiters(diffML)
}

func parseDecimalToMilli(raw string) (int64, error) {
	var intPart int64
	var fracPart int64
	fracDigits := 0
	sawDigit := false
	seenSep := false
	roundDigit := int64(-1)

	for i := 0; i < len(raw); i++ {
		b := raw[i]
		switch {
		case b >= '0' && b <= '9':
			sawDigit = true
			if !seenSep {
				intPart = intPart*10 + int64(b-'0')
			} else if fracDigits < 3 {
				fracPart = fracPart*10 + int64(b-'0')
				fracDigits++
			} else if roundDigit == -1 {
				roundDigit = int64(b - '0')
			}
		case b == '.' || b == ',':
			if seenSep {
				return 0, fmt.Errorf("invalid number: %s", raw)
			}
			seenSep = true
		case b == ' ' || b == '\t':
			continue
		default:
			return 0, fmt.Errorf("invalid number: %s", raw)
		}
	}

	if !sawDigit {
		return 0, fmt.Errorf("empty value")
	}

	for fracDigits < 3 {
		fracPart *= 10
		fracDigits++
	}

	ml := intPart*1000 + fracPart
	if roundDigit >= 5 {
		ml++
	}
	return ml, nil
}

func parseWholeNumber(raw string) (int64, bool, error) {
	var intPart int64
	sawDigit := false
	seenSep := false
	nonZeroFraction := false

	for i := 0; i < len(raw); i++ {
		b := raw[i]
		switch {
		case b >= '0' && b <= '9':
			sawDigit = true
			if !seenSep {
				intPart = intPart*10 + int64(b-'0')
			} else if b != '0' {
				nonZeroFraction = true
			}
		case b == '.' || b == ',':
			if seenSep {
				return 0, false, fmt.Errorf("invalid number: %s", raw)
			}
			seenSep = true
		case b == ' ' || b == '\t':
			continue
		default:
			return 0, false, fmt.Errorf("invalid number: %s", raw)
		}
	}

	if !sawDigit {
		return 0, false, fmt.Errorf("empty value")
	}
	if nonZeroFraction {
		return intPart, false, nil
	}
	return intPart, true, nil
}

func parseLitersFromProduct(product string) (int64, bool, error) {
	for i := 0; i < len(product); i++ {
		b := product[i]
		if b < '0' || b > '9' {
			continue
		}
		start := i
		i++
		for i < len(product) {
			b = product[i]
			if (b >= '0' && b <= '9') || b == '.' || b == ',' {
				i++
				continue
			}
			break
		}
		numStr := product[start:i]
		j := i
		for j < len(product) && (product[j] == ' ' || product[j] == '\t') {
			j++
		}
		if j < len(product) && (product[j] == 'l' || product[j] == 'L') {
			ml, err := parseDecimalToMilli(numStr)
			return ml, true, err
		}
	}
	return 0, false, nil
}

func hasPrefixFold(s, prefix string) bool {
	if prefix == "" {
		return true
	}
	if s == "" {
		return false
	}
	i := 0
	for _, pr := range prefix {
		if i >= len(s) {
			return false
		}
		sr, size := utf8.DecodeRuneInString(s[i:])
		if sr == utf8.RuneError && size == 1 {
			return false
		}
		if !runeEqualFold(sr, pr) {
			return false
		}
		i += size
	}
	return true
}

func runeEqualFold(a, b rune) bool {
	if a == b {
		return true
	}
	for r := unicode.SimpleFold(a); r != a; r = unicode.SimpleFold(r) {
		if r == b {
			return true
		}
	}
	return false
}

func detectCSVDelimiter(f *os.File) (rune, error) {
	if _, err := f.Seek(0, io.SeekStart); err != nil {
		return ',', err
	}

	reader := bufio.NewReader(f)
	line, err := reader.ReadString('\n')
	if err != nil && err != io.EOF {
		return ',', err
	}

	if _, err := f.Seek(0, io.SeekStart); err != nil {
		return ',', err
	}

	return detectDelimiterFromLine(line), nil
}

func detectDelimiterFromLine(line string) rune {
	var comma, semi, tab int
	inQuotes := false

	for i := 0; i < len(line); i++ {
		ch := line[i]
		if ch == '"' {
			if inQuotes && i+1 < len(line) && line[i+1] == '"' {
				i++
				continue
			}
			inQuotes = !inQuotes
			continue
		}
		if inQuotes {
			continue
		}
		switch ch {
		case ',':
			comma++
		case ';':
			semi++
		case '\t':
			tab++
		}
	}

	if semi > comma && semi >= tab {
		return ';'
	}
	if tab > comma && tab > semi {
		return '\t'
	}
	return ','
}
