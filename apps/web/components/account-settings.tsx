"use client";
import { apiFetch } from "@cineforge/api-client";
import { useState } from "react";
import { useRouter } from "next/navigation";
import { useShell } from "@/components/app-shell";
import { Button, Field, Input, PageHeader } from "@/components/ui";
import { MfaSettings } from "@/components/mfa-settings";
export function AccountSettings() {
  const { user } = useShell();
  const router = useRouter();
  const [msg, setMsg] = useState("");
  async function save(e: React.FormEvent<HTMLFormElement>) {
    e.preventDefault();
    const f = new FormData(e.currentTarget);
    await apiFetch("/v1/me", {
      method: "PATCH",
      etag: String(user.version),
      body: {
        displayName: f.get("name"),
        avatarUrl: f.get("avatar") || null,
        locale: f.get("locale"),
        timezone: f.get("timezone"),
      },
    });
    setMsg("Profile saved");
  }
  async function remove() {
    if (
      !confirm(
        "Schedule account deletion in 30 days? You must first transfer ownership of team workspaces.",
      )
    )
      return;
    await apiFetch("/v1/me/deletion", {
      method: "POST",
      idempotencyKey: crypto.randomUUID() + crypto.randomUUID(),
      body: {},
    });
    router.push("/sign-in");
  }
  return (
    <>
      <PageHeader eyebrow="Account" title="Profile and security" description={user.email} />
      <form onSubmit={save} className="mt-9 grid max-w-2xl gap-6">
        <Field label="Display name">
          <Input name="name" defaultValue={user.displayName} />
        </Field>
        <Field label="Avatar URL">
          <Input name="avatar" type="url" defaultValue={user.avatarUrl ?? ""} />
        </Field>
        <div className="grid gap-6 sm:grid-cols-2">
          <Field label="Locale">
            <Input name="locale" defaultValue={user.locale} />
          </Field>
          <Field label="Timezone">
            <Input name="timezone" defaultValue={user.timezone} />
          </Field>
        </div>
        <div className="flex items-center gap-4">
          <Button>Save profile</Button>
          {msg ? <span className="text-sm text-[var(--success)]">{msg}</span> : null}
        </div>
      </form>
      <section className="mt-14 border-t border-[var(--line)] pt-7">
        <h2 className="text-base font-semibold">Sessions</h2>
        <a className="text-sm text-[var(--accent)]" href="/app/settings/sessions">
          Review active sessions
        </a>
      </section>
      <MfaSettings />
      <section className="mt-14 border-t border-[var(--line)] pt-7">
        <h2 className="text-base font-semibold text-[var(--danger)]">Delete account</h2>
        <p className="max-w-xl text-sm leading-6 text-[var(--muted)]">
          Deletion has a 30-day recovery period and cannot orphan a team workspace.
        </p>
        <Button variant="danger" onClick={remove}>
          Schedule deletion
        </Button>
      </section>
    </>
  );
}
