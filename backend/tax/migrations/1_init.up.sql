CREATE TABLE companies (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name       TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE uploads (
    id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    company_id     UUID NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
    filename       TEXT NOT NULL,
    content_type   TEXT NOT NULL,
    size_bytes     BIGINT NOT NULL,
    storage_key    TEXT NOT NULL,
    status         TEXT NOT NULL DEFAULT 'pending',
    error          TEXT,
    rows_processed INT NOT NULL DEFAULT 0,
    uploaded_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    processed_at   TIMESTAMPTZ
);

CREATE INDEX idx_uploads_company ON uploads(company_id);

CREATE TABLE suppliers (
    cnpj               TEXT PRIMARY KEY,
    razao_social       TEXT,
    simples_situacao   TEXT,
    is_simples_optante BOOLEAN NOT NULL,
    last_checked_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    last_response      JSONB
);

CREATE INDEX idx_suppliers_optante ON suppliers(is_simples_optante);

CREATE TABLE invoice_entries (
    id                BIGSERIAL PRIMARY KEY,
    company_id        UUID NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
    upload_id         UUID NOT NULL REFERENCES uploads(id) ON DELETE CASCADE,
    supplier_cnpj     TEXT NOT NULL REFERENCES suppliers(cnpj),
    supplier_name     TEXT NOT NULL,
    escrituracao_date DATE NOT NULL,
    numero            TEXT,
    especie           TEXT,
    cst               TEXT,
    uf                TEXT,
    cfop              TEXT,
    valor_contabil    NUMERIC(14,2) NOT NULL,
    vl_outros_icms    NUMERIC(14,2) NOT NULL DEFAULT 0,
    vl_isentos_icms   NUMERIC(14,2) NOT NULL DEFAULT 0,
    flagged           BOOLEAN NOT NULL,
    created_at        TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_entries_company_date ON invoice_entries(company_id, escrituracao_date);
CREATE INDEX idx_entries_company_flagged ON invoice_entries(company_id) WHERE flagged = TRUE;
CREATE INDEX idx_entries_upload ON invoice_entries(upload_id);
CREATE INDEX idx_entries_supplier ON invoice_entries(supplier_cnpj);
