import { Badge } from "@/components/Badge";
import { Skeleton } from "@/components/Skeleton";
import type { tax } from "@/lib/client.gen";
import { formatBytes, formatDateTime, formatNumber } from "@/lib/format";
import { useUploads } from "../hooks/useUploads";

export function UploadHistory({ companyId }: { companyId: string }) {
  const { data: uploads, isLoading, isError, error } = useUploads(companyId);

  if (isLoading) {
    return (
      <div className="space-y-2">
        {Array.from({ length: 3 }).map((_, i) => (
          <Skeleton key={i} className="h-12" />
        ))}
      </div>
    );
  }

  if (isError) {
    return (
      <div className="rounded-md border border-line bg-muted px-4 py-3 text-sm text-ink">
        Erro ao listar uploads: {(error as Error).message}
      </div>
    );
  }

  if (!uploads || uploads.length === 0) {
    return (
      <div className="rounded-md border border-dashed border-line bg-white px-6 py-8 text-center text-sm text-ink-soft">
        Ainda não há uploads para esta empresa.
      </div>
    );
  }

  return (
    <div className="overflow-hidden rounded-md border border-line">
      <div className="overflow-x-auto">
        <table className="w-full border-collapse text-sm">
          <thead className="border-b border-line bg-muted text-left text-xs uppercase tracking-wide text-ink-faint">
            <tr>
              <th className="px-4 py-3 font-medium">Arquivo</th>
              <th className="px-4 py-3 font-medium">Status</th>
              <th className="px-4 py-3 text-right font-medium">Linhas</th>
              <th className="px-4 py-3 text-right font-medium">Tamanho</th>
              <th className="px-4 py-3 font-medium">Enviado em</th>
            </tr>
          </thead>
          <tbody>
            {uploads.map((u) => (
              <tr key={u.id} className="border-t border-line align-top">
                <td className="px-4 py-3">
                  <div className="font-medium text-ink">{u.filename}</div>
                  {u.error ? (
                    <div className="mt-1 text-xs text-ink-soft">
                      {u.error}
                    </div>
                  ) : null}
                </td>
                <td className="px-4 py-3">
                  <UploadStatusBadge status={u.status} />
                </td>
                <td className="px-4 py-3 text-right text-ink-soft">
                  {formatNumber(u["rows_processed"])}
                </td>
                <td className="px-4 py-3 text-right text-ink-soft">
                  {formatBytes(u["size_bytes"])}
                </td>
                <td className="px-4 py-3 whitespace-nowrap text-ink-soft">
                  {formatDateTime(u["uploaded_at"])}
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </div>
  );
}

function UploadStatusBadge({ status }: { status: tax.Upload["status"] }) {
  switch (status) {
    case "completed":
      return <Badge tone="strong">Concluído</Badge>;
    case "processing":
    case "pending":
      return <Badge tone="subtle">Processando</Badge>;
    case "failed":
      return <Badge tone="warn">Falhou</Badge>;
    default:
      return <Badge tone="subtle">{status}</Badge>;
  }
}
