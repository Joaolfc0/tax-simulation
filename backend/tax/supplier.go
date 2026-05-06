package tax

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"time"

	"tax-project/backend/infosimples"

	"encore.dev/rlog"
	"encore.dev/storage/sqldb"
)

// Pause between InfoSimples calls when resolving multiple distinct CNPJs in one upload.
const infosimplesCNPJThrottle = 900 * time.Millisecond

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

	res, err := r.lookupInfoSimples(ctx, cnpj)
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

func (r *cachedSupplierResolver) lookupInfoSimples(ctx context.Context, cnpj string) (*infosimples.Result, error) {
	return r.client.Lookup(ctx, cnpj)
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

// resolveDistinctCNPJOptante builds optante[CNPJ] for all keys in uniq, spacing out
// live InfoSimples calls to reduce bursts (code 609). Cache hits do not throttle.
func resolveDistinctCNPJOptante(ctx context.Context, suppliers SupplierResolver, uploadID, companyID string, uniq map[string]struct{}) (map[string]bool, error) {
	r, ok := suppliers.(*cachedSupplierResolver)
	if !ok {
		return resolveDistinctSequential(ctx, suppliers, uniq)
	}
	cnpjs := make([]string, 0, len(uniq))
	for c := range uniq {
		cnpjs = append(cnpjs, c)
	}
	slices.Sort(cnpjs)
	optante := make(map[string]bool, len(uniq))
	remoteIx := 0
	for idx, cnpj := range cnpjs {
		cached, err := supplierCached(ctx, cnpj)
		if err != nil {
			return nil, err
		}
		var info *Supplier
		var errLu error
		if cached {
			info, errLu = suppliers.Lookup(ctx, cnpj)
		} else {
			if remoteIx > 0 {
				wait := infosimplesCNPJThrottle
				rlog.Info("infosimples throttle pause between distinct CNPJ lookups",
					"upload_id", uploadID, "company_id", companyID,
					"cnpj_index", idx+1, "cnpj_total", len(cnpjs), "pause_ms", wait.Milliseconds(),
				)
				select {
				case <-ctx.Done():
					return nil, ctx.Err()
				case <-time.After(wait):
				}
			}
			info, errLu = r.lookupAndPersist(ctx, cnpj)
			if errLu == nil {
				remoteIx++
			}
		}
		if errLu != nil {
			var apiErr *infosimples.APIError
			is609 := errors.As(errLu, &apiErr) && apiErr.IsOriginRetryExhausted()
			rlog.Warn("supplier resolve failed during upload",
				"upload_id", uploadID, "company_id", companyID,
				"cnpj", cnpj,
				"cnpj_index", idx+1,
				"cnpj_total", len(cnpjs),
				"infosimples_609", is609,
				"had_row_in_suppliers_first", cached,
				"error", errLu.Error(),
			)
			return nil, fmt.Errorf("lookup CNPJ %s: %w", cnpj, errLu)
		}
		optante[cnpj] = info.IsSimplesOptante
	}
	return optante, nil
}

func resolveDistinctSequential(ctx context.Context, suppliers SupplierResolver, uniq map[string]struct{}) (map[string]bool, error) {
	optante := make(map[string]bool, len(uniq))
	for cnpj := range uniq {
		info, err := suppliers.Lookup(ctx, cnpj)
		if err != nil {
			return nil, fmt.Errorf("lookup CNPJ %s: %w", cnpj, err)
		}
		optante[cnpj] = info.IsSimplesOptante
	}
	return optante, nil
}

func (r *cachedSupplierResolver) lookupAndPersist(ctx context.Context, cnpj string) (*Supplier, error) {
	res, err := r.lookupInfoSimples(ctx, cnpj)
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

// supplierCached reports whether we already have a row for this CNPJ in `suppliers`
// so we need not hit InfoSimples again.
func supplierCached(ctx context.Context, cnpj string) (bool, error) {
	var one int
	err := db.QueryRow(ctx, `
		SELECT 1 FROM suppliers WHERE cnpj = $1 LIMIT 1
	`, cnpj).Scan(&one)
	if errors.Is(err, sqldb.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}
