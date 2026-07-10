"use client";
import { apiFetch, ApiError } from "@cineforge/api-client";
import type { components } from "@cineforge/api-client";
import { Archive, Clock3, Copy, FileStack, RotateCcw, Trash2 } from "lucide-react";
import { useRouter } from "next/navigation";
import { useEffect, useState } from "react";
import { Button, Field, Input, PageHeader, Select } from "@/components/ui";
import { useShell } from "@/components/app-shell";
import { formValuesFromFormData, serializeProjectSettings } from "@/lib/project-settings";
type Project = components["schemas"]["Project"];
type Version = components["schemas"]["ProjectVersion"];
export function ProjectSettings({
  projectId,
  workspaceSlug,
}: {
  projectId: string;
  workspaceSlug: string;
}) {
  const { workspace, canEdit } = useShell();
  const router = useRouter();
  const [p, setProject] = useState<Project>();
  const [versions, setVersions] = useState<Version[]>([]);
  const [message, setMessage] = useState("");
  const [error, setError] = useState("");
  const [busy, setBusy] = useState(false);
  // Project changes only when its route id changes.
  // biome-ignore lint/correctness/useExhaustiveDependencies: load is the request helper for projectId.
  useEffect(() => {
    void load();
  }, [projectId]);
  async function load() {
    setError("");
    try {
      const out = await apiFetch<Project>(`/v1/projects/${projectId}`);
      if (!workspace || workspace.slug !== workspaceSlug || out.workspaceId !== workspace.id) {
        setError("This project does not belong to the workspace in this URL.");
        return;
      }
      setProject(out);
      const history = await apiFetch<{ items: Version[] }>(`/v1/projects/${projectId}/versions`);
      setVersions(history.items);
    } catch (e) {
      setError(
        e instanceof ApiError && e.status === 403
          ? "You do not have access to this project."
          : e instanceof Error
            ? e.message
            : "Project not found",
      );
    }
  }
  if (error)
    return (
      <div className="shell-message min-h-[70vh]">
        <h1>Project unavailable</h1>
        <p>{error}</p>
        <Button onClick={() => router.push(`/app/${workspace?.slug}/projects`)}>
          Back to projects
        </Button>
      </div>
    );
  if (!p) return <div className="project-skeleton" />;
  const project = p,
    s = project.settings;
  async function save(e: React.FormEvent<HTMLFormElement>) {
    e.preventDefault();
    if (!canEdit) return;
    setBusy(true);
    setError("");
    try {
      const body = serializeProjectSettings(
        s,
        formValuesFromFormData(new FormData(e.currentTarget)),
      );
      const out = await apiFetch<Project>(`/v1/projects/${projectId}`, {
        method: "PUT",
        etag: String(project.currentVersion),
        body,
      });
      setProject(out);
      setMessage("Settings saved as a new project version");
      await load();
    } catch (e) {
      setError(
        e instanceof ApiError && e.status === 412
          ? "This project changed elsewhere. Reload and review before saving."
          : e instanceof Error
            ? e.message
            : "Could not save",
      );
    } finally {
      setBusy(false);
    }
  }
  async function duplicate() {
    if (!workspace) return;
    const name = prompt("Name the duplicated project", `${s.name} copy`);
    if (!name) return;
    const copy = await apiFetch<Project>(`/v1/projects/${projectId}/duplicate`, {
      method: "POST",
      idempotencyKey: crypto.randomUUID() + crypto.randomUUID(),
      body: { name },
    });
    router.push(`/app/${workspace.slug}/projects/${copy.id}`);
  }
  async function saveTemplate() {
    if (!workspace) return;
    const name = prompt("Template name", `${s.name} template`);
    if (!name) return;
    await apiFetch(`/v1/workspaces/${workspace.id}/project-templates`, {
      method: "POST",
      idempotencyKey: crypto.randomUUID() + crypto.randomUUID(),
      body: { name, description: `Created from ${s.name}`, projectId },
    });
    setMessage("Workspace template created");
  }
  async function action(path: string) {
    if (!confirm(path === "" ? "Move this project to trash for 30 days?" : "Archive this project?"))
      return;
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
        eyebrow={`Private ${project.status} project`}
        title={s.name}
        description={`${s.projectType === "series" ? "Series" : "Single production"} · Version ${project.currentVersion}`}
        action={
          canEdit ? (
            <div className="header-actions">
              <Button variant="quiet" onClick={duplicate}>
                <Copy size={16} />
                Duplicate
              </Button>
            </div>
          ) : undefined
        }
      />
      <div className="project-detail-grid">
        <form onSubmit={save} className="settings-card">
          <div className="settings-title">
            <div>
              <h2>Production settings</h2>
              <p>These defaults become the foundation for later story and generation tools.</p>
            </div>
            {!canEdit ? <span className="readonly-chip">Read only</span> : null}
          </div>
          <div className="form-grid">
            <Field label="Name">
              <Input name="name" defaultValue={s.name} required disabled={!canEdit} />
            </Field>
            <Field label="Project type">
              <Select name="projectType" defaultValue={s.projectType} disabled={!canEdit}>
                <option value="single">Single production</option>
                <option value="series">Series</option>
              </Select>
            </Field>
            <Field label="Production format">
              <Select name="productionFormat" defaultValue={s.productionFormat} disabled={!canEdit}>
                {[
                  "short_film",
                  "feature",
                  "episodic",
                  "music_video",
                  "advertisement",
                  "trailer",
                  "other",
                ].map((v) => (
                  <option value={v} key={v}>
                    {label(v)}
                  </option>
                ))}
              </Select>
            </Field>
            <Field label="Aspect ratio">
              <Select
                name="aspect"
                defaultValue={`${s.aspectWidth ?? 16}:${s.aspectHeight ?? 9}`}
                disabled={!canEdit}
              >
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
                disabled={!canEdit}
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
              <Input name="language" defaultValue={s.audioLanguage ?? "en"} disabled={!canEdit} />
            </Field>
            <Field label="Content rating">
              <Select name="rating" defaultValue={s.rating ?? "moderate"} disabled={!canEdit}>
                <option value="family">Family</option>
                <option value="moderate">Moderate</option>
                <option value="mature">Mature</option>
              </Select>
            </Field>
            <Field label="Default quality">
              <Select
                name="quality"
                defaultValue={s.qualityPolicy ?? "balanced"}
                disabled={!canEdit}
              >
                <option value="draft">Draft</option>
                <option value="balanced">Balanced</option>
                <option value="final">Final</option>
              </Select>
            </Field>
            <Field
              label="Project cost ceiling (USD)"
              hint="Optional ceiling for production work in this project."
            >
              <Input
                name="costCeiling"
                type="number"
                min="0"
                step="0.01"
                defaultValue={s.costCeilingMicros ? String(s.costCeilingMicros / 1_000_000) : ""}
                disabled={!canEdit}
              />
            </Field>
          </div>
          <Field
            label="Style direction"
            hint="The visual language inherited by later story and generation tools."
          >
            <textarea
              name="style"
              defaultValue={s.styleDirection ?? ""}
              disabled={!canEdit}
              maxLength={2000}
            />
          </Field>
          {error ? <p className="auth-error">{error}</p> : null}
          {canEdit ? (
            <div className="save-row">
              <Button disabled={busy}>{busy ? "Saving…" : "Save new version"}</Button>
              {message ? <span role="status">{message}</span> : null}
            </div>
          ) : null}
        </form>
        <aside className="project-side">
          <section>
            <div className="side-title">
              <Clock3 />
              <h2>Version history</h2>
            </div>
            <div className="version-list">
              {versions.map((v, i) => (
                <div key={v.id}>
                  <span>v{v.version}</span>
                  <div>
                    <strong>{v.settings.name}</strong>
                    <small>
                      {i === 0 ? "Current · " : ""}
                      {new Date(v.createdAt).toLocaleDateString()}
                    </small>
                  </div>
                </div>
              ))}
            </div>
          </section>
          {canEdit ? (
            <section>
              <div className="side-title">
                <FileStack />
                <h2>Project actions</h2>
              </div>
              <div className="action-list">
                <Button variant="quiet" onClick={saveTemplate}>
                  <FileStack size={15} />
                  Save as template
                </Button>
                <Button variant="quiet" onClick={() => action("/archive")}>
                  <Archive size={15} />
                  Archive project
                </Button>
                <Button variant="quiet" onClick={() => action("")}>
                  <Trash2 size={15} />
                  Move to trash
                </Button>
              </div>
              <p>
                <RotateCcw size={13} /> Trashed projects remain recoverable for 30 days.
              </p>
            </section>
          ) : null}
        </aside>
      </div>
    </>
  );
}
function label(v: string) {
  return v.replaceAll("_", " ").replace(/\b\w/g, (c) => c.toUpperCase());
}
