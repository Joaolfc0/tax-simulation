package tax

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"tax-project/backend/tax/helper"

	"encore.dev/rlog"
)

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

type UploadProcessResult struct {
	UploadID      string `json:"upload_id"`
	Status        string `json:"status"`
	RowsProcessed int    `json:"rows_processed"`
	FlaggedRows   int    `json:"flagged_rows"`
	UniqueCNPJs   int    `json:"unique_cnpjs"`
}

type ListUploadsResponse struct {
	Uploads []Upload `json:"uploads"`
}

func insertUploadRow(ctx context.Context, uploadID, companyID, filename, contentType, storageKey string, sizeBytes int64) error {
	_, err := db.Exec(ctx, `
		INSERT INTO uploads (id, company_id, filename, content_type, size_bytes, storage_key, status)
		VALUES ($1, $2, $3, $4, $5, $6, 'processing')
	`, uploadID, companyID, filename, contentType, sizeBytes, storageKey)
	return err
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

func selectUploadsByCompany(ctx context.Context, companyID string) (*ListUploadsResponse, error) {
	rows, err := db.Query(ctx, `
		SELECT id, company_id, filename, content_type, size_bytes,
		       status, COALESCE(error, ''), rows_processed, uploaded_at, processed_at
		FROM uploads
		WHERE company_id = $1
		ORDER BY uploaded_at DESC
	`, companyID)
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

// processUpload parses spreadsheet bytes, resolves Simples Nacional status
// per supplier via the injected resolver, persists invoice rows, and marks
// the upload row completed. Errors are returned to the caller, which decides
// whether to mark the upload failed.
func processUpload(
	ctx context.Context,
	suppliers SupplierResolver,
	companyID, uploadID, filename string,
	data []byte,
) (*UploadProcessResult, error) {
	rows, err := helper.ParseSpreadsheet(filename, data)
	if err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return &UploadProcessResult{UploadID: uploadID, Status: "completed"}, markUploadCompleted(ctx, uploadID, 0)
	}

	uniq := make(map[string]struct{}, len(rows))
	for _, r := range rows {
		uniq[r.SupplierCNPJ] = struct{}{}
	}

	optante, err := resolveDistinctCNPJOptante(ctx, suppliers, uploadID, companyID, uniq)
	if err != nil {
		return nil, err
	}

	flagged := 0
	for _, r := range rows {
		if !optante[r.SupplierCNPJ] {
			flagged++
		}
	}

	if err := insertInvoiceEntries(ctx, companyID, uploadID, rows, optante); err != nil {
		return nil, fmt.Errorf("insert entries: %w", err)
	}
	if err := markUploadCompleted(ctx, uploadID, len(rows)); err != nil {
		return nil, err
	}

	rlog.Info("upload processed",
		"upload_id", uploadID, "company_id", companyID,
		"rows", len(rows), "flagged", flagged, "unique_cnpjs", len(uniq))

	return &UploadProcessResult{
		UploadID:      uploadID,
		Status:        "completed",
		RowsProcessed: len(rows),
		FlaggedRows:   flagged,
		UniqueCNPJs:   len(uniq),
	}, nil
}
