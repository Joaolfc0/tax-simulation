import { useState } from "react";
import { Button } from "@/components/Button";
import { EmptyState } from "@/components/EmptyState";
import { Skeleton } from "@/components/Skeleton";
import { CompanyCard } from "./components/CompanyCard";
import { NewCompanyDialog } from "./components/NewCompanyDialog";
import { useCompanies } from "./hooks";

export function CompanySelectPage() {
  const [dialogOpen, setDialogOpen] = useState(false);
  const { data: companies, isLoading, isError, error, refetch } = useCompanies();

  return (
    <div className="space-y-8">
      <section className="rounded-lg border border-line bg-muted/40 px-6 py-8">
        <div className="flex flex-wrap items-center gap-4">
          <img
            src="/logo.webp?v=2"
            alt="Soma Control"
            className="h-24 w-24 rounded-lg border border-ink bg-ink p-3 object-contain shadow-soft"
          />
          <div>
            <p className="text-xs font-medium uppercase tracking-wide text-ink-faint">
              Plataforma
            </p>
            <h1 className="text-2xl font-semibold tracking-tight text-ink">
              Soma Control
            </h1>
          </div>
        </div>
      </section>

      <header className="flex flex-wrap items-end justify-between gap-4">
        <div>
          <h2 className="text-2xl font-semibold tracking-tight text-ink">
            Empresas
          </h2>
          <p className="mt-1 text-sm text-ink-soft">
            Selecione uma empresa para acessar suas ferramentas, ou crie uma
            nova.
          </p>
        </div>
        <Button onClick={() => setDialogOpen(true)}>+ Nova empresa</Button>
      </header>

      {isLoading ? (
        <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-3">
          {Array.from({ length: 3 }).map((_, i) => (
            <Skeleton key={i} className="h-24" />
          ))}
        </div>
      ) : isError ? (
        <EmptyState
          title="Não foi possível carregar as empresas"
          description={(error as Error).message}
          action={
            <Button variant="secondary" onClick={() => refetch()}>
              Tentar novamente
            </Button>
          }
        />
      ) : !companies || companies.length === 0 ? (
        <EmptyState
          title="Nenhuma empresa ainda"
          description="Crie a primeira empresa para começar a usar a simulação de impostos."
          action={
            <Button onClick={() => setDialogOpen(true)}>
              Criar primeira empresa
            </Button>
          }
        />
      ) : (
        <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-3">
          {companies.map((company) => (
            <CompanyCard key={company.id} company={company} />
          ))}
        </div>
      )}

      <NewCompanyDialog
        open={dialogOpen}
        onClose={() => setDialogOpen(false)}
      />
    </div>
  );
}
