// Package infosimples is a self-contained HTTP client for the InfoSimples
// "Receita Federal Simples Nacional" lookup endpoint. It has no Encore
// dependencies so it can be unit-tested in isolation.
package infosimples

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
)

const defaultURL = "https://api.infosimples.com/api/v2/consultas/receita-federal/simples"

// Result is the value returned for a successful CNPJ lookup. Raw contains the
// untouched JSON response so callers can persist it for auditability.
type Result struct {
	CNPJ             string
	RazaoSocial      string
	SimplesSituacao  string
	IsSimplesOptante bool
	Raw              []byte
}

// Client talks to InfoSimples. URL/HTTP can be overridden in tests.
type Client struct {
	URL   string
	Token string
	HTTP  *http.Client
}

// New returns a Client configured for the production InfoSimples endpoint.
// Tests can construct a Client literal directly to override URL and HTTP.
func New(token string) *Client {
	return &Client{
		URL:   defaultURL,
		Token: token,
		HTTP:  &http.Client{Timeout: 30 * time.Second},
	}
}

type apiResponse struct {
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

// Lookup queries InfoSimples for a 14-digit CNPJ.
func (c *Client) Lookup(ctx context.Context, cnpj string) (*Result, error) {
	if len(cnpj) != 14 {
		return nil, fmt.Errorf("invalid cnpj length: %d", len(cnpj))
	}
	if c.Token == "" {
		return nil, errors.New("InfoSimplesToken is not set")
	}

	q := url.Values{}
	q.Set("token", c.Token)
	q.Set("timeout", "600")
	q.Set("ignore_site_receipt", "1")
	q.Set("cnpj", cnpj)

	endpoint := c.URL
	if endpoint == "" {
		endpoint = defaultURL
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint+"?"+q.Encode(), nil)
	if err != nil {
		return nil, err
	}
	httpClient := c.HTTP
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("infosimples request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("infosimples read body: %w", err)
	}

	var parsed apiResponse
	if err := json.Unmarshal(body, &parsed); err != nil {
		return nil, fmt.Errorf("infosimples decode: %w", err)
	}
	if parsed.Code != 200 {
		return nil, fmt.Errorf("infosimples cnpj=%s code=%d message=%q", cnpj, parsed.Code, parsed.CodeMessage)
	}
	if len(parsed.Data) == 0 {
		return nil, fmt.Errorf("infosimples cnpj=%s returned empty data", cnpj)
	}

	d := parsed.Data[0]
	return &Result{
		CNPJ:             cnpj,
		RazaoSocial:      d.RazaoSocial,
		SimplesSituacao:  d.SimplesNacionalSituacao,
		IsSimplesOptante: isSimplesOptante(d.SimplesNacionalSituacao),
		Raw:              body,
	}, nil
}

// isSimplesOptante interprets the textual situation. Anything starting with
// "NÃO" (case- and accent-insensitive) is considered NOT optante.
func isSimplesOptante(situacao string) bool {
	s := strings.TrimSpace(situacao)
	if s == "" {
		return false
	}
	folded := strings.ToLower(strings.Map(foldRune, s))
	return !strings.HasPrefix(folded, "nao")
}

func foldRune(r rune) rune {
	switch r {
	case 'á', 'à', 'â', 'ã', 'ä':
		return 'a'
	case 'Á', 'À', 'Â', 'Ã', 'Ä':
		return 'A'
	case 'é', 'è', 'ê', 'ë':
		return 'e'
	case 'É', 'È', 'Ê', 'Ë':
		return 'E'
	case 'í', 'ì', 'î', 'ï':
		return 'i'
	case 'Í', 'Ì', 'Î', 'Ï':
		return 'I'
	case 'ó', 'ò', 'ô', 'õ', 'ö':
		return 'o'
	case 'Ó', 'Ò', 'Ô', 'Õ', 'Ö':
		return 'O'
	case 'ú', 'ù', 'û', 'ü':
		return 'u'
	case 'Ú', 'Ù', 'Û', 'Ü':
		return 'U'
	case 'ç':
		return 'c'
	case 'Ç':
		return 'C'
	}
	return r
}
