package helper

import "time"

// EntryRow is one parsed spreadsheet line. Defined in helper because the
// parser produces it; the tax package consumes it via insertInvoiceEntries.
type EntryRow struct {
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
