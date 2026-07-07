"use client";

import type { components } from "@cineforge/api-client";
import { apiFetch, ApiError } from "@cineforge/api-client";
import { BookOpen, FolderKanban, Settings, UserRound } from "lucide-react";
import Link from "next/link";
import { usePathname, useRouter } from "next/navigation";
import { createContext, useContext, useEffect, useMemo, useState } from "react";
import { Brand } from "@/components/brand";

type Workspace = components["schemas"]["Workspace"];
type User = components["schemas"]["User"];
type ShellState = { user: User; workspaces: Workspace[]; workspace?: Workspace };
const ShellContext = createContext<ShellState | undefined>(undefined);
export function useShell() {
  const value = useContext(ShellContext);
  if (!value) throw new Error("App shell not ready");
  return value;
}

export function AppShell({ children }: { children: React.ReactNode }) {
  const router = useRouter();
  const pathname = usePathname();
  const [user, setUser] = useState<User>();
  const [workspaces, setWorkspaces] = useState<Workspace[]>([]);
  const [loading, setLoading] = useState(true);
  // The initial shell bootstrap intentionally runs once for the mounted session.
  // biome-ignore lint/correctness/useExhaustiveDependencies: rerunning on router identity changes would duplicate bootstrap.
  useEffect(() => {
    void load();
  }, []);
  async function load() {
    try {
      const me = await apiFetch<User>("/v1/me");
      if (!me.emailVerified) {
        router.replace("/verify-email");
        return;
      }
      let list = await apiFetch<{ items: Workspace[] }>("/v1/workspaces");
      if (!list.items.length) {
        const ws = await apiFetch<Workspace>("/v1/workspaces/bootstrap", {
          method: "POST",
          body: {},
        });
        list = { items: [ws] };
      }
      setUser(me);
      setWorkspaces(list.items);
    } catch (e) {
      if (e instanceof ApiError && e.status === 401) router.replace("/sign-in");
      else console.error(e);
    } finally {
      setLoading(false);
    }
  }
  const slug = pathname.split("/")[2];
  const workspace = workspaces.find((w) => w.slug === slug) ?? workspaces[0];
  const value = useMemo(
    () => (user ? { user, workspaces, workspace } : undefined),
    [user, workspaces, workspace],
  );
  if (loading || !value)
    return (
      <div className="grid min-h-screen place-items-center text-sm text-[var(--muted)]">
        Loading workspace…
      </div>
    );
  const base = `/app/${workspace?.slug}`;
  return (
    <ShellContext.Provider value={value}>
      <div className="grid min-h-screen grid-cols-1 md:grid-cols-[232px_1fr]">
        <aside className="hidden border-r border-[var(--line)] bg-[var(--surface)] px-5 py-7 md:flex md:flex-col">
          <Brand />
          <select
            aria-label="Workspace"
            className="mt-10 w-full bg-transparent text-sm font-medium"
            value={workspace?.slug}
            onChange={(e) => router.push(`/app/${e.target.value}/projects`)}
          >
            {workspaces.map((w) => (
              <option key={w.id} value={w.slug}>
                {w.name}
              </option>
            ))}
          </select>
          <nav className="mt-8 grid gap-1">
            <Nav
              href={`${base}/projects`}
              active={pathname.includes("/projects")}
              icon={<FolderKanban size={17} />}
            >
              Projects
            </Nav>
            <Nav
              href={`${base}/library`}
              active={pathname.includes("/library")}
              icon={<BookOpen size={17} />}
            >
              Library
            </Nav>
            <Nav
              href={`${base}/settings/general`}
              active={pathname.includes("/settings")}
              icon={<Settings size={17} />}
            >
              Workspace settings
            </Nav>
          </nav>
          <div className="mt-auto">
            <Nav
              href="/app/settings/account"
              active={pathname.includes("/settings/account")}
              icon={<UserRound size={17} />}
            >
              {value.user.displayName}
            </Nav>
          </div>
        </aside>
        <main className="min-w-0">
          <div className="mx-auto max-w-6xl px-5 py-8 sm:px-8 sm:py-12">{children}</div>
        </main>
      </div>
    </ShellContext.Provider>
  );
}
function Nav({
  href,
  active,
  icon,
  children,
}: {
  href: string;
  active: boolean;
  icon: React.ReactNode;
  children: React.ReactNode;
}) {
  return (
    <Link
      href={href}
      className={`flex items-center gap-3 rounded-lg px-3 py-2.5 text-sm ${active ? "bg-[var(--surface-raised)] text-[var(--text)]" : "text-[var(--muted)] hover:text-[var(--text)]"}`}
    >
      {icon}
      {children}
    </Link>
  );
}
