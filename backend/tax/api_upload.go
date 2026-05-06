package tax

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"path/filepath"

	"tax-project/backend/tax/helper"

	"encore.dev"
	"encore.dev/beta/errs"
	"github.com/google/uuid"
)

//encore:api public raw method=POST path=/companies/:id/uploads
func (s *Service) UploadFile(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()
	companyID := encore.CurrentRequest().PathParams.Get("id")
	if companyID == "" {
		companyID = req.PathValue("id")
	}
	if _, err := uuid.Parse(companyID); err != nil {
		helper.WriteErr(w, http.StatusBadRequest, "invalid company id")
		return
	}

	exists, err := companyExists(ctx, companyID)
	if err != nil {
		helper.WriteErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	if !exists {
		helper.WriteErr(w, http.StatusNotFound, "company not found")
		return
	}

	req.Body = http.MaxBytesReader(w, req.Body, helper.MaxUploadBytes)
	if err := req.ParseMultipartForm(helper.MaxUploadBytes); err != nil {
		helper.WriteErr(w, http.StatusBadRequest, fmt.Sprintf("parse multipart: %v", err))
		return
	}
	file, header, err := req.FormFile("file")
	if err != nil {
		helper.WriteErr(w, http.StatusBadRequest, "missing 'file' multipart field")
		return
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		helper.WriteErr(w, http.StatusBadRequest, fmt.Sprintf("read upload: %v", err))
		return
	}

	uploadID := uuid.NewString()
	filename := filepath.Base(header.Filename)
	storageKey := fmt.Sprintf("%s/%s/%s", companyID, uploadID, filename)
	contentType := header.Header.Get("Content-Type")
	if contentType == "" {
		contentType = http.DetectContentType(data)
	}

	if err := insertUploadRow(ctx, uploadID, companyID, filename, contentType, storageKey, int64(len(data))); err != nil {
		helper.WriteErr(w, http.StatusInternalServerError, err.Error())
		return
	}

	if err := helper.PutUploadObject(ctx, storageKey, data); err != nil {
		_ = markUploadFailed(ctx, uploadID, fmt.Sprintf("upload to bucket: %v", err))
		helper.WriteErr(w, http.StatusInternalServerError, err.Error())
		return
	}

	result, err := processUpload(ctx, s.Suppliers, companyID, uploadID, filename, data)
	if err != nil {
		_ = markUploadFailed(ctx, uploadID, err.Error())
		helper.WriteErr(w, http.StatusBadRequest, err.Error())
		return
	}

	helper.WriteJSON(w, http.StatusOK, result)
}

//encore:api public method=GET path=/companies/:id/uploads
func (s *Service) ListUploads(ctx context.Context, id string) (*ListUploadsResponse, error) {
	exists, err := companyExists(ctx, id)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, &errs.Error{Code: errs.NotFound, Message: "company not found"}
	}

	return selectUploadsByCompany(ctx, id)
}
