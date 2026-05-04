package tax

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

type sheetRow struct {
	LineNo           int
	EscrituracaoDate time.Time
	Numero           string
	Especie          string
	CST              string
	UF               string
	SupplierCNPJ     string
	SupplierName     string
	CFOP             string
	ValorContabil    float64
	VlOutrosICMS     float64
	VlIsentosICMS    float64
}

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

// Map of normalized (lowercased, accent-stripped, non-alnum-stripped) header
// strings to the logical column they represent. Multiple aliases per key are
// allowed so we tolerate small variations in source files.
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

// requiredKeys are the headers that MUST be present in the source sheet.
var requiredKeys = []headerKey{
	hDate, hCNPJ, hNome, hValorContabil,
}

func parseSheet(filename string, data []byte) ([]sheetRow, error) {
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

func parseCSV(data []byte) ([]sheetRow, error) {
	data = bytes.TrimPrefix(data, []byte{0xEF, 0xBB, 0xBF}) // strip UTF-8 BOM
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

	var rows []sheetRow
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

func parseXLSX(data []byte) ([]sheetRow, error) {
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

	var rows []sheetRow
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

// buildRow returns nil (no error) for blank rows so the caller can skip them.
func buildRow(rec []string, idx map[headerKey]int, lineNo int) (*sheetRow, error) {
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

	if dateStr == "" && cnpjStr == "" && valorStr == "" {
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

	return &sheetRow{
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

// parseBRNumber accepts both BR-locale ("1.433,01") and plain ("1433.01")
// numeric strings. Empty input is treated as 0.
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

// parseBRDate accepts dd/mm/yyyy, dd-mm-yyyy and ISO yyyy-mm-dd.
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

// normalizeCNPJ strips non-digit characters and validates the length is 14.
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

// normalizeHeader lowercases, folds Portuguese accents and strips non-alphanumeric characters.
func normalizeHeader(s string) string {
	var b strings.Builder
	for _, r := range s {
		r = foldRune(r)
		if r >= 'A' && r <= 'Z' {
			r = r + ('a' - 'A')
		}
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			b.WriteRune(r)
		}
	}
	return b.String()
}

func foldRune(r rune) rune {
	switch r {
	case 'á', 'à', 'â', 'ã', 'ä':
		return 'a'
	case 'Á', 'À', 'Â', 'Ã', 'Ä':
		return 'A'
	case 'é', 'è', 'ê', 'ë':
		return 'e'
	case 'É', 'È', 'Ê', 'Ë':
		return 'E'
	case 'í', 'ì', 'î', 'ï':
		return 'i'
	case 'Í', 'Ì', 'Î', 'Ï':
		return 'I'
	case 'ó', 'ò', 'ô', 'õ', 'ö':
		return 'o'
	case 'Ó', 'Ò', 'Ô', 'Õ', 'Ö':
		return 'O'
	case 'ú', 'ù', 'û', 'ü':
		return 'u'
	case 'Ú', 'Ù', 'Û', 'Ü':
		return 'U'
	case 'ç':
		return 'c'
	case 'Ç':
		return 'C'
	}
	return r
}
