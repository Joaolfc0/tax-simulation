import { apiClient } from "./client";

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
  const fd = new FormData();
  fd.append("file", file);

  const res = await apiClient.tax.UploadFile("POST", companyId, fd);
  if (!res.ok) {
    const text = await res.text().catch(() => res.statusText);
    throw new Error(text || `Upload falhou (HTTP ${res.status})`);
  }
  return (await res.json()) as UploadProcessResult;
}
