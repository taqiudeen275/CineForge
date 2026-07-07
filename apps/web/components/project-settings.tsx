"use client";
import { apiFetch } from "@cineforge/api-client";
import type { components } from "@cineforge/api-client";
import { Archive, Copy, Trash2 } from "lucide-react";
import { useRouter } from "next/navigation";
import { useEffect, useState } from "react";
import { Button, Field, Input, PageHeader, Select } from "@/components/ui";
import { useShell } from "@/components/app-shell";
import { formValuesFromFormData, serializeProjectSettings } from "@/lib/project-settings";
type Project = components["schemas"]["Project"];
export function ProjectSettings({ projectId }: { projectId: string }) {
  const { workspace } = useShell();
  const router = useRouter();
  const [p, setProject] = useState<Project>();
  const [message, setMessage] = useState("");
  useEffect(() => {
    apiFetch<Project>(`/v1/projects/${projectId}`).then(setProject);
  }, [projectId]);
  if (!p) return <p className="text-sm text-[var(--muted)]">Loading project…</p>;
  const project = p;
  const s = project.settings;
  async function save(e: React.FormEvent<HTMLFormElement>) {
    e.preventDefault();
    const form = new FormData(e.currentTarget);
    const body = serializeProjectSettings(s, formValuesFromFormData(form));
    const out = await apiFetch<Project>(`/v1/projects/${projectId}`, {
      method: "PUT",
      etag: String(project.currentVersion),
      body,
    });
    setProject(out);
    setMessage("Settings saved");
  }
  async function duplicate() {
    if (!workspace) return;
    const copy = await apiFetch<Project>(`/v1/projects/${projectId}/duplicate`, {
      method: "POST",
      idempotencyKey: crypto.randomUUID() + crypto.randomUUID(),
      body: { name: `${s.name} copy` },
    });
    router.push(`/app/${workspace.slug}/projects/${copy.id}`);
  }
  async function action(path: string) {
    await apiFetch(`/v1/projects/${projectId}${path}`, {
      method: path === "" ? "DELETE" : "POST",
      idempotencyKey: path === "" ? crypto.randomUUID() + crypto.randomUUID() : undefined,
      body: {},
    });
    router.push(`/app/${workspace?.slug}/projects`);
  }
  return (
    <>
      <PageHeader
        eyebrow="Private project"
        title={s.name}
        description={`${s.projectType === "series" ? "Series" : "Single production"} · Version ${p.currentVersion}`}
        action={
          <div className="flex gap-2">
            <Button variant="quiet" onClick={duplicate}>
              <Copy size={16} />
              Duplicate
            </Button>
          </div>
        }
      />
      <form onSubmit={save} className="mt-9 grid max-w-3xl gap-7">
        <div className="grid gap-6 sm:grid-cols-2">
          <Field label="Name">
            <Input name="name" defaultValue={s.name} required />
          </Field>
          <Field label="Aspect ratio">
            <Select name="aspect" defaultValue={`${s.aspectWidth ?? 16}:${s.aspectHeight ?? 9}`}>
              <option>16:9</option>
              <option>9:16</option>
              <option>21:9</option>
              <option>4:3</option>
              <option>1:1</option>
            </Select>
          </Field>
          <Field label="Frame rate">
            <Select
              name="fps"
              defaultValue={`${s.frameRateNumerator ?? 24}/${s.frameRateDenominator ?? 1}`}
            >
              <option value="12/1">12 fps</option>
              <option value="24000/1001">23.976 fps</option>
              <option value="24/1">24 fps</option>
              <option value="25/1">25 fps</option>
              <option value="30000/1001">29.97 fps</option>
              <option value="30/1">30 fps</option>
            </Select>
          </Field>
          <Field label="Audio language">
            <Input name="language" defaultValue={s.audioLanguage ?? "en"} />
          </Field>
          <Field label="Rating">
            <Select name="rating" defaultValue={s.rating ?? "moderate"}>
              <option value="family">Family</option>
              <option value="moderate">Moderate</option>
              <option value="mature">Mature</option>
            </Select>
          </Field>
          <Field label="Default quality">
            <Select name="quality" defaultValue={s.qualityPolicy ?? "balanced"}>
              <option value="draft">Draft</option>
              <option value="balanced">Balanced</option>
              <option value="final">Final</option>
            </Select>
          </Field>
        </div>
        <Field
          label="Style direction"
          hint="A concise visual language inherited by later story and generation tools."
        >
          <textarea
            name="style"
            defaultValue={s.styleDirection ?? ""}
            className="min-h-32 rounded-lg border border-[var(--line)] bg-[var(--surface)] p-3 text-sm"
            maxLength={2000}
          />
        </Field>
        <div className="flex items-center gap-4">
          <Button type="submit">Save settings</Button>
          {message ? <span className="text-sm text-[var(--success)]">{message}</span> : null}
        </div>
      </form>
      <section className="mt-16 border-t border-[var(--line)] pt-8">
        <h2 className="text-base font-semibold">Project lifecycle</h2>
        <div className="mt-4 flex flex-wrap gap-2">
          <Button variant="quiet" onClick={() => action("/archive")}>
            <Archive size={16} />
            Archive
          </Button>
          <Button variant="quiet" onClick={() => action("")}>
            <Trash2 size={16} />
            Move to trash
          </Button>
        </div>
      </section>
    </>
  );
}
