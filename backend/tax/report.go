package tax

import (
	"context"
	"time"
)

type ReportEntry struct {
	ID              int64     `json:"id"`
	Date            time.Time `json:"date"`
	SupplierCNPJ    string    `json:"supplier_cnpj"`
	SupplierName    string    `json:"supplier_name"`
	Numero          string    `json:"numero"`
	Especie         string    `json:"especie"`
	CFOP            string    `json:"cfop"`
	UF              string    `json:"uf"`
	ValorContabil   float64   `json:"valor_contabil"`
	VlOutrosICMS    float64   `json:"vl_outros_icms"`
	VlIsentosICMS   float64   `json:"vl_isentos_icms"`
	Flagged         bool      `json:"flagged"`
	SimplesSituacao string    `json:"simples_situacao,omitempty"`
}

type MonthReport struct {
	YearMonth          string        `json:"year_month"`
	TotalEntries       int           `json:"total_entries"`
	TotalValorContabil float64       `json:"total_valor_contabil"`
	FlaggedCount       int           `json:"flagged_count"`
	FlaggedSum         float64       `json:"flagged_sum"`
	Entries            []ReportEntry `json:"entries"`
}

func queryMonthReport(ctx context.Context, companyID, yearMonth string, start, end time.Time) (*MonthReport, error) {
	rows, err := db.Query(ctx, `
		SELECT e.id, e.escrituracao_date, e.supplier_cnpj, e.supplier_name,
		       COALESCE(e.numero, ''), COALESCE(e.especie, ''),
		       COALESCE(e.cfop, ''), COALESCE(e.uf, ''),
		       e.valor_contabil, e.vl_outros_icms, e.vl_isentos_icms,
		       e.flagged, COALESCE(s.simples_situacao, '')
		FROM invoice_entries e
		LEFT JOIN suppliers s ON s.cnpj = e.supplier_cnpj
		WHERE e.company_id = $1
		  AND e.escrituracao_date >= $2
		  AND e.escrituracao_date < $3
		ORDER BY e.escrituracao_date, e.id
	`, companyID, start, end)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	report := &MonthReport{YearMonth: yearMonth, Entries: []ReportEntry{}}
	for rows.Next() {
		var e ReportEntry
		if err := rows.Scan(
			&e.ID, &e.Date, &e.SupplierCNPJ, &e.SupplierName,
			&e.Numero, &e.Especie, &e.CFOP, &e.UF,
			&e.ValorContabil, &e.VlOutrosICMS, &e.VlIsentosICMS,
			&e.Flagged, &e.SimplesSituacao,
		); err != nil {
			return nil, err
		}
		report.Entries = append(report.Entries, e)
		report.TotalEntries++
		report.TotalValorContabil += e.ValorContabil
		if e.Flagged {
			report.FlaggedCount++
			report.FlaggedSum += e.ValorContabil
		}
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return report, nil
}
