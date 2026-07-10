"use client";
import { useEffect, useState } from "react";
import {
  isSignInWithEmailLink,
  signInWithEmailLink,
  signOut,
  type UserCredential,
} from "firebase/auth";
import { useRouter, useSearchParams } from "next/navigation";
import { apiFetch } from "@cineforge/api-client";
import { Button, Field, Input } from "@/components/ui";
import { configureEphemeralIdentityPersistence, getIdentity } from "@/lib/firebase";
import { ensureCsrf, safeNext } from "@/components/auth-form";
import { resolveMfa } from "@/lib/mfa";

export function EmailLinkCompletion() {
  const router = useRouter();
  const params = useSearchParams();
  const [email, setEmail] = useState("");
  const [needsEmail, setNeedsEmail] = useState(false);
  const [error, setError] = useState("");
  const next = safeNext(params.get("next") ?? localStorageSafe("cf-auth-next"));
  // Email-link completion runs once for the incoming one-time URL.
  // biome-ignore lint/correctness/useExhaustiveDependencies: complete captures the immutable callback URL.
  useEffect(() => {
    const saved = localStorage.getItem("cf-email-for-sign-in");
    if (saved) {
      void complete(saved);
    } else setNeedsEmail(true);
  }, []);
  async function complete(address: string) {
    setError("");
    try {
      await configureEphemeralIdentityPersistence();
      const auth = getIdentity();
      if (!isSignInWithEmailLink(auth, window.location.href))
        throw new Error("This sign-in link is invalid or has expired");
      let credential: UserCredential;
      try {
        credential = await signInWithEmailLink(auth, address, window.location.href);
      } catch (e) {
        credential = await resolveMfa(e);
      }
      const token = await credential.user.getIdToken();
      await ensureCsrf();
      await apiFetch("/v1/auth/session", {
        method: "POST",
        body: { idToken: token, deviceLabel: "Email link · Web browser" },
      });
      if (!next.includes("mfa=setup")) await signOut(auth);
      localStorage.removeItem("cf-email-for-sign-in");
      localStorage.removeItem("cf-auth-next");
      router.replace(next);
      router.refresh();
    } catch (e) {
      setError(e instanceof Error ? e.message : "Could not complete sign-in");
      setNeedsEmail(true);
    }
  }
  if (!needsEmail && !error)
    return (
      <section className="auth-card">
        <div className="auth-kicker">SECURE LINK</div>
        <h1>Opening your studio…</h1>
        <p>We are verifying the link and creating your private session.</p>
      </section>
    );
  return (
    <section className="auth-card">
      <div className="auth-kicker">CONFIRM YOUR EMAIL</div>
      <h1>Finish signing in</h1>
      <p>For your security, enter the same address that received this link.</p>
      <form
        className="auth-form"
        onSubmit={(e) => {
          e.preventDefault();
          void complete(email);
        }}
      >
        <Field label="Email address">
          <Input
            value={email}
            onChange={(e) => setEmail(e.target.value)}
            type="email"
            required
            autoFocus
          />
        </Field>
        {error ? (
          <p role="alert" className="auth-error">
            {error}
          </p>
        ) : null}
        <Button>Continue securely</Button>
      </form>
    </section>
  );
}
function localStorageSafe(key: string) {
  return typeof window === "undefined" ? null : localStorage.getItem(key);
}
