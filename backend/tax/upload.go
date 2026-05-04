package tax

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"time"

	"encore.dev/beta/errs"
	"encore.dev/rlog"
	"github.com/google/uuid"
)

const maxUploadBytes int64 = 50 << 20 // 50 MiB cap

type Upload struct {
	ID            string     `json:"id"`
	CompanyID     string     `json:"company_id"`
	Filename      string     `json:"filename"`
	ContentType   string     `json:"content_type"`
	SizeBytes     int64      `json:"size_bytes"`
	Status        string     `json:"status"`
	Error         string     `json:"error,omitempty"`
	RowsProcessed int        `json:"rows_processed"`
	UploadedAt    time.Time  `json:"uploaded_at"`
	ProcessedAt   *time.Time `json:"processed_at,omitempty"`
}

type uploadResult struct {
	UploadID      string `json:"upload_id"`
	Status        string `json:"status"`
	RowsProcessed int    `json:"rows_processed"`
	FlaggedRows   int    `json:"flagged_rows"`
	UniqueCNPJs   int    `json:"unique_cnpjs"`
}

//encore:api public raw method=POST path=/companies/:id/uploads
func UploadFile(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()
	companyID := req.PathValue("id")

	exists, err := companyExists(ctx, companyID)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	if !exists {
		writeErr(w, http.StatusNotFound, "company not found")
		return
	}

	req.Body = http.MaxBytesReader(w, req.Body, maxUploadBytes)
	if err := req.ParseMultipartForm(maxUploadBytes); err != nil {
		writeErr(w, http.StatusBadRequest, fmt.Sprintf("parse multipart: %v", err))
		return
	}
	file, header, err := req.FormFile("file")
	if err != nil {
		writeErr(w, http.StatusBadRequest, "missing 'file' multipart field")
		return
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		writeErr(w, http.StatusBadRequest, fmt.Sprintf("read upload: %v", err))
		return
	}

	uploadID := uuid.NewString()
	filename := filepath.Base(header.Filename)
	storageKey := fmt.Sprintf("%s/%s/%s", companyID, uploadID, filename)
	contentType := header.Header.Get("Content-Type")
	if contentType == "" {
		contentType = http.DetectContentType(data)
	}

	if _, err := db.Exec(ctx, `
		INSERT INTO uploads (id, company_id, filename, content_type, size_bytes, storage_key, status)
		VALUES ($1, $2, $3, $4, $5, $6, 'processing')
	`, uploadID, companyID, filename, contentType, int64(len(data)), storageKey); err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}

	if err := putBucket(ctx, storageKey, data); err != nil {
		_ = markUploadFailed(ctx, uploadID, fmt.Sprintf("upload to bucket: %v", err))
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}

	result, err := processUpload(ctx, companyID, uploadID, filename, data)
	if err != nil {
		_ = markUploadFailed(ctx, uploadID, err.Error())
		writeErr(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, result)
}

func processUpload(ctx context.Context, companyID, uploadID, filename string, data []byte) (*uploadResult, error) {
	rows, err := parseSheet(filename, data)
	if err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return &uploadResult{UploadID: uploadID, Status: "completed"}, markUploadCompleted(ctx, uploadID, 0)
	}

	uniq := make(map[string]struct{}, len(rows))
	for _, r := range rows {
		uniq[r.SupplierCNPJ] = struct{}{}
	}

	optante := make(map[string]bool, len(uniq))
	for cnpj := range uniq {
		info, err := lookupCNPJ(ctx, cnpj)
		if err != nil {
			return nil, fmt.Errorf("lookup CNPJ %s: %w", cnpj, err)
		}
		optante[cnpj] = info.IsSimplesOptante
	}

	flagged := 0
	for _, r := range rows {
		if !optante[r.SupplierCNPJ] {
			flagged++
		}
	}

	if err := insertEntries(ctx, companyID, uploadID, rows, optante); err != nil {
		return nil, fmt.Errorf("insert entries: %w", err)
	}
	if err := markUploadCompleted(ctx, uploadID, len(rows)); err != nil {
		return nil, err
	}

	rlog.Info("upload processed",
		"upload_id", uploadID, "company_id", companyID,
		"rows", len(rows), "flagged", flagged, "unique_cnpjs", len(uniq))

	return &uploadResult{
		UploadID:      uploadID,
		Status:        "completed",
		RowsProcessed: len(rows),
		FlaggedRows:   flagged,
		UniqueCNPJs:   len(uniq),
	}, nil
}

// insertEntries does a single set-based INSERT using PostgreSQL's unnest().
// pgx serializes Go slices to PG arrays natively, so we pass one slice per
// column instead of building per-row placeholders.
func insertEntries(ctx context.Context, companyID, uploadID string, rows []sheetRow, optante map[string]bool) error {
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

func putBucket(ctx context.Context, key string, data []byte) error {
	w := uploadsBucket.Upload(ctx, key)
	if _, err := w.Write(data); err != nil {
		w.Abort(err)
		return err
	}
	return w.Close()
}

func markUploadCompleted(ctx context.Context, uploadID string, rowsProcessed int) error {
	_, err := db.Exec(ctx, `
		UPDATE uploads
		SET status = 'completed', rows_processed = $2, processed_at = NOW(), error = NULL
		WHERE id = $1
	`, uploadID, rowsProcessed)
	return err
}

func markUploadFailed(ctx context.Context, uploadID, msg string) error {
	_, err := db.Exec(ctx, `
		UPDATE uploads
		SET status = 'failed', error = $2, processed_at = NOW()
		WHERE id = $1
	`, uploadID, msg)
	return err
}

type ListUploadsResponse struct {
	Uploads []Upload `json:"uploads"`
}

//encore:api public method=GET path=/companies/:id/uploads
func ListUploads(ctx context.Context, id string) (*ListUploadsResponse, error) {
	exists, err := companyExists(ctx, id)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, &errs.Error{Code: errs.NotFound, Message: "company not found"}
	}

	rows, err := db.Query(ctx, `
		SELECT id, company_id, filename, content_type, size_bytes,
		       status, COALESCE(error, ''), rows_processed, uploaded_at, processed_at
		FROM uploads
		WHERE company_id = $1
		ORDER BY uploaded_at DESC
	`, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := &ListUploadsResponse{Uploads: []Upload{}}
	for rows.Next() {
		var u Upload
		var processedAt sql.NullTime
		if err := rows.Scan(
			&u.ID, &u.CompanyID, &u.Filename, &u.ContentType, &u.SizeBytes,
			&u.Status, &u.Error, &u.RowsProcessed, &u.UploadedAt, &processedAt,
		); err != nil {
			return nil, err
		}
		if processedAt.Valid {
			t := processedAt.Time
			u.ProcessedAt = &t
		}
		out.Uploads = append(out.Uploads, u)
	}
	return out, rows.Err()
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeErr(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}
