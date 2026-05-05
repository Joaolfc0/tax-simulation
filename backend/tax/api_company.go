package tax

import (
	"context"
	"strings"

	"encore.dev/beta/errs"
)

//encore:api public method=POST path=/companies
func (s *Service) CreateCompany(ctx context.Context, p *CreateCompanyParams) (*Company, error) {
	name := strings.TrimSpace(p.Name)
	if name == "" {
		return nil, &errs.Error{Code: errs.InvalidArgument, Message: "name is required"}
	}
	return insertCompany(ctx, name)
}

//encore:api public method=GET path=/companies
func (s *Service) ListCompanies(ctx context.Context) (*ListCompaniesResponse, error) {
	return selectAllCompanies(ctx)
}

//encore:api public method=GET path=/companies/:id
func (s *Service) GetCompany(ctx context.Context, id string) (*Company, error) {
	return selectCompanyByID(ctx, id)
}
