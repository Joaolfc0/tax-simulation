package tax

import (
	"context"
	"errors"
	"time"

	"encore.dev/beta/errs"
	"encore.dev/storage/sqldb"
)

type Company struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
}

type CreateCompanyParams struct {
	Name string `json:"name"`
}

type ListCompaniesResponse struct {
	Companies []Company `json:"companies"`
}

func companyExists(ctx context.Context, id string) (bool, error) {
	var exists bool
	err := db.QueryRow(ctx, `SELECT EXISTS (SELECT 1 FROM companies WHERE id = $1)`, id).Scan(&exists)
	return exists, err
}

func insertCompany(ctx context.Context, name string) (*Company, error) {
	var c Company
	err := db.QueryRow(ctx, `
		INSERT INTO companies (name)
		VALUES ($1)
		RETURNING id, name, created_at
	`, name).Scan(&c.ID, &c.Name, &c.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &c, nil
}

func selectAllCompanies(ctx context.Context) (*ListCompaniesResponse, error) {
	rows, err := db.Query(ctx, `
		SELECT id, name, created_at
		FROM companies
		ORDER BY created_at DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := &ListCompaniesResponse{Companies: []Company{}}
	for rows.Next() {
		var c Company
		if err := rows.Scan(&c.ID, &c.Name, &c.CreatedAt); err != nil {
			return nil, err
		}
		out.Companies = append(out.Companies, c)
	}
	return out, rows.Err()
}

func selectCompanyByID(ctx context.Context, id string) (*Company, error) {
	var c Company
	err := db.QueryRow(ctx, `
		SELECT id, name, created_at
		FROM companies
		WHERE id = $1
	`, id).Scan(&c.ID, &c.Name, &c.CreatedAt)
	if errors.Is(err, sqldb.ErrNoRows) {
		return nil, &errs.Error{Code: errs.NotFound, Message: "company not found"}
	}
	if err != nil {
		return nil, err
	}
	return &c, nil
}
