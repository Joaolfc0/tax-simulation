package tax

import (
	"context"
	"errors"
	"fmt"

	"tax-project/backend/infosimples"

	"encore.dev/storage/sqldb"
)

// Supplier is the cached Simples Nacional status for a CNPJ. It is internal to
// the tax service (not exposed by any //encore:api endpoint) but accessible to
// other tax-package files and tests.
type Supplier struct {
	CNPJ             string
	RazaoSocial      string
	SimplesSituacao  string
	IsSimplesOptante bool
}

// SupplierResolver is the port the tax service uses to obtain Simples Nacional
// information. The default implementation hits Postgres first and falls back
// to InfoSimples; tests inject fakes via the Service struct.
type SupplierResolver interface {
	Lookup(ctx context.Context, cnpj string) (*Supplier, error)
}

// cachedSupplierResolver implements SupplierResolver by checking the local
// `suppliers` cache table first and only calling the InfoSimples adapter on a
// miss. New entries are persisted so subsequent lookups hit the cache.
type cachedSupplierResolver struct {
	client *infosimples.Client
}

func newSupplierResolver(token string) *cachedSupplierResolver {
	return &cachedSupplierResolver{client: infosimples.New(token)}
}

func (r *cachedSupplierResolver) Lookup(ctx context.Context, cnpj string) (*Supplier, error) {
	if len(cnpj) != 14 {
		return nil, fmt.Errorf("invalid cnpj length: %d", len(cnpj))
	}

	if s, err := selectSupplier(ctx, cnpj); err == nil {
		return s, nil
	} else if !errors.Is(err, sqldb.ErrNoRows) {
		return nil, err
	}

	res, err := r.client.Lookup(ctx, cnpj)
	if err != nil {
		return nil, err
	}
	s := &Supplier{
		CNPJ:             res.CNPJ,
		RazaoSocial:      res.RazaoSocial,
		SimplesSituacao:  res.SimplesSituacao,
		IsSimplesOptante: res.IsSimplesOptante,
	}
	if err := upsertSupplier(ctx, s, res.Raw); err != nil {
		return nil, err
	}
	return s, nil
}

func selectSupplier(ctx context.Context, cnpj string) (*Supplier, error) {
	var s Supplier
	err := db.QueryRow(ctx, `
		SELECT cnpj, COALESCE(razao_social, ''), COALESCE(simples_situacao, ''), is_simples_optante
		FROM suppliers
		WHERE cnpj = $1
	`, cnpj).Scan(&s.CNPJ, &s.RazaoSocial, &s.SimplesSituacao, &s.IsSimplesOptante)
	if err != nil {
		return nil, err
	}
	return &s, nil
}

func upsertSupplier(ctx context.Context, s *Supplier, raw []byte) error {
	_, err := db.Exec(ctx, `
		INSERT INTO suppliers (cnpj, razao_social, simples_situacao, is_simples_optante, last_checked_at, last_response)
		VALUES ($1, $2, $3, $4, NOW(), $5::jsonb)
		ON CONFLICT (cnpj) DO UPDATE SET
			razao_social       = EXCLUDED.razao_social,
			simples_situacao   = EXCLUDED.simples_situacao,
			is_simples_optante = EXCLUDED.is_simples_optante,
			last_checked_at    = EXCLUDED.last_checked_at,
			last_response      = EXCLUDED.last_response
	`, s.CNPJ, s.RazaoSocial, s.SimplesSituacao, s.IsSimplesOptante, string(raw))
	return err
}
