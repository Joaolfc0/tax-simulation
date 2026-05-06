import { apiClient } from "./client";
import { isUUID } from "@/lib/format";

export interface UploadProcessResult {
  upload_id: string;
  status: string;
  rows_processed: number;
  flagged_rows: number;
  unique_cnpjs: number;
}

/**
 * Sends an invoice spreadsheet (CSV/XLSX) for the given company.
 *
 * Wraps the raw `UploadFile` endpoint exposed by the generated client.
 * The endpoint is `raw` on the backend, so the body is `multipart/form-data`
 * (browser sets the boundary automatically).
 */
export async function uploadInvoiceFile(
  companyId: string,
  file: File,
): Promise<UploadProcessResult> {
  const normalizedCompanyId = companyId.trim();
  if (!isUUID(normalizedCompanyId)) {
    throw new Error(
      "Empresa inválida para upload. Volte para a tela de empresas e selecione uma empresa válida.",
    );
  }

  const fd = new FormData();
  fd.append("file", file);

  const res = await apiClient.tax.UploadFile("POST", normalizedCompanyId, fd);
  if (!res.ok) {
    const text = await res.text().catch(() => res.statusText);
    throw new Error(text || `Upload falhou (HTTP ${res.status})`);
  }
  return (await res.json()) as UploadProcessResult;
}
