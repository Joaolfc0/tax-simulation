import { useState } from "react";
import { Link, useParams } from "react-router-dom";
import { Card, CardBody, CardHeader } from "@/components/Card";
import { EmptyState } from "@/components/EmptyState";
import { Skeleton } from "@/components/Skeleton";
import { Button } from "@/components/Button";
import { useCompany } from "@/features/companies/hooks";
import { currentYearMonth, isUUID } from "@/lib/format";
import { MonthPicker } from "./components/MonthPicker";
import { ReportSummary } from "./components/ReportSummary";
import { ReportTable } from "./components/ReportTable";
import { UploadDropzone } from "./components/UploadDropzone";
import { UploadHistory } from "./components/UploadHistory";
import { useMonthReport } from "./hooks/useMonthReport";

export function SimulationPage() {
  const { companyId } = useParams<{ companyId: string }>();
  const company = useCompany(companyId);
  const [yearMonth, setYearMonth] = useState(currentYearMonth());
  const report = useMonthReport(companyId, yearMonth);

  if (!companyId) {
    return (
      <EmptyState
        title="Empresa não informada"
        description="Selecione uma empresa para abrir a simulação."
        action={
          <Link to="/companies">
            <Button variant="secondary">Voltar para empresas</Button>
          </Link>
        }
      />
    );
  }

  if (!isUUID(companyId)) {
    return (
      <EmptyState
        title="Empresa inválida"
        description="O identificador da empresa nesta URL é inválido. Volte para a lista e selecione a empresa novamente."
        action={
          <Link to="/companies">
            <Button variant="secondary">Voltar para empresas</Button>
          </Link>
        }
      />
    );
  }

  return (
    <div className="space-y-8">
      <div>
        <Link
          to={`/companies/${companyId}`}
          className="text-xs text-ink-faint hover:text-ink focus-ring rounded-sm"
        >
          ← {company.data?.name ?? "Empresa"}
        </Link>
        <h1 className="mt-2 text-2xl font-semibold tracking-tight text-ink">
          Simulação de impostos
        </h1>
        <p className="mt-1 text-sm text-ink-soft">
          Importe planilhas de notas fiscais para identificar fornecedores fora
          do Simples Nacional e revisar o impacto contábil mês a mês.
        </p>
      </div>

      <Card>
        <CardHeader
          title="1. Importar planilha"
          description="Aceita CSV ou XLSX com as colunas Dt. Escrituração, CNPJ, Nome do fornecedor, CFOP, UF, Valor Contábil e ICMS."
        />
        <CardBody>
          <UploadDropzone companyId={companyId} />
        </CardBody>
      </Card>

      <Card>
        <CardHeader
          title="2. Relatório mensal"
          description="Selecione o mês de referência. Linhas marcadas indicam fornecedores que não estão optantes pelo Simples na consulta InfoSimples."
          action={<MonthPicker value={yearMonth} onChange={setYearMonth} />}
        />
        <CardBody className="space-y-5">
          {report.isLoading ? (
            <div className="space-y-4">
              <Skeleton className="h-24" />
              <Skeleton className="h-64" />
            </div>
          ) : report.isError ? (
            <div className="rounded-md border border-line bg-muted px-4 py-3 text-sm text-ink">
              Erro ao buscar relatório: {(report.error as Error).message}
            </div>
          ) : report.data ? (
            <>
              <ReportSummary report={report.data} />
              <ReportTable entries={report.data.entries} />
            </>
          ) : null}
        </CardBody>
      </Card>

      <Card>
        <CardHeader
          title="3. Histórico de uploads"
          description="Lista cronológica das planilhas enviadas para esta empresa."
        />
        <CardBody>
          <UploadHistory companyId={companyId} />
        </CardBody>
      </Card>
    </div>
  );
}
