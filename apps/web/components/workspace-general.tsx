"use client";
import { apiFetch } from "@cineforge/api-client";
import { useRouter } from "next/navigation";
import { useState } from "react";
import { useShell } from "@/components/app-shell";
import { Button, PageHeader } from "@/components/ui";
import { SettingsNav } from "@/components/settings-nav";
export function WorkspaceGeneral() {
  const { workspace, user } = useShell();
  const router = useRouter();
  const [busy, setBusy] = useState(false);
  if (!workspace) return null;
  const currentWorkspace = workspace;
  async function remove() {
    if (!confirm("Move this workspace to trash for 30 days?")) return;
    setBusy(true);
    await apiFetch(`/v1/workspaces/${currentWorkspace.id}`, {
      method: "DELETE",
      idempotencyKey: crypto.randomUUID() + crypto.randomUUID(),
      body: {},
    });
    router.push("/app");
  }
  return (
    <>
      <PageHeader
        eyebrow="Workspace settings"
        title={workspace.name}
        description={`${workspace.kind === "personal" ? "Personal" : "Team"} workspace · ${workspace.slug}`}
      />
      <div className="mt-8 max-w-3xl">
        <SettingsNav slug={workspace.slug} />
        <dl className="grid gap-5 text-sm sm:grid-cols-2">
          <div>
            <dt className="text-[var(--muted)]">Privacy</dt>
            <dd className="ml-0 mt-1 font-medium">Private by default</dd>
          </div>
          <div>
            <dt className="text-[var(--muted)]">Owner</dt>
            <dd className="ml-0 mt-1 font-medium">
              {workspace.ownerUserId === user.id ? "You" : "Workspace owner"}
            </dd>
          </div>
        </dl>
        {workspace.ownerUserId === user.id ? (
          <section className="mt-14 border-t border-[var(--line)] pt-7">
            <h2 className="m-0 text-base font-semibold text-[var(--danger)]">
              Move workspace to trash
            </h2>
            <p className="max-w-xl text-sm leading-6 text-[var(--muted)]">
              The workspace can be restored for 30 days before permanent deletion is queued.
            </p>
            <Button variant="danger" disabled={busy} onClick={remove}>
              Move to trash
            </Button>
          </section>
        ) : null}
      </div>
    </>
  );
}
