"use client";
import { apiFetch } from "@cineforge/api-client";
import type { components } from "@cineforge/api-client";
import { useEffect, useState } from "react";
import { useShell } from "@/components/app-shell";
import { Button, Field, Input, PageHeader } from "@/components/ui";
import { SettingsNav } from "@/components/settings-nav";
type Policy = components["schemas"]["BudgetPolicy"];
export function SpendSettings() {
  const { workspace, user } = useShell();
  const [p, setP] = useState<Policy>();
  const [msg, setMsg] = useState("");
  useEffect(() => {
    if (workspace) apiFetch<Policy>(`/v1/workspaces/${workspace.id}/budget-policy`).then(setP);
  }, [workspace]);
  if (!workspace || !p) return <p className="text-sm text-[var(--muted)]">Loading policy…</p>;
  const currentWorkspace = workspace;
  const policy = p;
  async function save(e: React.FormEvent<HTMLFormElement>) {
    e.preventDefault();
    const f = new FormData(e.currentTarget);
    const dollars = (name: string) => {
      const v = String(f.get(name) || "");
      return v ? Math.round(Number(v) * 1_000_000) : null;
    };
    const next = await apiFetch<Policy>(`/v1/workspaces/${currentWorkspace.id}/budget-policy`, {
      method: "PUT",
      etag: String(policy.version),
      body: {
        ...policy,
        monthlyLimitMicros: dollars("monthly"),
        perRunApprovalThresholdMicros: dollars("approval"),
        editorCanPublishLibrary: f.get("publish") === "on",
      },
    });
    setP(next);
    setMsg("Policy saved");
  }
  const owner = workspace.ownerUserId === user.id;
  return (
    <>
      <PageHeader
        eyebrow="Workspace settings"
        title="Spend policy"
        description="Controls are persisted now and will be consumed by generation and rendering later."
      />
      <div className="mt-8 max-w-3xl">
        <SettingsNav slug={workspace.slug} />
        <form onSubmit={save} className="grid gap-6">
          <div className="grid gap-6 sm:grid-cols-2">
            <Field label="Monthly limit (USD)" hint="Leave blank for no hard limit.">
              <Input
                name="monthly"
                type="number"
                min="0"
                step="0.01"
                defaultValue={p.monthlyLimitMicros ? String(p.monthlyLimitMicros / 1_000_000) : ""}
                disabled={!owner}
              />
            </Field>
            <Field
              label="Approval threshold (USD)"
              hint="Runs at or above this amount require approval."
            >
              <Input
                name="approval"
                type="number"
                min="0"
                step="0.01"
                defaultValue={
                  p.perRunApprovalThresholdMicros ? String(p.perRunApprovalThresholdMicros / 1_000_000) : ""
                }
                disabled={!owner}
              />
            </Field>
          </div>
          <label className="flex items-center gap-3 text-sm">
            <input
              name="publish"
              type="checkbox"
              defaultChecked={p.editorCanPublishLibrary}
              disabled={!owner}
            />
            Allow editors to approve workspace library versions
          </label>
          {owner ? (
            <div className="flex items-center gap-4">
              <Button>Save policy</Button>
              {msg ? <span className="text-sm text-[var(--success)]">{msg}</span> : null}
            </div>
          ) : (
            <p className="text-sm text-[var(--muted)]">
              Only the workspace owner can change spend policy.
            </p>
          )}
        </form>
      </div>
    </>
  );
}
