import { useEffect, useRef, useState } from "react";

import { useNavigate } from "react-router-dom";
import { Button } from "@/components/Button";
import { Input, Label } from "@/components/Input";
import { useCreateCompany } from "../hooks";

export function NewCompanyDialog({
  open,
  onClose,
}: {
  open: boolean;
  onClose: () => void;
}) {
  const [name, setName] = useState("");
  const inputRef = useRef<HTMLInputElement>(null);
  const navigate = useNavigate();
  const createCompany = useCreateCompany();

  const resetMutation = createCompany.reset;
  const resetRef = useRef(resetMutation);
  resetRef.current = resetMutation;

  useEffect(() => {
    if (!open) return;
    setName("");
    resetRef.current();
    const handle = window.setTimeout(() => inputRef.current?.focus(), 50);
    return () => window.clearTimeout(handle);
  }, [open]);

  useEffect(() => {
    if (!open) return;
    const handleKey = (e: KeyboardEvent) => {
      if (e.key === "Escape") onClose();
    };
    window.addEventListener("keydown", handleKey);
    return () => window.removeEventListener("keydown", handleKey);
  }, [open, onClose]);

  if (!open) return null;

  const submit = async (e: React.FormEvent) => {
    e.preventDefault();
    const trimmed = name.trim();
    if (!trimmed) return;
    const company = await createCompany.mutateAsync({ name: trimmed });
    onClose();
    navigate(`/companies/${company.id}`);
  };

  return (
    <div
      className="fixed inset-0 z-50 flex items-center justify-center bg-black/30 px-4"
      role="dialog"
      aria-modal="true"
      aria-labelledby="new-company-title"
      onMouseDown={(e) => {
        if (e.target === e.currentTarget) onClose();
      }}
    >
      <form
        onSubmit={submit}
        className="w-full max-w-md rounded-lg border border-line bg-white shadow-lg"
      >
        <div className="border-b border-line px-6 py-5">
          <h2 id="new-company-title" className="text-base font-semibold text-ink">
            Nova empresa
          </h2>
          <p className="mt-1 text-sm text-ink-soft">
            Dê um nome para a empresa que você quer simular.
          </p>
        </div>
        <div className="space-y-2 px-6 py-5">
          <Label htmlFor="company-name">Nome da empresa</Label>
          <Input
            id="company-name"
            ref={inputRef}
            value={name}
            onChange={(e) => setName(e.target.value)}
            placeholder="Ex.: Acme Comércio Ltda."
            disabled={createCompany.isPending}
            required
          />
          {createCompany.isError ? (
            <p className="text-xs text-ink">
              Erro ao criar empresa: {(createCompany.error as Error).message}
            </p>
          ) : null}
        </div>
        <div className="flex items-center justify-end gap-2 border-t border-line bg-muted/40 px-6 py-4">
          <Button
            type="button"
            variant="ghost"
            onClick={onClose}
            disabled={createCompany.isPending}
          >
            Cancelar
          </Button>
          <Button
            type="submit"
            disabled={createCompany.isPending || !name.trim()}
          >
            {createCompany.isPending ? "Criando..." : "Criar empresa"}
          </Button>
        </div>
      </form>
    </div>
  );
}
