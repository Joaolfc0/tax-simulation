import { Link } from "react-router-dom";
import type { tax } from "@/lib/client.gen";
import { formatDate } from "@/lib/format";

export function CompanyCard({ company }: { company: tax.Company }) {
  return (
    <Link
      to={`/companies/${company.id}`}
      className="group block rounded-lg border border-line bg-white p-5 shadow-soft transition-colors hover:border-ink/40 focus-ring"
    >
      <div className="flex items-start justify-between gap-3">
        <div className="min-w-0">
          <h3 className="truncate text-base font-semibold text-ink">
            {company.name}
          </h3>
          <p className="mt-1 text-xs text-ink-faint">
            Criada em {formatDate(company["created_at"])}
          </p>
        </div>
        <span
          aria-hidden
          className="text-ink-faint transition-colors group-hover:text-ink"
        >
          →
        </span>
      </div>
    </Link>
  );
}
