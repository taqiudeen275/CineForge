"use client";
import type { components } from "@cineforge/api-client";
import { apiFetch } from "@cineforge/api-client";
import { ArchiveRestore, ShieldCheck, Trash2 } from "lucide-react";
import { useRouter } from "next/navigation";
import { useEffect, useState } from "react";
import { useShell } from "@/components/app-shell";
import { Button, Field, Input, PageHeader } from "@/components/ui";
import { SettingsNav } from "@/components/settings-nav";
type Workspace = components["schemas"]["Workspace"];
export function WorkspaceGeneral() {
  const { workspace, user, canAdmin, refresh } = useShell();
  const router = useRouter();
  const [busy, setBusy] = useState(false);
  const [msg, setMsg] = useState("");
  const [deleted, setDeleted] = useState<Workspace[]>([]);
  useEffect(() => {
    apiFetch<{ items: Workspace[] }>("/v1/workspaces?status=trashed")
      .then((v) => setDeleted(v.items))
      .catch(() => undefined);
  }, []);
  if (!workspace) return null;
  const currentWorkspace = workspace;
  const owner = currentWorkspace.currentMembership?.role === "owner";
  async function save(e: React.FormEvent<HTMLFormElement>) {
    e.preventDefault();
    const name = String(new FormData(e.currentTarget).get("name"));
    const out = await apiFetch<Workspace>(`/v1/workspaces/${currentWorkspace.id}`, {
      method: "PATCH",
      etag: String(currentWorkspace.version),
      body: { name },
    });
    setMsg("Workspace updated");
    await refresh();
    if (out.slug !== currentWorkspace.slug) router.replace(`/app/${out.slug}/settings/general`);
  }
  async function remove() {
    if (!confirm("Move this workspace to trash for 30 days?")) return;
    setBusy(true);
    try {
      await apiFetch(`/v1/workspaces/${currentWorkspace.id}`, {
        method: "DELETE",
        idempotencyKey: crypto.randomUUID() + crypto.randomUUID(),
        body: {},
      });
      router.push("/app");
      await refresh();
    } finally {
      setBusy(false);
    }
  }
  async function restore(id: string) {
    await apiFetch(`/v1/workspaces/${id}/restore`, { method: "POST", body: {} });
    setDeleted((v) => v.filter((w) => w.id !== id));
    await refresh();
  }
  return (
    <>
      <PageHeader
        eyebrow="Workspace settings"
        title={workspace.name}
        description={`${workspace.kind === "personal" ? "Personal" : "Team"} workspace · ${workspace.currentMembership?.role ?? "member"}`}
      />
      <div className="mt-8 max-w-3xl">
        <SettingsNav slug={workspace.slug} />
        <section className="settings-card">
          <div className="settings-title">
            <div>
              <h2>Workspace profile</h2>
              <p>The shared identity and privacy boundary for your productions.</p>
            </div>
            <span className="privacy-chip">
              <ShieldCheck size={12} />
              Private
            </span>
          </div>
          <form onSubmit={save} className="grid gap-5">
            <Field label="Workspace name">
              <Input name="name" defaultValue={workspace.name} disabled={!canAdmin} />
            </Field>
            <div className="form-grid">
              <Field label="Workspace slug">
                <Input value={workspace.slug} disabled readOnly />
              </Field>
              <Field label="Owner">
                <Input
                  value={workspace.ownerUserId === user.id ? "You" : "Workspace owner"}
                  disabled
                  readOnly
                />
              </Field>
            </div>
            {canAdmin ? (
              <div className="save-row">
                <Button>Save workspace</Button>
                {msg ? <span>{msg}</span> : null}
              </div>
            ) : (
              <p className="muted text-sm">Only owners and admins can edit workspace details.</p>
            )}
          </form>
        </section>
        {deleted.length ? (
          <section className="settings-card mt-5">
            <div className="settings-title">
              <div>
                <h2>Recently deleted workspaces</h2>
                <p>Restore an owned workspace before its recovery window ends.</p>
              </div>
            </div>
            {deleted.map((w) => (
              <div key={w.id} className="deleted-row">
                <div>
                  <strong>{w.name}</strong>
                  <small>
                    Purges {w.purgeAt ? new Date(w.purgeAt).toLocaleDateString() : "after 30 days"}
                  </small>
                </div>
                <Button variant="quiet" onClick={() => restore(w.id)}>
                  <ArchiveRestore size={15} />
                  Restore
                </Button>
              </div>
            ))}
          </section>
        ) : null}
        {owner ? (
          <section className="danger-card">
            <Trash2 />
            <div>
              <h2>Move workspace to trash</h2>
              <p>
                Projects remain recoverable for 30 days. High-risk actions may require fresh
                authentication.
              </p>
            </div>
            <Button variant="danger" disabled={busy} onClick={remove}>
              Move to trash
            </Button>
          </section>
        ) : null}
      </div>
    </>
  );
}
