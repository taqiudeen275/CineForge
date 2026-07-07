"use client";
import { apiFetch } from "@cineforge/api-client";
import { useEffect, useState } from "react";
import { useShell } from "@/components/app-shell";
import { Button, Field, Input, PageHeader, Select } from "@/components/ui";
import { SettingsNav } from "@/components/settings-nav";
type Member = { id: string; userId: string; role: string; status: string; canSpend: boolean };
export function MemberSettings() {
  const { workspace } = useShell();
  const [items, setItems] = useState<Member[]>([]);
  const [show, setShow] = useState(false);
  const [token, setToken] = useState("");
  // Workspace changes are the only reason to refresh this list.
  // biome-ignore lint/correctness/useExhaustiveDependencies: load is a local request helper.
  useEffect(() => {
    if (workspace) void load();
  }, [workspace]);
  async function load() {
    const r = await apiFetch<{ items: Member[] }>(`/v1/workspaces/${workspace?.id}/members`);
    setItems(r.items);
  }
  async function invite(e: React.FormEvent<HTMLFormElement>) {
    e.preventDefault();
    const f = new FormData(e.currentTarget);
    const r = await apiFetch<{ developmentToken: string }>(
      `/v1/workspaces/${workspace?.id}/invitations`,
      {
        method: "POST",
        idempotencyKey: crypto.randomUUID() + crypto.randomUUID(),
        body: { email: f.get("email"), role: f.get("role") },
      },
    );
    setToken(r.developmentToken);
    setShow(false);
  }
  if (!workspace) return null;
  return (
    <>
      <PageHeader
        eyebrow="Workspace settings"
        title="Members"
        description="Roles apply across every private project in this workspace."
        action={<Button onClick={() => setShow(true)}>Invite member</Button>}
      />
      <div className="mt-8 max-w-3xl">
        <SettingsNav slug={workspace.slug} />
        {show ? (
          <form
            onSubmit={invite}
            className="mb-8 grid gap-5 bg-[var(--surface)] p-6 sm:grid-cols-[1fr_180px_auto]"
          >
            <Field label="Email">
              <Input name="email" type="email" required />
            </Field>
            <Field label="Role">
              <Select name="role">
                <option value="editor">Editor</option>
                <option value="reviewer">Reviewer</option>
                <option value="viewer">Viewer</option>
              </Select>
            </Field>
            <div className="self-end">
              <Button>Send invite</Button>
            </div>
          </form>
        ) : null}
        {token ? (
          <p className="rounded-lg bg-[var(--surface-raised)] p-4 text-xs text-[var(--muted)]">
            Local development invite token: <code>{token}</code>
          </p>
        ) : null}
        <div className="divide-y divide-[var(--line)]">
          {items.map((m) => (
            <div key={m.id} className="flex justify-between py-5 text-sm">
              <span>{m.userId}</span>
              <span className="capitalize text-[var(--muted)]">{m.role}</span>
            </div>
          ))}
        </div>
      </div>
    </>
  );
}
