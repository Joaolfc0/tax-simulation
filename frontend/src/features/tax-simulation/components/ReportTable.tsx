import { Badge } from "@/components/Badge";
import { cn } from "@/components/cn";
import type { tax } from "@/lib/client.gen";
import { formatBRL, formatCNPJ, formatDate } from "@/lib/format";

export function ReportTable({ entries }: { entries: tax.ReportEntry[] }) {
  if (entries.length === 0) {
    return (
      <div className="rounded-md border border-dashed border-line bg-white px-6 py-10 text-center text-sm text-ink-soft">
        Nenhuma nota nesse mês.
      </div>
    );
  }

  return (
    <div className="overflow-hidden rounded-md border border-line">
      <div className="overflow-x-auto">
        <table className="w-full border-collapse text-sm">
          <thead className="border-b border-line bg-muted text-left text-xs uppercase tracking-wide text-ink-faint">
            <tr>
              <Th>Data</Th>
              <Th>Fornecedor</Th>
              <Th>Nº</Th>
              <Th>CFOP</Th>
              <Th>UF</Th>
              <Th align="right">Valor contábil</Th>
              <Th>Situação</Th>
            </tr>
          </thead>
          <tbody>
            {entries.map((entry) => (
              <tr
                key={entry.id}
                className={cn(
                  "border-t border-line align-top transition-colors hover:bg-muted/40",
                  entry.flagged && "bg-muted/30",
                )}
              >
                <Td className="whitespace-nowrap text-ink-soft">
                  {formatDate(entry.date)}
                </Td>
                <Td>
                  <div className="font-medium text-ink">
                    {entry["supplier_name"] || "—"}
                  </div>
                  <div className="font-mono text-xs text-ink-faint">
                    {formatCNPJ(entry["supplier_cnpj"])}
                  </div>
                </Td>
                <Td className="whitespace-nowrap text-ink-soft">
                  {entry.numero || "—"}
                </Td>
                <Td className="whitespace-nowrap text-ink-soft">
                  {entry.cfop || "—"}
                </Td>
                <Td className="whitespace-nowrap text-ink-soft">
                  {entry.uf || "—"}
                </Td>
                <Td align="right" className="whitespace-nowrap font-medium text-ink">
                  {formatBRL(entry["valor_contabil"])}
                </Td>
                <Td>
                  {entry.flagged ? (
                    <Badge tone="strong">Não optante</Badge>
                  ) : (
                    <Badge tone="subtle">Optante</Badge>
                  )}
                  {entry["simples_situacao"] ? (
                    <div className="mt-1 text-xs text-ink-faint">
                      {entry["simples_situacao"]}
                    </div>
                  ) : null}
                </Td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </div>
  );
}

function Th({
  children,
  align = "left",
}: {
  children: React.ReactNode;
  align?: "left" | "right";
}) {
  return (
    <th
      className={cn(
        "px-4 py-3 font-medium",
        align === "right" ? "text-right" : "text-left",
      )}
    >
      {children}
    </th>
  );
}

function Td({
  children,
  align = "left",
  className,
}: {
  children: React.ReactNode;
  align?: "left" | "right";
  className?: string;
}) {
  return (
    <td
      className={cn(
        "px-4 py-3",
        align === "right" ? "text-right" : "text-left",
        className,
      )}
    >
      {children}
    </td>
  );
}
