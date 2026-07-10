"use client";
import { apiFetch } from "@cineforge/api-client";
import type { components } from "@cineforge/api-client";
import { Check, Film, LayoutTemplate } from "lucide-react";
import { useRouter } from "next/navigation";
import { useEffect, useState } from "react";
import { Button, Field, Input, PageHeader } from "@/components/ui";
import { useShell } from "@/components/app-shell";
type Project = components["schemas"]["Project"];
type Template = components["schemas"]["ProjectTemplate"];
export function ProjectCreateForm() {
  const { workspace, canEdit } = useShell();
  const router = useRouter();
  const [name, setName] = useState("");
  const [templates, setTemplates] = useState<Template[]>([]);
  const [selected, setSelected] = useState<string | undefined>();
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState("");
  useEffect(() => {
    if (workspace)
      apiFetch<{ items: Template[] }>(`/v1/project-templates?workspaceId=${workspace.id}`)
        .then((v) => setTemplates(v.items))
        .catch((e) => setError(e.message));
  }, [workspace]);
  async function submit(e: React.FormEvent) {
    e.preventDefault();
    if (!workspace) return;
    setBusy(true);
    setError("");
    try {
      const p = await apiFetch<Project>(`/v1/workspaces/${workspace.id}/projects`, {
        method: "POST",
        idempotencyKey: crypto.randomUUID() + crypto.randomUUID(),
        body: { name, templateId: selected ?? null },
      });
      router.push(`/app/${workspace.slug}/projects/${p.id}`);
    } catch (e) {
      setError(e instanceof Error ? e.message : "Could not create project");
    } finally {
      setBusy(false);
    }
  }
  if (!workspace) return null;
  if (!canEdit)
    return <div className="inline-alert">Your role can view projects but cannot create them.</div>;
  return (
    <>
      <PageHeader
        eyebrow="Quick create"
        title="Start a new production"
        description="Name the story, choose a starting point, and refine every production setting after creation."
      />
      <form onSubmit={submit} className="quick-create">
        <Field label="Project name">
          <Input
            autoFocus
            value={name}
            onChange={(e) => setName(e.target.value)}
            placeholder="Untitled story"
            required
            maxLength={160}
          />
        </Field>
        <fieldset>
          <legend>Starting point</legend>
          <div className="template-grid">
            <button
              type="button"
              className={!selected ? "selected" : ""}
              onClick={() => setSelected(undefined)}
            >
              <span>
                <Film />
              </span>
              <strong>Blank project</strong>
              <small>16:9 · 24 fps · Balanced</small>
              {!selected ? <Check /> : null}
            </button>
            {templates.map((t) => (
              <button
                type="button"
                key={t.id}
                className={selected === t.id ? "selected" : ""}
                onClick={() => setSelected(t.id)}
              >
                <span>
                  <LayoutTemplate />
                </span>
                <strong>{t.name}</strong>
                <small>
                  {t.description ||
                    `${t.settings.aspectWidth ?? 16}:${t.settings.aspectHeight ?? 9} production`}
                </small>
                {selected === t.id ? <Check /> : null}
              </button>
            ))}
          </div>
        </fieldset>
        {error ? (
          <p className="auth-error" role="alert">
            {error}
          </p>
        ) : null}
        <div className="quick-actions">
          <Button type="button" variant="quiet" onClick={() => router.back()}>
            Cancel
          </Button>
          <Button disabled={busy}>{busy ? "Creating…" : "Create private project"}</Button>
        </div>
      </form>
    </>
  );
}
