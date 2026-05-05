import type { tax } from "@/lib/client.gen";
import { formatBRL, formatNumber } from "@/lib/format";

export function ReportSummary({ report }: { report: tax.MonthReport }) {
  const flaggedShare =
    report["total_valor_contabil"] > 0
      ? (report["flagged_sum"] / report["total_valor_contabil"]) * 100
      : 0;

  return (
    <div className="grid grid-cols-1 gap-px overflow-hidden rounded-md border border-line bg-line sm:grid-cols-4">
      <Stat
        label="Notas no mês"
        value={formatNumber(report["total_entries"])}
      />
      <Stat
        label="Total contábil"
        value={formatBRL(report["total_valor_contabil"])}
      />
      <Stat
        label="Notas marcadas"
        value={formatNumber(report["flagged_count"])}
        hint="Fornecedor fora do Simples"
      />
      <Stat
        label="Valor marcado"
        value={formatBRL(report["flagged_sum"])}
        hint={`${flaggedShare.toFixed(1)}% do total`}
      />
    </div>
  );
}

function Stat({
  label,
  value,
  hint,
}: {
  label: string;
  value: string;
  hint?: string;
}) {
  return (
    <div className="bg-white px-5 py-4">
      <div className="text-xs font-medium uppercase tracking-wide text-ink-faint">
        {label}
      </div>
      <div className="mt-1 text-xl font-semibold text-ink">{value}</div>
      {hint ? <div className="mt-0.5 text-xs text-ink-faint">{hint}</div> : null}
    </div>
  );
}
