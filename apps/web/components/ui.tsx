import {
  cloneElement,
  isValidElement,
  useId,
  type ButtonHTMLAttributes,
  type InputHTMLAttributes,
  type ReactElement,
  type ReactNode,
} from "react";
import { cn } from "@/lib/cn";

export function Button({
  className,
  variant = "primary",
  ...props
}: ButtonHTMLAttributes<HTMLButtonElement> & { variant?: "primary" | "quiet" | "danger" }) {
  return (
    <button
      className={cn(
        "inline-flex min-h-10 items-center justify-center gap-2 rounded-lg px-4 text-sm font-medium transition disabled:cursor-not-allowed disabled:opacity-50",
        variant === "primary" && "bg-[var(--accent)] text-white hover:bg-[var(--accent-strong)]",
        variant === "quiet" &&
          "text-[var(--muted)] hover:bg-[var(--surface-raised)] hover:text-[var(--text)]",
        variant === "danger" && "bg-[var(--danger)] text-white",
        className,
      )}
      {...props}
    />
  );
}

export function Input({ className, ...props }: InputHTMLAttributes<HTMLInputElement>) {
  return (
    <input
      className={cn(
        "min-h-11 w-full rounded-lg border border-[var(--line)] bg-[var(--surface)] px-3 text-sm text-[var(--text)] placeholder:text-[var(--subtle)]",
        className,
      )}
      {...props}
    />
  );
}

export function Select({
  className,
  children,
  ...props
}: React.SelectHTMLAttributes<HTMLSelectElement>) {
  return (
    <select
      className={cn(
        "min-h-11 w-full rounded-lg border border-[var(--line)] bg-[var(--surface)] px-3 text-sm",
        className,
      )}
      {...props}
    >
      {children}
    </select>
  );
}

export function Field({
  label,
  hint,
  error,
  children,
}: {
  label: string;
  hint?: string;
  error?: string;
  children: ReactNode;
}) {
  const generatedId = useId();
  const element = isValidElement(children)
    ? cloneElement(children as ReactElement<{ id?: string }>, {
        id: (children.props as { id?: string }).id ?? generatedId,
      })
    : children;
  return (
    <div className="grid gap-2 text-sm">
      <label className="font-medium" htmlFor={generatedId}>
        {label}
      </label>
      {element}
      {error ? (
        <span className="text-xs text-[var(--danger)]">{error}</span>
      ) : hint ? (
        <span className="text-xs text-[var(--muted)]">{hint}</span>
      ) : null}
    </div>
  );
}

export function EmptyState({
  title,
  description,
  action,
}: {
  title: string;
  description: string;
  action?: ReactNode;
}) {
  return (
    <div className="mx-auto grid max-w-md justify-items-center gap-3 py-20 text-center">
      <div className="text-lg font-semibold">{title}</div>
      <p className="m-0 text-sm leading-6 text-[var(--muted)]">{description}</p>
      {action}
    </div>
  );
}

export function PageHeader({
  eyebrow,
  title,
  description,
  action,
}: {
  eyebrow?: string;
  title: string;
  description?: string;
  action?: ReactNode;
}) {
  return (
    <header className="flex flex-col justify-between gap-5 border-b border-[var(--line)] pb-7 sm:flex-row sm:items-end">
      {" "}
      <div className="max-w-2xl">
        {eyebrow ? (
          <div className="mb-2 text-xs font-semibold uppercase tracking-[.14em] text-[var(--accent)]">
            {eyebrow}
          </div>
        ) : null}
        <h1 className="m-0 text-3xl font-semibold tracking-[-.03em]">{title}</h1>
        {description ? (
          <p className="mb-0 mt-2 text-sm leading-6 text-[var(--muted)]">{description}</p>
        ) : null}
      </div>
      {action}
    </header>
  );
}
