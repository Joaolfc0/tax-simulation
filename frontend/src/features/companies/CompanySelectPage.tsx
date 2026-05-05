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
      <header className="flex flex-wrap items-end justify-between gap-4">
        <div>
          <h1 className="text-2xl font-semibold tracking-tight text-ink">
            Empresas
          </h1>
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
