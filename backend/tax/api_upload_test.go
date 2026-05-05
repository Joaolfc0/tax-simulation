package tax

import (
	"bytes"
	"context"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"
)

const sampleCSV = "Dt. Escrituração,CNPJ Forn/Cliente,Nome Forn/Cliente,Valor Contábil,UF,CFOP\n" +
	"01/05/2026,11222333000181,Acme Optante Ltda,1500.50,SP,5102\n" +
	"02/05/2026,99888777000166,NotOptante SA,2500.00,RJ,6102\n"

// fakeSupplier is a deterministic SupplierResolver. It also persists each
// supplier so foreign keys on invoice_entries.supplier_cnpj are satisfied.
type fakeSupplier struct {
	t *testing.T
}

func (f fakeSupplier) Lookup(ctx context.Context, cnpj string) (*Supplier, error) {
	f.t.Helper()
	s := &Supplier{
		CNPJ:             cnpj,
		RazaoSocial:      "Fake " + cnpj,
		SimplesSituacao:  "Sim",
		IsSimplesOptante: cnpj == "11222333000181", // first CNPJ is optante; the other is not (= flagged)
	}
	if cnpj == "99888777000166" {
		s.SimplesSituacao = "Não optante"
	}
	if err := upsertSupplier(ctx, s, []byte(`{}`)); err != nil {
		return nil, err
	}
	return s, nil
}

func fakeSupplierResolver(t *testing.T) SupplierResolver {
	t.Helper()
	return fakeSupplier{t: t}
}

func newUploadRequest(t *testing.T, companyID string, csv string) (*http.Request, *httptest.ResponseRecorder) {
	t.Helper()

	var body bytes.Buffer
	mw := multipart.NewWriter(&body)
	fw, err := mw.CreateFormFile("file", "entries.csv")
	if err != nil {
		t.Fatalf("CreateFormFile: %v", err)
	}
	if _, err := fw.Write([]byte(csv)); err != nil {
		t.Fatalf("write csv: %v", err)
	}
	if err := mw.Close(); err != nil {
		t.Fatalf("multipart close: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/companies/"+companyID+"/uploads", &body)
	req.Header.Set("Content-Type", mw.FormDataContentType())
	req.SetPathValue("id", companyID)
	return req, httptest.NewRecorder()
}

func TestUploadFile_HappyPath(t *testing.T) {
	ctx := context.Background()
	s := newTestService(t)

	c, err := s.CreateCompany(ctx, &CreateCompanyParams{Name: "Upload Test Co"})
	if err != nil {
		t.Fatalf("create company: %v", err)
	}

	req, rec := newUploadRequest(t, c.ID, sampleCSV)
	s.UploadFile(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("upload status=%d body=%s", rec.Code, rec.Body.String())
	}

	var result UploadProcessResult
	if err := json.Unmarshal(rec.Body.Bytes(), &result); err != nil {
		t.Fatalf("decode result: %v body=%s", err, rec.Body.String())
	}
	if result.Status != "completed" {
		t.Fatalf("expected status=completed, got %q", result.Status)
	}
	if result.RowsProcessed != 2 {
		t.Fatalf("expected 2 rows processed, got %d", result.RowsProcessed)
	}
	if result.UniqueCNPJs != 2 {
		t.Fatalf("expected 2 unique cnpjs, got %d", result.UniqueCNPJs)
	}
	if result.FlaggedRows != 1 {
		t.Fatalf("expected 1 flagged row, got %d", result.FlaggedRows)
	}

	list, err := s.ListUploads(ctx, c.ID)
	if err != nil {
		t.Fatalf("list uploads: %v", err)
	}
	if len(list.Uploads) != 1 {
		t.Fatalf("expected 1 upload, got %d", len(list.Uploads))
	}
	if list.Uploads[0].Status != "completed" {
		t.Fatalf("expected upload status=completed, got %q", list.Uploads[0].Status)
	}
}

func TestUploadFile_CompanyNotFound(t *testing.T) {
	s := newTestService(t)

	req, rec := newUploadRequest(t, "00000000-0000-0000-0000-000000000000", sampleCSV)
	s.UploadFile(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestUploadFile_RejectsUnsupportedExtension(t *testing.T) {
	ctx := context.Background()
	s := newTestService(t)

	c, err := s.CreateCompany(ctx, &CreateCompanyParams{Name: "Bad Ext Co"})
	if err != nil {
		t.Fatalf("create company: %v", err)
	}

	var body bytes.Buffer
	mw := multipart.NewWriter(&body)
	fw, _ := mw.CreateFormFile("file", "entries.txt")
	_, _ = fw.Write([]byte("nope"))
	mw.Close()

	req := httptest.NewRequest(http.MethodPost, "/companies/"+c.ID+"/uploads", &body)
	req.Header.Set("Content-Type", mw.FormDataContentType())
	req.SetPathValue("id", c.ID)
	rec := httptest.NewRecorder()

	s.UploadFile(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d body=%s", rec.Code, rec.Body.String())
	}
}
