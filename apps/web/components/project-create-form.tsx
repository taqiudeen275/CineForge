"use client";
import { apiFetch } from "@cineforge/api-client";
import type { components } from "@cineforge/api-client";
import { useRouter } from "next/navigation";
import { useState } from "react";
import { Button, Field, Input, PageHeader, Select } from "@/components/ui";
import { useShell } from "@/components/app-shell";
type Project = components["schemas"]["Project"];
export function ProjectCreateForm() {
  const { workspace } = useShell();
  const router = useRouter();
  const [name, setName] = useState("");
  const [type, setType] = useState<"single" | "series">("single");
  const [format, setFormat] = useState("short_film");
  const [error, setError] = useState("");
  const [busy, setBusy] = useState(false);
  async function submit(e: React.FormEvent) {
    e.preventDefault();
    if (!workspace) return;
    setBusy(true);
    setError("");
    try {
      const p = await apiFetch<Project>(`/v1/workspaces/${workspace.id}/projects`, {
        method: "POST",
        idempotencyKey: crypto.randomUUID() + crypto.randomUUID(),
        body: { name, projectType: type, productionFormat: format },
      });
      router.push(`/app/${workspace.slug}/projects/${p.id}`);
    } catch (e) {
      setError(e instanceof Error ? e.message : "Could not create project");
    } finally {
      setBusy(false);
    }
  }
  return (
    <div className="max-w-2xl">
      <PageHeader
        eyebrow="New project"
        title="Start with the shape of the story"
        description="Production defaults can be refined immediately after creation."
      />
      <form onSubmit={submit} className="mt-9 grid gap-6">
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
        <div className="grid gap-6 sm:grid-cols-2">
          <Field label="Project type">
            <Select value={type} onChange={(e) => setType(e.target.value as "single" | "series")}>
              <option value="single">Single production</option>
              <option value="series">Series</option>
            </Select>
          </Field>
          <Field label="Production format">
            <Select value={format} onChange={(e) => setFormat(e.target.value)}>
              <option value="short_film">Short film</option>
              <option value="feature">Feature film</option>
              <option value="episodic">Episodic</option>
              <option value="music_video">Music video</option>
              <option value="advertisement">Advertisement</option>
              <option value="trailer">Trailer</option>
              <option value="other">Other</option>
            </Select>
          </Field>
        </div>
        {error ? (
          <p role="alert" className="text-sm text-[var(--danger)]">
            {error}
          </p>
        ) : null}
        <div>
          <Button disabled={busy}>{busy ? "Creating…" : "Create project"}</Button>
        </div>
      </form>
    </div>
  );
}
