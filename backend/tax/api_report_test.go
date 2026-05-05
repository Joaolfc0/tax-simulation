package tax

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"encore.dev/beta/errs"
)

func TestGetMonthReport_AggregatesUploadedEntries(t *testing.T) {
	ctx := context.Background()
	s := newTestService(t)

	c, err := s.CreateCompany(ctx, &CreateCompanyParams{Name: "Report Test Co"})
	if err != nil {
		t.Fatalf("create company: %v", err)
	}

	req, rec := newUploadRequest(t, c.ID, sampleCSV)
	s.UploadFile(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("upload failed: %d %s", rec.Code, rec.Body.String())
	}

	report, err := s.GetMonthReport(ctx, c.ID, "2026-05")
	if err != nil {
		t.Fatalf("month report: %v", err)
	}

	if report.YearMonth != "2026-05" {
		t.Fatalf("year_month = %q", report.YearMonth)
	}
	if report.TotalEntries != 2 {
		t.Fatalf("expected 2 entries, got %d", report.TotalEntries)
	}
	const wantTotal = 1500.50 + 2500.00
	if report.TotalValorContabil != wantTotal {
		t.Fatalf("expected total %.2f, got %.2f", wantTotal, report.TotalValorContabil)
	}
	if report.FlaggedCount != 1 {
		t.Fatalf("expected 1 flagged entry, got %d", report.FlaggedCount)
	}
	if report.FlaggedSum != 2500.00 {
		t.Fatalf("expected flagged sum 2500, got %.2f", report.FlaggedSum)
	}
}

func TestGetMonthReport_InvalidYearMonth(t *testing.T) {
	ctx := context.Background()
	s := newTestService(t)

	c, err := s.CreateCompany(ctx, &CreateCompanyParams{Name: "Bad YM"})
	if err != nil {
		t.Fatalf("create: %v", err)
	}

	_, err = s.GetMonthReport(ctx, c.ID, "not-a-date")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	var e *errs.Error
	if !errors.As(err, &e) || e.Code != errs.InvalidArgument {
		t.Fatalf("expected InvalidArgument, got %v", err)
	}
}

func TestGetMonthReport_CompanyNotFound(t *testing.T) {
	s := newTestService(t)

	_, err := s.GetMonthReport(context.Background(), "00000000-0000-0000-0000-000000000000", "2026-05")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	var e *errs.Error
	if !errors.As(err, &e) || e.Code != errs.NotFound {
		t.Fatalf("expected NotFound, got %v", err)
	}
}
