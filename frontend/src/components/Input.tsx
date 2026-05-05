import { forwardRef, type InputHTMLAttributes } from "react";
import { cn } from "./cn";

export type InputProps = InputHTMLAttributes<HTMLInputElement>;

export const Input = forwardRef<HTMLInputElement, InputProps>(
  ({ className, ...props }, ref) => {
    return (
      <input
        ref={ref}
        className={cn(
          "h-10 w-full rounded-md border border-line bg-white px-3 text-sm text-ink placeholder:text-ink-faint",
          "focus-ring transition-colors hover:border-ink-faint focus:border-ink",
          "disabled:cursor-not-allowed disabled:bg-muted disabled:text-ink-faint",
          className,
        )}
        {...props}
      />
    );
  },
);
Input.displayName = "Input";

export function Label({
  htmlFor,
  children,
  className,
}: {
  htmlFor?: string;
  children: React.ReactNode;
  className?: string;
}) {
  return (
    <label
      htmlFor={htmlFor}
      className={cn("text-sm font-medium text-ink", className)}
    >
      {children}
    </label>
  );
}
