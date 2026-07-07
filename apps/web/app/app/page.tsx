"use client";
import { useEffect } from "react";
import { useRouter } from "next/navigation";
import { useShell } from "@/components/app-shell";
export default function AppIndex() {
  const { workspace } = useShell();
  const router = useRouter();
  useEffect(() => {
    if (workspace) router.replace(`/app/${workspace.slug}/projects`);
  }, [workspace, router]);
  return <div className="py-20 text-center text-sm text-[var(--muted)]">Opening projects…</div>;
}
