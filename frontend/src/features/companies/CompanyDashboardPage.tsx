import { Link, useParams } from "react-router-dom";
import { Badge } from "@/components/Badge";
import { Button } from "@/components/Button";
import { EmptyState } from "@/components/EmptyState";
import { Skeleton } from "@/components/Skeleton";
import { listEnabledFeatures } from "@/features/registry";
import { useCompany } from "./hooks";

export function CompanyDashboardPage() {
  const { companyId } = useParams<{ companyId: string }>();
  const { data: company, isLoading, isError, error } = useCompany(companyId);
  const features = listEnabledFeatures();

  return (
    <div className="space-y-8">
      <div>
        <Link
          to="/companies"
          className="text-xs text-ink-faint hover:text-ink focus-ring rounded-sm"
        >
          ← Empresas
        </Link>
        <div className="mt-2">
          {isLoading ? (
            <Skeleton className="h-8 w-64" />
          ) : isError ? (
            <h1 className="text-2xl font-semibold text-ink">
              Empresa não encontrada
            </h1>
          ) : (
            <h1 className="text-2xl font-semibold tracking-tight text-ink">
              {company?.name}
            </h1>
          )}
          <p className="mt-1 text-sm text-ink-soft">
            Escolha uma ferramenta para esta empresa.
          </p>
        </div>
      </div>

      {isError ? (
        <EmptyState
          title="Não foi possível carregar a empresa"
          description={(error as Error).message}
          action={
            <Link to="/companies">
              <Button variant="secondary">Voltar</Button>
            </Link>
          }
        />
      ) : features.length === 0 ? (
        <EmptyState
          title="Nenhuma feature disponível"
          description="Cadastre features no registry para que elas apareçam aqui."
        />
      ) : (
        <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-3">
          {features.map((feature) => (
            <Link
              key={feature.id}
              to={companyId ? feature.to(companyId) : "#"}
              className="group flex h-full flex-col justify-between rounded-lg border border-line bg-white p-5 shadow-soft transition-colors hover:border-ink/40 focus-ring"
            >
              <div>
                <div className="flex items-center gap-2">
                  <h3 className="text-base font-semibold text-ink">
                    {feature.label}
                  </h3>
                  {feature.badge ? (
                    <Badge tone="subtle">{feature.badge}</Badge>
                  ) : null}
                </div>
                <p className="mt-2 text-sm text-ink-soft">
                  {feature.description}
                </p>
              </div>
              <div className="mt-6 inline-flex items-center gap-1 text-sm font-medium text-ink transition-colors group-hover:text-ink-soft">
                Abrir
                <span aria-hidden>→</span>
              </div>
            </Link>
          ))}
        </div>
      )}
    </div>
  );
}
