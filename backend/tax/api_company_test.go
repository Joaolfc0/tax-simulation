package tax

import (
	"context"
	"errors"
	"testing"

	"encore.dev/beta/errs"
)

func newTestService(t *testing.T) *Service {
	t.Helper()
	return &Service{Suppliers: fakeSupplierResolver(t)}
}

func TestCreateCompany_RequiresName(t *testing.T) {
	s := newTestService(t)

	_, err := s.CreateCompany(context.Background(), &CreateCompanyParams{Name: "  "})
	if err == nil {
		t.Fatal("expected error for empty name, got nil")
	}
	var e *errs.Error
	if !errors.As(err, &e) || e.Code != errs.InvalidArgument {
		t.Fatalf("expected InvalidArgument, got %v", err)
	}
}

func TestCompanyCRUD(t *testing.T) {
	ctx := context.Background()
	s := newTestService(t)

	created, err := s.CreateCompany(ctx, &CreateCompanyParams{Name: "Acme Tax Co"})
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if created.ID == "" || created.Name != "Acme Tax Co" {
		t.Fatalf("unexpected created company: %+v", created)
	}

	got, err := s.GetCompany(ctx, created.ID)
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if got.ID != created.ID {
		t.Fatalf("get returned wrong id: got %s want %s", got.ID, created.ID)
	}

	list, err := s.ListCompanies(ctx)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	found := false
	for _, c := range list.Companies {
		if c.ID == created.ID {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("created company not present in list")
	}
}

func TestGetCompany_NotFound(t *testing.T) {
	s := newTestService(t)

	_, err := s.GetCompany(context.Background(), "00000000-0000-0000-0000-000000000000")
	if err == nil {
		t.Fatal("expected NotFound error, got nil")
	}
	var e *errs.Error
	if !errors.As(err, &e) || e.Code != errs.NotFound {
		t.Fatalf("expected NotFound, got %v", err)
	}
}
