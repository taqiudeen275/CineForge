"use client";
import type { components } from "@cineforge/api-client";
import { apiFetch, ApiError } from "@cineforge/api-client";
import {
  BookOpen,
  ChevronDown,
  FolderKanban,
  LogOut,
  Menu,
  Plus,
  Settings,
  UserRound,
  X,
} from "lucide-react";
import Link from "next/link";
import { usePathname, useRouter } from "next/navigation";
import { createContext, useContext, useEffect, useMemo, useState } from "react";
import { Brand } from "@/components/brand";
import { Button, Input } from "@/components/ui";
import { ThemeToggle } from "@/components/theme-toggle";

type Workspace = components["schemas"]["Workspace"];
type User = components["schemas"]["User"];
type ShellState = {
  user: User;
  workspaces: Workspace[];
  workspace?: Workspace;
  canEdit: boolean;
  canAdmin: boolean;
  refresh: () => Promise<void>;
};
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
  const [status, setStatus] = useState<"loading" | "ready" | "error" | "deletion" | "profile">(
    "loading",
  );
  const [error, setError] = useState("");
  const [mobile, setMobile] = useState(false);
  const [creating, setCreating] = useState(false);
  const [workspaceName, setWorkspaceName] = useState("");
  // Shell bootstrap intentionally runs once; explicit refreshes call load directly.
  // biome-ignore lint/correctness/useExhaustiveDependencies: load is the session bootstrap helper.
  useEffect(() => {
    void load();
  }, []);
  async function load() {
    setStatus("loading");
    setError("");
    try {
      const me = await apiFetch<User>("/v1/me");
      setUser(me);
      if (!me.emailVerified) {
        router.replace("/verify-email");
        return;
      }
      if (me.deletionScheduledAt) {
        setStatus("deletion");
        return;
      }
      let list = await apiFetch<{ items: Workspace[] }>("/v1/workspaces");
      if (!list.items.length && !me.personalWorkspaceId) {
        setStatus("profile");
        return;
      }
      if (!list.items.length) {
        const ws = await apiFetch<Workspace>("/v1/workspaces/bootstrap", {
          method: "POST",
          body: {},
        });
        list = { items: [ws] };
      }
      setWorkspaces(list.items);
      setStatus("ready");
    } catch (e) {
      if (e instanceof ApiError && e.status === 401) {
        router.replace(`/sign-in?next=${encodeURIComponent(pathname)}`);
        return;
      }
      setError(e instanceof Error ? e.message : "CineForge could not load your workspace");
      setStatus("error");
    }
  }
  const slug = pathname.split("/")[2];
  const exact = workspaces.find((w) => w.slug === slug);
  const workspace = exact ?? workspaces[0];
  useEffect(() => {
    if (status === "ready" && slug && workspace && !exact && !pathname.startsWith("/app/settings"))
      router.replace(`/app/${workspace.slug}/projects`);
  }, [status, slug, workspace, exact, pathname, router]);
  const role = workspace?.currentMembership?.role;
  const canEdit = role === "owner" || role === "admin" || role === "editor";
  const canAdmin = role === "owner" || role === "admin";
  // biome-ignore lint/correctness/useExhaustiveDependencies: load is stable for the mounted shell.
  const value = useMemo(
    () => (user ? { user, workspaces, workspace, canEdit, canAdmin, refresh: load } : undefined),
    [user, workspaces, workspace, canEdit, canAdmin],
  );
  async function createWorkspace(e: React.FormEvent) {
    e.preventDefault();
    if (!workspaceName.trim()) return;
    const ws = await apiFetch<Workspace>("/v1/workspaces", {
      method: "POST",
      idempotencyKey: crypto.randomUUID() + crypto.randomUUID(),
      body: { name: workspaceName.trim() },
    });
    setCreating(false);
    setWorkspaceName("");
    await load();
    router.push(`/app/${ws.slug}/projects`);
  }
  async function logout() {
    await apiFetch("/v1/auth/logout", { method: "POST", body: {} }).catch(() => undefined);
    router.replace("/sign-in");
    router.refresh();
  }
  async function cancelDeletion() {
    await apiFetch("/v1/me/deletion/cancel", { method: "POST", body: {} });
    await load();
  }
  async function completeProfile(e: React.FormEvent<HTMLFormElement>) {
    e.preventDefault();
    if (!user) return;
    const f = new FormData(e.currentTarget);
    await apiFetch("/v1/me", {
      method: "PATCH",
      etag: String(user.version),
      body: {
        displayName: f.get("name"),
        avatarUrl: null,
        locale: navigator.language || "en",
        timezone: Intl.DateTimeFormat().resolvedOptions().timeZone || "UTC",
      },
    });
    const ws = await apiFetch<Workspace>("/v1/workspaces/bootstrap", { method: "POST", body: {} });
    setWorkspaces([ws]);
    await load();
  }
  if (status === "loading") return <ShellLoading />;
  if (status === "error")
    return (
      <ShellMessage
        title="We couldn’t open your studio"
        copy={error}
        action={<Button onClick={() => void load()}>Try again</Button>}
      />
    );
  if (status === "deletion")
    return (
      <ShellMessage
        title="Your account is scheduled for deletion"
        copy="Cancel deletion to restore access to your private workspaces."
        action={<Button onClick={cancelDeletion}>Cancel deletion</Button>}
      />
    );
  if (status === "profile" && user)
    return (
      <div className="profile-onboarding">
        <Brand />
        <section>
          <span>WELCOME TO CINEFORGE</span>
          <h1>Name your creative studio</h1>
          <p>
            This name identifies you to collaborators. We’ll create a private personal workspace
            next.
          </p>
          <form onSubmit={completeProfile}>
            <Input
              name="name"
              defaultValue={user.displayName}
              placeholder="Your display name"
              required
              maxLength={120}
            />
            <Button>Enter my studio</Button>
          </form>
          <small>{user.email}</small>
        </section>
      </div>
    );
  if (!value || !workspace) return null;
  const base = `/app/${workspace.slug}`;
  return (
    <ShellContext.Provider value={value}>
      <div className="app-frame">
        <aside className={`app-sidebar ${mobile ? "is-open" : ""}`}>
          <div className="sidebar-head">
            <Brand href="/app" />
            <Button
              variant="ghost"
              size="icon"
              className="mobile-only"
              onClick={() => setMobile(false)}
              aria-label="Close navigation"
            >
              <X size={18} />
            </Button>
          </div>
          <button type="button" className="workspace-switcher" onClick={() => setCreating(true)}>
            <span className="workspace-avatar">{workspace.name.slice(0, 1).toUpperCase()}</span>
            <span>
              <strong>{workspace.name}</strong>
              <small>
                {role ?? "member"} · {workspace.kind}
              </small>
            </span>
            <ChevronDown size={15} />
          </button>
          <nav className="app-nav">
            <Nav
              href={`${base}/projects`}
              active={pathname.includes("/projects")}
              icon={<FolderKanban />}
            >
              Projects
            </Nav>
            <Nav
              href={`${base}/library`}
              active={pathname.includes("/library")}
              icon={<BookOpen />}
            >
              Library
            </Nav>
            <Nav
              href={`${base}/settings/general`}
              active={pathname.includes(`/${workspace.slug}/settings`)}
              icon={<Settings />}
            >
              Workspace settings
            </Nav>
          </nav>
          <div className="sidebar-foot">
            <Link href="/app/settings/account" className="user-chip">
              <span>{value.user.displayName.slice(0, 1).toUpperCase()}</span>
              <div>
                <strong>{value.user.displayName}</strong>
                <small>{value.user.email}</small>
              </div>
            </Link>
            <div className="sidebar-tools">
              <ThemeToggle />
              <Button variant="ghost" size="icon" onClick={logout} aria-label="Sign out">
                <LogOut size={17} />
              </Button>
            </div>
          </div>
        </aside>
        <div className="app-content">
          <header className="mobile-header">
            <Button
              variant="ghost"
              size="icon"
              onClick={() => setMobile(true)}
              aria-label="Open navigation"
            >
              <Menu size={19} />
            </Button>
            <Brand href="/app" />
            <Link href="/app/settings/account" aria-label="Account">
              <UserRound size={19} />
            </Link>
          </header>
          <main>
            <div className="content-wrap">{children}</div>
          </main>
        </div>
        {mobile ? (
          <button
            type="button"
            className="nav-scrim"
            aria-label="Close navigation"
            onClick={() => setMobile(false)}
          />
        ) : null}
      </div>
      {creating ? (
        <div className="modal-backdrop">
          <section
            className="modal-card"
            role="dialog"
            aria-modal="true"
            aria-labelledby="workspace-dialog"
          >
            <div className="modal-icon">
              <Plus />
            </div>
            <h2 id="workspace-dialog">Create a team workspace</h2>
            <p>Invite collaborators, assign roles, and manage production spend separately.</p>
            <form onSubmit={createWorkspace}>
              <Input
                value={workspaceName}
                onChange={(e) => setWorkspaceName(e.target.value)}
                placeholder="Studio or team name"
                autoFocus
                required
              />
              <div>
                <Button type="button" variant="quiet" onClick={() => setCreating(false)}>
                  Cancel
                </Button>
                <Button>Create workspace</Button>
              </div>
            </form>
            <div className="workspace-list">
              {workspaces.map((w) => (
                <button
                  type="button"
                  key={w.id}
                  onClick={() => {
                    setCreating(false);
                    router.push(`/app/${w.slug}/projects`);
                  }}
                >
                  <span>{w.name.slice(0, 1).toUpperCase()}</span>
                  <div>
                    <strong>{w.name}</strong>
                    <small>{w.kind}</small>
                  </div>
                </button>
              ))}
            </div>
          </section>
        </div>
      ) : null}
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
    <Link href={href} className={active ? "active" : ""}>
      {icon}
      {children}
    </Link>
  );
}
function ShellLoading() {
  return (
    <div className="shell-loading">
      <div className="brand-glyph">
        <span />
      </div>
      <div>
        <strong>Opening your studio</strong>
        <span>Loading private workspace…</span>
      </div>
    </div>
  );
}
function ShellMessage({
  title,
  copy,
  action,
}: {
  title: string;
  copy: string;
  action: React.ReactNode;
}) {
  return (
    <div className="shell-message">
      <Brand />
      <h1>{title}</h1>
      <p>{copy}</p>
      {action}
    </div>
  );
}
