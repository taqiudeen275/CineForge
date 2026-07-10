"use client";
import type { components } from "@cineforge/api-client";
import { apiFetch } from "@cineforge/api-client";
import { ArchiveRestore, Film, Plus, Search, ShieldCheck } from "lucide-react";
import Link from "next/link";
import { useEffect, useState } from "react";
import { Button, EmptyState, Input, PageHeader } from "@/components/ui";
import { useShell } from "@/components/app-shell";
type Project = components["schemas"]["Project"];
type Status = "active" | "archived" | "trashed";
export function ProjectList() {
  const { workspace, canEdit } = useShell();
  const [items, setItems] = useState<Project[]>([]);
  const [status, setStatus] = useState<Status>("active");
  const [query, setQuery] = useState("");
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");
  // The query inputs intentionally control the debounced reload.
  // biome-ignore lint/correctness/useExhaustiveDependencies: load is the request helper for these inputs.
  useEffect(() => {
    const timer = setTimeout(() => void load(), query ? 250 : 0);
    return () => clearTimeout(timer);
  }, [workspace, status, query]);
  async function load() {
    if (!workspace) return;
    setLoading(true);
    setError("");
    try {
      const r = await apiFetch<{ items: Project[] }>(
        `/v1/workspaces/${workspace.id}/projects?status=${status}&q=${encodeURIComponent(query)}`,
      );
      setItems(r.items);
    } catch (e) {
      setError(e instanceof Error ? e.message : "Could not load projects");
    } finally {
      setLoading(false);
    }
  }
  async function restore(id: string) {
    await apiFetch(`/v1/projects/${id}/restore`, { method: "POST", body: {} });
    await load();
  }
  if (!workspace) return null;
  return (
    <>
      <PageHeader
        eyebrow={`${workspace.name} · ${workspace.currentMembership?.role ?? "member"}`}
        title="Projects"
        description="Private productions, templates, and production settings in one place."
        action={
          canEdit ? (
            <Link className="button button-primary" href={`/app/${workspace.slug}/projects/new`}>
              <Plus size={16} />
              New project
            </Link>
          ) : undefined
        }
      />
      <div className="project-toolbar">
        <div className="status-tabs" role="tablist">
          {(["active", "archived", "trashed"] as Status[]).map((v) => (
            <button
              type="button"
              key={v}
              className={status === v ? "active" : ""}
              onClick={() => setStatus(v)}
            >
              {v === "trashed" ? "Trash" : label(v)}
            </button>
          ))}
        </div>
        <div className="project-search">
          <Search size={16} />
          <Input
            aria-label="Search projects"
            value={query}
            onChange={(e) => setQuery(e.target.value)}
            placeholder="Search projects"
          />
        </div>
      </div>
      {error ? (
        <div className="inline-alert" role="alert">
          {error}
          <Button variant="quiet" onClick={() => void load()}>
            Retry
          </Button>
        </div>
      ) : loading ? (
        <div className="project-grid">
          {[1, 2, 3].map((v) => (
            <div key={v} className="project-skeleton" />
          ))}
        </div>
      ) : items.length ? (
        <div className="project-grid">
          {items.map((p, i) => (
            <article className="project-card" key={p.id}>
              <Link href={`/app/${workspace.slug}/projects/${p.id}`}>
                <div className={`project-poster tone-${i % 4}`}>
                  <Film />
                  <span>
                    {p.settings.aspectWidth ?? 16}:{p.settings.aspectHeight ?? 9}
                  </span>
                </div>
                <div className="project-card-body">
                  <div>
                    <span className="privacy-chip">
                      <ShieldCheck size={12} /> Private
                    </span>
                    <h2>{p.settings.name}</h2>
                    <p>
                      {label(p.settings.productionFormat)} · {p.settings.frameRateNumerator ?? 24}{" "}
                      fps · v{p.currentVersion}
                    </p>
                  </div>
                </div>
              </Link>
              {status !== "active" && canEdit ? (
                <Button variant="quiet" onClick={() => void restore(p.id)}>
                  <ArchiveRestore size={15} />
                  Restore
                </Button>
              ) : null}
            </article>
          ))}
        </div>
      ) : (
        <EmptyState
          title={status === "active" ? "Create your first production" : `No ${status} projects`}
          description={
            status === "active"
              ? "Start blank or choose a production template. Every project is private by default."
              : "Projects moved here will appear with their recovery options."
          }
          action={
            status === "active" && canEdit ? (
              <Link className="button button-primary" href={`/app/${workspace.slug}/projects/new`}>
                Create project
              </Link>
            ) : undefined
          }
        />
      )}
    </>
  );
}
function label(v: string) {
  return v.replaceAll("_", " ").replace(/\b\w/g, (c) => c.toUpperCase());
}
