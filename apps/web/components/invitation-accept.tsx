"use client";
import type { components } from "@cineforge/api-client";
import { apiFetch, ApiError } from "@cineforge/api-client";
import { MailCheck } from "lucide-react";
import { useRouter, useSearchParams } from "next/navigation";
import { useEffect, useState } from "react";
import { Brand } from "@/components/brand";
import { Button } from "@/components/ui";
type Workspace = components["schemas"]["Workspace"];
export function InvitationAccept() {
  const params = useSearchParams();
  const router = useRouter();
  const [state, setState] = useState<"checking" | "ready" | "accepting" | "error">("checking");
  const [error, setError] = useState("");
  const token = params.get("token") ?? "";
  useEffect(() => {
    if (!token) {
      setError("This invitation link is incomplete.");
      setState("error");
      return;
    }
    apiFetch("/v1/me")
      .then(() => setState("ready"))
      .catch((e) => {
        if (e instanceof ApiError && e.status === 401)
          router.replace(`/sign-in?next=${encodeURIComponent(`/invite?token=${token}`)}`);
        else {
          setError(e.message);
          setState("error");
        }
      });
  }, [token, router]);
  async function accept() {
    setState("accepting");
    try {
      const ws = await apiFetch<Workspace>("/v1/invitations/accept", {
        method: "POST",
        body: { token },
      });
      router.replace(`/app/${ws.slug}/projects`);
    } catch (e) {
      setError(
        e instanceof ApiError && e.status === 403
          ? "This invitation belongs to a different email address."
          : e instanceof Error
            ? e.message
            : "This invitation is invalid or expired.",
      );
      setState("error");
    }
  }
  return (
    <main className="invite-page">
      <Brand />
      <section>
        <div className="modal-icon">
          <MailCheck />
        </div>
        <h1>{state === "error" ? "Invitation unavailable" : "Join the workspace"}</h1>
        <p>
          {state === "error"
            ? error
            : "You were invited to collaborate in a private CineForge workspace. Your assigned role applies across its projects."}
        </p>
        {state === "ready" ? (
          <Button onClick={accept}>Accept invitation</Button>
        ) : state === "accepting" ? (
          <Button disabled>Joining workspace…</Button>
        ) : state === "checking" ? (
          <span className="muted">Checking your secure session…</span>
        ) : (
          <Button variant="quiet" onClick={() => router.replace("/app")}>
            Go to CineForge
          </Button>
        )}
      </section>
    </main>
  );
}
