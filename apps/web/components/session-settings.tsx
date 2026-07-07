"use client";
import { apiFetch } from "@cineforge/api-client";
import { useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import { Button, PageHeader } from "@/components/ui";
type Session = {
  id: string;
  deviceLabel: string;
  createdAt: string;
  lastSeenAt: string;
  expiresAt: string;
};
export function SessionSettings() {
  const router = useRouter();
  const [items, setItems] = useState<Session[]>([]);
  // Active sessions load once on entry and again after explicit revocation.
  // biome-ignore lint/correctness/useExhaustiveDependencies: load is a local request helper.
  useEffect(() => {
    void load();
  }, []);
  async function load() {
    const r = await apiFetch<{ items: Session[] }>("/v1/auth/sessions");
    setItems(r.items);
  }
  async function revoke(id: string) {
    await apiFetch(`/v1/auth/sessions/${id}`, { method: "DELETE", body: {} });
    await load();
  }
  async function all() {
    await apiFetch("/v1/auth/logout-all", { method: "POST", body: {} });
    router.push("/sign-in");
  }
  return (
    <>
      <PageHeader
        eyebrow="Account security"
        title="Active sessions"
        description="Revoke browsers you no longer recognize."
        action={
          <Button variant="danger" onClick={all}>
            Sign out everywhere
          </Button>
        }
      />
      <div className="mt-8 divide-y divide-[var(--line)]">
        {items.map((s) => (
          <div key={s.id} className="flex items-center justify-between gap-4 py-5">
            <div>
              <div className="text-sm font-medium">{s.deviceLabel}</div>
              <div className="mt-1 text-xs text-[var(--muted)]">
                Last active {new Date(s.lastSeenAt).toLocaleString()} · Expires{" "}
                {new Date(s.expiresAt).toLocaleDateString()}
              </div>
            </div>
            <Button variant="quiet" onClick={() => revoke(s.id)}>
              Revoke
            </Button>
          </div>
        ))}
      </div>
    </>
  );
}
