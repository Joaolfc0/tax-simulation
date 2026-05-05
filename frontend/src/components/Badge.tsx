import type { ReactNode } from "react";
import { cn } from "./cn";

type Tone = "neutral" | "subtle" | "strong" | "warn";

const tones: Record<Tone, string> = {
  neutral: "bg-muted text-ink border-line",
  subtle: "bg-white text-ink-soft border-line",
  strong: "bg-ink text-white border-ink",
  warn: "bg-white text-ink border-ink",
};

export function Badge({
  children,
  tone = "neutral",
  className,
}: {
  children: ReactNode;
  tone?: Tone;
  className?: string;
}) {
  return (
    <span
      className={cn(
        "inline-flex items-center gap-1 rounded-full border px-2 py-0.5 text-xs font-medium",
        tones[tone],
        className,
      )}
    >
      {children}
    </span>
  );
}
