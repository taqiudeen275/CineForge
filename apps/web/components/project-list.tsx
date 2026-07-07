"use client";
import type { components } from "@cineforge/api-client";
import { apiFetch } from "@cineforge/api-client";
import { Plus } from "lucide-react";
import Link from "next/link";
import { useEffect, useState } from "react";
import { EmptyState, PageHeader } from "@/components/ui";
import { useShell } from "@/components/app-shell";
type Project = components["schemas"]["Project"];
export function ProjectList() {
  const { workspace } = useShell();
  const [items, setItems] = useState<Project[]>([]);
  const [loading, setLoading] = useState(true);
  useEffect(() => {
    if (workspace)
      apiFetch<{ items: Project[] }>(`/v1/workspaces/${workspace.id}/projects?status=active`)
        .then((v) => setItems(v.items))
        .finally(() => setLoading(false));
  }, [workspace]);
  if (!workspace) return null;
  return (
    <>
      <PageHeader
        eyebrow="Workspace"
        title="Projects"
        description="Private story productions and their current settings."
        action={
          <Link
            className="inline-flex min-h-10 items-center gap-2 rounded-lg bg-[var(--accent)] px-4 text-sm font-medium text-white"
            href={`/app/${workspace.slug}/projects/new`}
          >
            <Plus size={16} />
            New project
          </Link>
        }
      />
      {loading ? (
        <p className="py-16 text-sm text-[var(--muted)]">Loading projects…</p>
      ) : items.length ? (
        <div className="divide-y divide-[var(--line)]">
          {items.map((p) => (
            <Link
              key={p.id}
              href={`/app/${workspace.slug}/projects/${p.id}`}
              className="grid gap-2 py-6 sm:grid-cols-[1fr_auto] sm:items-center"
            >
              <div>
                <h2 className="m-0 text-base font-medium">{p.settings.name}</h2>
                <p className="mb-0 mt-1 text-sm text-[var(--muted)]">
                  {label(p.settings.productionFormat)} · {p.settings.aspectWidth ?? 16}:
                  {p.settings.aspectHeight ?? 9} · Private
                </p>
              </div>
              <span className="text-sm text-[var(--muted)]">
                Updated version {p.currentVersion}
              </span>
            </Link>
          ))}
        </div>
      ) : (
        <EmptyState
          title="Create the first project"
          description="Start with a name and format. Every project is private by default."
          action={
            <Link
              className="rounded-lg bg-[var(--accent)] px-4 py-2.5 text-sm font-medium text-white"
              href={`/app/${workspace.slug}/projects/new`}
            >
              New project
            </Link>
          }
        />
      )}
    </>
  );
}
function label(v: string) {
  return v.replaceAll("_", " ").replace(/\b\w/g, (c) => c.toUpperCase());
}
