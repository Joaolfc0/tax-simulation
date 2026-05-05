import { useRef, useState, type DragEvent } from "react";
import { Button } from "@/components/Button";
import { cn } from "@/components/cn";
import { formatBytes, formatNumber } from "@/lib/format";
import { useUploadFile } from "../hooks/useUploadFile";

const ACCEPTED = ".csv,.xlsx";
const ACCEPTED_LABEL = "CSV ou XLSX";

export function UploadDropzone({ companyId }: { companyId: string }) {
  const inputRef = useRef<HTMLInputElement>(null);
  const [dragOver, setDragOver] = useState(false);
  const [pendingFile, setPendingFile] = useState<File | null>(null);
  const upload = useUploadFile(companyId);

  const pickFile = (file: File | undefined | null) => {
    if (!file) return;
    setPendingFile(file);
    upload.reset();
  };

  const onDrop = (e: DragEvent<HTMLLabelElement>) => {
    e.preventDefault();
    setDragOver(false);
    pickFile(e.dataTransfer.files?.[0]);
  };

  const onDragOver = (e: DragEvent<HTMLLabelElement>) => {
    e.preventDefault();
    setDragOver(true);
  };

  const submit = () => {
    if (!pendingFile) return;
    upload.mutate(pendingFile);
  };

  const result = upload.data;

  return (
    <div className="space-y-4">
      <label
        htmlFor="invoice-file"
        onDrop={onDrop}
        onDragOver={onDragOver}
        onDragLeave={() => setDragOver(false)}
        className={cn(
          "flex cursor-pointer flex-col items-center justify-center gap-2 rounded-md border border-dashed bg-white px-6 py-10 text-center transition-colors",
          dragOver
            ? "border-ink bg-muted"
            : "border-line hover:border-ink/40 hover:bg-muted/40",
        )}
      >
        <span className="text-sm font-medium text-ink">
          {pendingFile
            ? pendingFile.name
            : "Arraste a planilha aqui ou clique para selecionar"}
        </span>
        <span className="text-xs text-ink-faint">
          {pendingFile
            ? formatBytes(pendingFile.size)
            : `Aceita ${ACCEPTED_LABEL}`}
        </span>
        <input
          id="invoice-file"
          ref={inputRef}
          type="file"
          accept={ACCEPTED}
          className="sr-only"
          onChange={(e) => pickFile(e.target.files?.[0])}
        />
      </label>

      <div className="flex flex-wrap items-center gap-3">
        <Button onClick={submit} disabled={!pendingFile || upload.isPending}>
          {upload.isPending ? "Processando..." : "Processar planilha"}
        </Button>
        {pendingFile ? (
          <Button
            variant="ghost"
            onClick={() => {
              setPendingFile(null);
              upload.reset();
              if (inputRef.current) inputRef.current.value = "";
            }}
            disabled={upload.isPending}
          >
            Limpar
          </Button>
        ) : null}
      </div>

      {upload.isError ? (
        <div className="rounded-md border border-line bg-muted px-4 py-3 text-sm text-ink">
          <strong className="font-medium">Falha no processamento:</strong>{" "}
          {(upload.error as Error).message}
        </div>
      ) : null}

      {result ? (
        <div className="grid grid-cols-1 gap-3 rounded-md border border-line bg-muted/40 px-4 py-4 sm:grid-cols-3">
          <ResultStat
            label="Linhas processadas"
            value={formatNumber(result.rows_processed)}
          />
          <ResultStat
            label="Fornecedores únicos"
            value={formatNumber(result.unique_cnpjs)}
          />
          <ResultStat
            label="Linhas marcadas"
            value={formatNumber(result.flagged_rows)}
            hint="Fornecedores fora do Simples"
          />
        </div>
      ) : null}
    </div>
  );
}

function ResultStat({
  label,
  value,
  hint,
}: {
  label: string;
  value: string;
  hint?: string;
}) {
  return (
    <div>
      <div className="text-xs font-medium uppercase tracking-wide text-ink-faint">
        {label}
      </div>
      <div className="mt-1 text-lg font-semibold text-ink">{value}</div>
      {hint ? <div className="text-xs text-ink-faint">{hint}</div> : null}
    </div>
  );
}
