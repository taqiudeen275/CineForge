"use client";

import { useState } from "react";
import Link from "next/link";
import { useRouter } from "next/navigation";
import {
  createUserWithEmailAndPassword,
  GoogleAuthProvider,
  sendEmailVerification,
  signInWithEmailAndPassword,
  signInWithPopup,
  signOut,
  updateProfile,
} from "firebase/auth";
import { apiFetch } from "@cineforge/api-client";
import { configureEphemeralIdentityPersistence, getIdentity } from "@/lib/firebase";
import { Button, Field, Input } from "@/components/ui";

export function AuthForm({ mode }: { mode: "sign-in" | "sign-up" }) {
  const router = useRouter();
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [name, setName] = useState("");
  const [error, setError] = useState("");
  const [busy, setBusy] = useState(false);
  async function exchange(user: import("firebase/auth").User) {
    const token = await user.getIdToken();
    await ensureCsrf();
    const result = await apiFetch<{ requiresVerification: boolean }>("/v1/auth/session", {
      method: "POST",
      body: { idToken: token, deviceLabel: browserLabel() },
    });
    await signOut(getIdentity());
    router.push(result.requiresVerification ? "/verify-email" : "/app");
    router.refresh();
  }
  async function submit(event: React.FormEvent) {
    event.preventDefault();
    setBusy(true);
    setError("");
    try {
      await configureEphemeralIdentityPersistence();
      const auth = getIdentity();
      const credential =
        mode === "sign-up"
          ? await createUserWithEmailAndPassword(auth, email, password)
          : await signInWithEmailAndPassword(auth, email, password);
      if (mode === "sign-up") {
        if (name.trim()) await updateProfile(credential.user, { displayName: name.trim() });
        await sendEmailVerification(credential.user);
      }
      await exchange(credential.user);
    } catch (e) {
      setError(e instanceof Error ? e.message : "Authentication failed");
    } finally {
      setBusy(false);
    }
  }
  async function google() {
    setBusy(true);
    setError("");
    try {
      await configureEphemeralIdentityPersistence();
      const credential = await signInWithPopup(getIdentity(), new GoogleAuthProvider());
      await exchange(credential.user);
    } catch (e) {
      setError(e instanceof Error ? e.message : "Google sign-in failed");
    } finally {
      setBusy(false);
    }
  }
  return (
    <section className="w-full max-w-sm">
      <h1 className="m-0 text-3xl font-semibold tracking-[-.04em]">
        {mode === "sign-in" ? "Welcome back" : "Create your account"}
      </h1>
      <p className="mb-8 mt-2 text-sm text-[var(--muted)]">
        {mode === "sign-in"
          ? "Continue building your story world."
          : "Start in a private personal workspace."}
      </p>
      <form className="grid gap-5" onSubmit={submit}>
        {mode === "sign-up" ? (
          <Field label="Display name">
            <Input
              value={name}
              onChange={(e) => setName(e.target.value)}
              autoComplete="name"
              required
            />
          </Field>
        ) : null}
        <Field label="Email">
          <Input
            value={email}
            onChange={(e) => setEmail(e.target.value)}
            type="email"
            autoComplete="email"
            required
          />
        </Field>
        <Field label="Password">
          <Input
            value={password}
            onChange={(e) => setPassword(e.target.value)}
            type="password"
            minLength={10}
            autoComplete={mode === "sign-in" ? "current-password" : "new-password"}
            required
          />
        </Field>
        {error ? (
          <p role="alert" className="m-0 text-sm text-[var(--danger)]">
            {error}
          </p>
        ) : null}
        <Button disabled={busy} type="submit">
          {busy ? "Working…" : mode === "sign-in" ? "Sign in" : "Create account"}
        </Button>
      </form>
      <div className="my-6 flex items-center gap-3 text-xs text-[var(--muted)]">
        <span className="h-px flex-1 bg-[var(--line)]" />
        OR
        <span className="h-px flex-1 bg-[var(--line)]" />
      </div>
      <Button className="w-full" variant="quiet" onClick={google} disabled={busy}>
        Continue with Google
      </Button>
      <p className="mt-8 text-sm text-[var(--muted)]">
        {mode === "sign-in" ? (
          <>
            New to CineForge?{" "}
            <Link className="text-[var(--text)] underline underline-offset-4" href="/sign-up">
              Create an account
            </Link>
          </>
        ) : (
          <>
            Already have an account?{" "}
            <Link className="text-[var(--text)] underline underline-offset-4" href="/sign-in">
              Sign in
            </Link>
          </>
        )}
      </p>
      {mode === "sign-in" ? (
        <Link className="mt-3 block text-sm text-[var(--muted)]" href="/forgot-password">
          Forgot password?
        </Link>
      ) : null}
    </section>
  );
}

async function ensureCsrf() {
  const response = await fetch("/api/v1/auth/csrf", { credentials: "include" });
  if (!response.ok) throw new Error("Could not establish a secure session");
}
function browserLabel() {
  return `${navigator.platform || "Web"} · ${navigator.userAgent.includes("Firefox") ? "Firefox" : navigator.userAgent.includes("Edg") ? "Edge" : navigator.userAgent.includes("Chrome") ? "Chrome" : "Browser"}`;
}
