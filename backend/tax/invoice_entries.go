package tax

import (
	"context"
	"time"

	"tax-project/backend/tax/helper"
)

// insertInvoiceEntries does a single set-based INSERT using PostgreSQL's
// unnest(). pgx serializes Go slices to PG arrays natively, so we pass one
// slice per column instead of building per-row placeholders.
func insertInvoiceEntries(ctx context.Context, companyID, uploadID string, rows []helper.EntryRow, optante map[string]bool) error {
	n := len(rows)
	cnpjs := make([]string, n)
	names := make([]string, n)
	dates := make([]time.Time, n)
	numeros := make([]string, n)
	especies := make([]string, n)
	csts := make([]string, n)
	ufs := make([]string, n)
	cfops := make([]string, n)
	valor := make([]float64, n)
	outros := make([]float64, n)
	isentos := make([]float64, n)
	flags := make([]bool, n)

	for i, r := range rows {
		cnpjs[i] = r.SupplierCNPJ
		names[i] = r.SupplierName
		dates[i] = r.EscrituracaoDate
		numeros[i] = r.Numero
		especies[i] = r.Especie
		csts[i] = r.CST
		ufs[i] = r.UF
		cfops[i] = r.CFOP
		valor[i] = r.ValorContabil
		outros[i] = r.VlOutrosICMS
		isentos[i] = r.VlIsentosICMS
		flags[i] = !optante[r.SupplierCNPJ]
	}

	_, err := db.Exec(ctx, `
		INSERT INTO invoice_entries (
			company_id, upload_id, supplier_cnpj, supplier_name, escrituracao_date,
			numero, especie, cst, uf, cfop,
			valor_contabil, vl_outros_icms, vl_isentos_icms, flagged
		)
		SELECT $1::uuid, $2::uuid,
		       c.supplier_cnpj, c.supplier_name, c.escrituracao_date,
		       NULLIF(c.numero, ''), NULLIF(c.especie, ''), NULLIF(c.cst, ''),
		       NULLIF(c.uf, ''), NULLIF(c.cfop, ''),
		       c.valor_contabil, c.vl_outros_icms, c.vl_isentos_icms, c.flagged
		FROM unnest(
			$3::text[], $4::text[], $5::date[],
			$6::text[], $7::text[], $8::text[], $9::text[], $10::text[],
			$11::numeric[], $12::numeric[], $13::numeric[], $14::bool[]
		) AS c(supplier_cnpj, supplier_name, escrituracao_date,
		       numero, especie, cst, uf, cfop,
		       valor_contabil, vl_outros_icms, vl_isentos_icms, flagged)
	`, companyID, uploadID,
		cnpjs, names, dates,
		numeros, especies, csts, ufs, cfops,
		valor, outros, isentos, flags)
	return err
}
