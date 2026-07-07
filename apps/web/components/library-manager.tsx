"use client";
import type { components } from "@cineforge/api-client";
import { apiFetch, apiUpload } from "@cineforge/api-client";
import { FileAudio, FileBox, FileImage, FileVideo, Plus, Upload } from "lucide-react";
import { useEffect, useState } from "react";
import { useShell } from "@/components/app-shell";
import { Button, EmptyState, Field, Input, PageHeader, Select } from "@/components/ui";
type Library = components["schemas"]["Library"];
type Entity = components["schemas"]["LibraryEntity"];
export function LibraryManager() {
  const { workspace } = useShell();
  const [libraries, setLibraries] = useState<Library[]>([]);
  const [selected, setSelected] = useState("");
  const [entities, setEntities] = useState<Entity[]>([]);
  const [showEntity, setShowEntity] = useState(false);
  useEffect(() => {
    if (!workspace) return;
    apiFetch<{ items: Library[] }>(`/v1/workspaces/${workspace.id}/libraries`).then(async (r) => {
      if (!r.items.length) {
        const l = await apiFetch<Library>(`/v1/workspaces/${workspace.id}/libraries`, {
          method: "POST",
          idempotencyKey: crypto.randomUUID() + crypto.randomUUID(),
          body: { name: "Story library", type: "story" },
        });
        r = { items: [l] };
      }
      setLibraries(r.items);
      setSelected(r.items[0]?.id ?? "");
    });
  }, [workspace]);
  // Selection is the sole input; reload is a local request helper.
  // biome-ignore lint/correctness/useExhaustiveDependencies: adding reload would recreate the effect on every render.
  useEffect(() => {
    if (selected) void reload();
  }, [selected]);
  async function reload() {
    const r = await apiFetch<{ items: Entity[] }>(`/v1/libraries/${selected}/entities`);
    setEntities(r.items);
  }
  async function createEntity(e: React.FormEvent<HTMLFormElement>) {
    e.preventDefault();
    const f = new FormData(e.currentTarget);
    await apiFetch(`/v1/libraries/${selected}/entities`, {
      method: "POST",
      idempotencyKey: crypto.randomUUID() + crypto.randomUUID(),
      body: {
        name: f.get("name"),
        type: f.get("type"),
        summary: f.get("summary"),
        description: "",
        aliases: [],
        tags: String(f.get("tags") ?? "")
          .split(",")
          .map((x) => x.trim())
          .filter(Boolean),
        attributes: {},
        changeSummary: "Initial draft",
      },
    });
    setShowEntity(false);
    await reload();
  }
  if (!workspace) return null;
  return (
    <>
      <PageHeader
        eyebrow="Shared workspace"
        title="Story library"
        description="Versioned characters, places, props, visual language, voice, and production references."
        action={
          <Button onClick={() => setShowEntity(true)}>
            <Plus size={16} />
            New entity
          </Button>
        }
      />
      <div className="mt-6 flex flex-wrap items-center gap-3">
        <label className="text-sm text-[var(--muted)]" htmlFor="library">
          Library
        </label>
        <select
          id="library"
          className="bg-transparent text-sm font-medium"
          value={selected}
          onChange={(e) => setSelected(e.target.value)}
        >
          {libraries.map((l) => (
            <option key={l.id} value={l.id}>
              {l.name}
            </option>
          ))}
        </select>
      </div>
      {showEntity ? (
        <form onSubmit={createEntity} className="my-8 grid max-w-xl gap-5 bg-[var(--surface)] p-6">
          <div className="flex items-center justify-between">
            <h2 className="m-0 text-lg font-semibold">New library entity</h2>
            <Button type="button" variant="quiet" onClick={() => setShowEntity(false)}>
              Cancel
            </Button>
          </div>
          <Field label="Name">
            <Input name="name" required autoFocus />
          </Field>
          <Field label="Type">
            <Select name="type">
              <option value="character">Character</option>
              <option value="location">Location</option>
              <option value="prop">Prop</option>
              <option value="faction">Faction</option>
              <option value="creature">Creature</option>
              <option value="vehicle">Vehicle</option>
              <option value="style_guide">Style guide</option>
              <option value="voice_profile">Voice profile</option>
              <option value="other">Other</option>
            </Select>
          </Field>
          <Field label="Summary">
            <Input name="summary" />
          </Field>
          <Field label="Tags" hint="Comma-separated">
            <Input name="tags" />
          </Field>
          <div>
            <Button>Create draft</Button>
          </div>
        </form>
      ) : null}
      {entities.length ? (
        <div className="mt-7 divide-y divide-[var(--line)]">
          {entities.map((e) => {
            const v = e.currentVersion;
            return (
              <div key={e.id} className="flex items-center justify-between gap-4 py-5">
                <div>
                  <div className="font-medium">{v?.name ?? "Untitled entity"}</div>
                  <div className="mt-1 text-sm capitalize text-[var(--muted)]">
                    {v?.type?.replaceAll("_", " ") ?? "entity"} · {v?.status ?? "draft"} · v
                    {v?.version ?? 1}
                  </div>
                </div>
              </div>
            );
          })}
        </div>
      ) : (
        <EmptyState
          title="Build reusable production truth"
          description="Create a character, location, prop, voice profile, or style guide. Project links remain pinned to approved versions."
        />
      )}
      <ReferenceUpload workspaceId={workspace.id} />
    </>
  );
}

function ReferenceUpload({ workspaceId }: { workspaceId: string }) {
  const [status, setStatus] = useState("");
  async function choose(file: File) {
    const kind = mediaKind(file);
    if (!kind) {
      setStatus("This format is not supported.");
      return;
    }
    setStatus("Authorizing upload…");
    try {
      const result = await apiFetch<{ asset: { id: string }; uploadUrl: string }>(
        `/v1/workspaces/${workspaceId}/media/uploads`,
        {
          method: "POST",
          idempotencyKey: crypto.randomUUID() + crypto.randomUUID(),
          body: {
            kind,
            filename: file.name,
            contentType: file.type || "application/octet-stream",
            sizeBytes: file.size,
          },
        },
      );
      setStatus("Uploading to quarantine…");
      await apiUpload(result.uploadUrl, file);
      setStatus("Upload complete. Security checks are queued…");
      await apiFetch(`/v1/media/uploads/${result.asset.id}/complete`, { method: "POST" });
      setStatus("Reference is being scanned and processed.");
    } catch (e) {
      setStatus(e instanceof Error ? e.message : "Upload could not be authorized");
    }
  }
  return (
    <section className="mt-14 border-t border-[var(--line)] pt-8">
      <div className="flex items-start justify-between gap-5">
        <div>
          <h2 className="m-0 text-lg font-semibold">Reference media</h2>
          <p className="mt-2 text-sm text-[var(--muted)]">
            Images, video, audio, voice, and binary glTF are quarantined before use.
          </p>
        </div>
        <label className="inline-flex min-h-10 cursor-pointer items-center gap-2 rounded-lg bg-[var(--surface-raised)] px-4 text-sm font-medium">
          <Upload size={16} />
          Choose file
          <input
            className="sr-only"
            type="file"
            accept="image/jpeg,image/png,image/webp,video/mp4,video/quicktime,video/webm,audio/*,.glb"
            onChange={(e) => {
              const f = e.target.files?.[0];
              if (f) void choose(f);
            }}
          />
        </label>
      </div>
      <div className="mt-6 grid grid-cols-4 gap-4 text-[var(--muted)]">
        <FileImage />
        <FileVideo />
        <FileAudio />
        <FileBox />
      </div>
      {status ? (
        <p role="status" className="mt-4 text-sm text-[var(--muted)]">
          {status}
        </p>
      ) : null}
    </section>
  );
}
function mediaKind(file: File) {
  if (file.name.toLowerCase().endsWith(".glb")) return "3d";
  if (file.type.startsWith("image/")) return "image";
  if (file.type.startsWith("video/")) return "video";
  if (file.type.startsWith("audio/")) return "audio";
  return null;
}
