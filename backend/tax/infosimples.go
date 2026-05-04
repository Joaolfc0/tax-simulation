package tax

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"encore.dev/storage/sqldb"
)

const infoSimplesURL = "https://api.infosimples.com/api/v2/consultas/receita-federal/simples"

var httpClient = &http.Client{Timeout: 30 * time.Second}

type supplierInfo struct {
	CNPJ             string
	RazaoSocial      string
	SimplesSituacao  string
	IsSimplesOptante bool
}

type infoSimplesResponse struct {
	Code        int               `json:"code"`
	CodeMessage string            `json:"code_message"`
	Errors      []json.RawMessage `json:"errors"`
	Data        []struct {
		CNPJ                    string `json:"cnpj"`
		NormalizedCNPJ          string `json:"normalizado_cnpj"`
		RazaoSocial             string `json:"razao_social"`
		SimplesNacionalSituacao string `json:"simples_nacional_situacao"`
	} `json:"data"`
}

// lookupCNPJ returns supplier info for a 14-digit CNPJ. It hits the local
// suppliers cache first and only calls InfoSimples on a miss, persisting the
// result before returning.
func lookupCNPJ(ctx context.Context, cnpj string) (*supplierInfo, error) {
	if len(cnpj) != 14 {
		return nil, fmt.Errorf("invalid cnpj length: %d", len(cnpj))
	}

	if s, err := selectSupplier(ctx, cnpj); err == nil {
		return s, nil
	} else if !errors.Is(err, sqldb.ErrNoRows) {
		return nil, err
	}

	s, raw, err := fetchInfoSimples(ctx, cnpj)
	if err != nil {
		return nil, err
	}
	if err := upsertSupplier(ctx, s, raw); err != nil {
		return nil, err
	}
	return s, nil
}

func selectSupplier(ctx context.Context, cnpj string) (*supplierInfo, error) {
	var s supplierInfo
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

func upsertSupplier(ctx context.Context, s *supplierInfo, raw []byte) error {
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

func fetchInfoSimples(ctx context.Context, cnpj string) (*supplierInfo, []byte, error) {
	token := secrets.InfoSimplesToken
	if token == "" {
		return nil, nil, errors.New("InfoSimplesToken secret is not set")
	}

	q := url.Values{}
	q.Set("token", token)
	q.Set("timeout", "600")
	q.Set("ignore_site_receipt", "1")
	q.Set("cnpj", cnpj)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, infoSimplesURL+"?"+q.Encode(), nil)
	if err != nil {
		return nil, nil, err
	}
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, nil, fmt.Errorf("infosimples request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, nil, fmt.Errorf("infosimples read body: %w", err)
	}

	var parsed infoSimplesResponse
	if err := json.Unmarshal(body, &parsed); err != nil {
		return nil, nil, fmt.Errorf("infosimples decode: %w", err)
	}
	if parsed.Code != 200 {
		return nil, nil, fmt.Errorf("infosimples cnpj=%s code=%d message=%q", cnpj, parsed.Code, parsed.CodeMessage)
	}
	if len(parsed.Data) == 0 {
		return nil, nil, fmt.Errorf("infosimples cnpj=%s returned empty data", cnpj)
	}

	d := parsed.Data[0]
	return &supplierInfo{
		CNPJ:             cnpj,
		RazaoSocial:      d.RazaoSocial,
		SimplesSituacao:  d.SimplesNacionalSituacao,
		IsSimplesOptante: isSimplesOptante(d.SimplesNacionalSituacao),
	}, body, nil
}

// isSimplesOptante interprets the textual situation. Anything starting with
// "NÃO" (case- and accent-insensitive) is considered NOT optante and triggers
// the row flag in the report.
func isSimplesOptante(situacao string) bool {
	s := strings.TrimSpace(situacao)
	if s == "" {
		return false
	}
	folded := strings.ToLower(strings.Map(foldRune, s))
	return !strings.HasPrefix(folded, "nao")
}
