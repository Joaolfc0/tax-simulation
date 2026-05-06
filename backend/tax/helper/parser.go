package helper

import (
	"bytes"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/xuri/excelize/v2"
)

type headerKey int

const (
	hDate headerKey = iota
	hNumero
	hEspecie
	hCST
	hUF
	hCNPJ
	hNome
	hCFOP
	hValorContabil
	hOutrosICMS
	hIsentosICMS
)

var headerAliases = map[string]headerKey{
	"dtescrituracao":     hDate,
	"dataescrituracao":   hDate,
	"numero":             hNumero,
	"especie":            hEspecie,
	"cst":                hCST,
	"csticms":            hCST,
	"ufforncliente":      hUF,
	"uf":                 hUF,
	"cnpjcpfforncliente": hCNPJ,
	"cnpjforncliente":    hCNPJ,
	"cnpjcpf":            hCNPJ,
	"cnpj":               hCNPJ,
	"nomeforncliente":    hNome,
	"nome":               hNome,
	"cfop":               hCFOP,
	"valorcontabil":      hValorContabil,
	"vloutrosicms":       hOutrosICMS,
	"valoroutrosicms":    hOutrosICMS,
	"vlisentosicms":      hIsentosICMS,
	"valorisentosicms":   hIsentosICMS,
}

var requiredKeys = []headerKey{
	hDate, hCNPJ, hNome, hValorContabil,
}

// ParseSpreadsheet parses CSV or XLSX bytes into entry rows.
func ParseSpreadsheet(filename string, data []byte) ([]EntryRow, error) {
	lower := strings.ToLower(filename)
	switch {
	case strings.HasSuffix(lower, ".xlsx"):
		return parseXLSX(data)
	case strings.HasSuffix(lower, ".csv"):
		return parseCSV(data)
	default:
		return nil, fmt.Errorf("unsupported file extension (only .xlsx and .csv are accepted): %s", filename)
	}
}

func parseCSV(data []byte) ([]EntryRow, error) {
	data = bytes.TrimPrefix(data, []byte{0xEF, 0xBB, 0xBF})
	r := csv.NewReader(bytes.NewReader(data))
	r.FieldsPerRecord = -1
	r.LazyQuotes = true

	header, err := r.Read()
	if err != nil {
		return nil, fmt.Errorf("read csv header: %w", err)
	}
	idx, err := buildHeaderIndex(header)
	if err != nil {
		return nil, err
	}

	var rows []EntryRow
	lineNo := 1
	for {
		rec, err := r.Read()
		if errors.Is(err, io.EOF) {
			break
		}
		lineNo++
		if err != nil {
			return nil, fmt.Errorf("csv line %d: %w", lineNo, err)
		}
		row, err := buildRow(rec, idx, lineNo)
		if err != nil {
			return nil, err
		}
		if row == nil {
			continue
		}
		rows = append(rows, *row)
	}
	return rows, nil
}

func parseXLSX(data []byte) ([]EntryRow, error) {
	f, err := excelize.OpenReader(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("open xlsx: %w", err)
	}
	defer f.Close()

	sheets := f.GetSheetList()
	if len(sheets) == 0 {
		return nil, errors.New("xlsx contains no sheets")
	}
	all, err := f.GetRows(sheets[0])
	if err != nil {
		return nil, fmt.Errorf("read xlsx rows: %w", err)
	}
	if len(all) == 0 {
		return nil, errors.New("xlsx is empty")
	}
	idx, err := buildHeaderIndex(all[0])
	if err != nil {
		return nil, err
	}

	var rows []EntryRow
	for i, rec := range all[1:] {
		lineNo := i + 2
		row, err := buildRow(rec, idx, lineNo)
		if err != nil {
			return nil, err
		}
		if row == nil {
			continue
		}
		rows = append(rows, *row)
	}
	return rows, nil
}

func buildHeaderIndex(header []string) (map[headerKey]int, error) {
	idx := make(map[headerKey]int)
	for i, h := range header {
		key, ok := headerAliases[normalizeHeader(h)]
		if !ok {
			continue
		}
		if _, dup := idx[key]; dup {
			continue
		}
		idx[key] = i
	}
	for _, k := range requiredKeys {
		if _, ok := idx[k]; !ok {
			return nil, fmt.Errorf("missing required column: %s", headerKeyLabel(k))
		}
	}
	return idx, nil
}

func headerKeyLabel(k headerKey) string {
	switch k {
	case hDate:
		return "Dt. Escrituração"
	case hNumero:
		return "Número"
	case hEspecie:
		return "Espécie"
	case hCST:
		return "CST ICMS"
	case hUF:
		return "UF Forn/Cliente"
	case hCNPJ:
		return "CNPJ/CPF Forn/Cliente"
	case hNome:
		return "Nome Forn/Cliente"
	case hCFOP:
		return "CFOP"
	case hValorContabil:
		return "Valor Contábil"
	case hOutrosICMS:
		return "Vl. Outros ICMS"
	case hIsentosICMS:
		return "Vl. Isentos ICMS"
	}
	return "?"
}

func buildRow(rec []string, idx map[headerKey]int, lineNo int) (*EntryRow, error) {
	get := func(k headerKey) string {
		i, ok := idx[k]
		if !ok || i >= len(rec) {
			return ""
		}
		return strings.TrimSpace(rec[i])
	}

	dateStr := get(hDate)
	cnpjStr := get(hCNPJ)
	nameStr := get(hNome)
	valorStr := get(hValorContabil)

	// Ignore footer/summary rows (e.g. "42 Notas" totals) that do not
	// represent a real invoice entry. They typically carry aggregate numeric
	// values but no date/CNPJ/name.
	if dateStr == "" && cnpjStr == "" && nameStr == "" {
		return nil, nil
	}

	date, err := parseBRDate(dateStr)
	if err != nil {
		return nil, fmt.Errorf("line %d: invalid Dt. Escrituração %q: %w", lineNo, dateStr, err)
	}
	cnpj, err := normalizeCNPJ(cnpjStr)
	if err != nil {
		return nil, fmt.Errorf("line %d: invalid CNPJ %q: %w", lineNo, cnpjStr, err)
	}
	valor, err := parseBRNumber(valorStr)
	if err != nil {
		return nil, fmt.Errorf("line %d: invalid Valor Contábil %q: %w", lineNo, valorStr, err)
	}
	outros, err := parseBRNumber(get(hOutrosICMS))
	if err != nil {
		return nil, fmt.Errorf("line %d: invalid Vl. Outros ICMS: %w", lineNo, err)
	}
	isentos, err := parseBRNumber(get(hIsentosICMS))
	if err != nil {
		return nil, fmt.Errorf("line %d: invalid Vl. Isentos ICMS: %w", lineNo, err)
	}

	return &EntryRow{
		LineNo:           lineNo,
		EscrituracaoDate: date,
		Numero:           get(hNumero),
		Especie:          get(hEspecie),
		CST:              get(hCST),
		UF:               get(hUF),
		SupplierCNPJ:     cnpj,
		SupplierName:     nameStr,
		CFOP:             get(hCFOP),
		ValorContabil:    valor,
		VlOutrosICMS:     outros,
		VlIsentosICMS:    isentos,
	}, nil
}

func parseBRNumber(s string) (float64, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0, nil
	}
	if strings.Contains(s, ",") {
		s = strings.ReplaceAll(s, ".", "")
		s = strings.ReplaceAll(s, ",", ".")
	}
	return strconv.ParseFloat(s, 64)
}

func parseBRDate(s string) (time.Time, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return time.Time{}, errors.New("empty date")
	}
	for _, layout := range []string{"02/01/2006", "2/1/2006", "02-01-2006", "2006-01-02"} {
		if t, err := time.Parse(layout, s); err == nil {
			return t, nil
		}
	}
	return time.Time{}, fmt.Errorf("unrecognized date format: %s", s)
}

func normalizeCNPJ(s string) (string, error) {
	var b strings.Builder
	for _, r := range s {
		if r >= '0' && r <= '9' {
			b.WriteRune(r)
		}
	}
	digits := b.String()
	if len(digits) != 14 {
		return "", fmt.Errorf("CNPJ must have 14 digits, got %d", len(digits))
	}
	return digits, nil
}
