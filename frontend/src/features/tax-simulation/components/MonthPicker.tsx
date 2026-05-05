import { Input, Label } from "@/components/Input";

export function MonthPicker({
  value,
  onChange,
  id = "month-picker",
}: {
  value: string;
  onChange: (value: string) => void;
  id?: string;
}) {
  return (
    <div className="flex items-center gap-3">
      <Label htmlFor={id} className="whitespace-nowrap">
        Mês de referência
      </Label>
      <Input
        id={id}
        type="month"
        value={value}
        onChange={(e) => onChange(e.target.value)}
        className="w-44"
      />
    </div>
  );
}
