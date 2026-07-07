"use client";
import { useState } from "react";
import Link from "next/link";
import { sendPasswordResetEmail } from "firebase/auth";
import { getIdentity } from "@/lib/firebase";
import { Button, Field, Input } from "@/components/ui";
export function ResetPasswordForm() {
  const [email, setEmail] = useState("");
  const [state, setState] = useState<"idle" | "sent" | "error">("idle");
  async function submit(e: React.FormEvent) {
    e.preventDefault();
    try {
      await sendPasswordResetEmail(getIdentity(), email);
      setState("sent");
    } catch {
      setState("sent");
    }
  }
  return (
    <section className="w-full max-w-sm">
      <h1 className="m-0 text-3xl font-semibold tracking-[-.04em]">Reset password</h1>
      <p className="mt-2 text-sm text-[var(--muted)]">
        Enter your email. If an account exists, we’ll send reset instructions.
      </p>
      {state === "sent" ? (
        <div className="mt-7 text-sm leading-6">
          Check your inbox, then{" "}
          <Link className="underline" href="/sign-in">
            return to sign in
          </Link>
          .
        </div>
      ) : (
        <form className="mt-7 grid gap-5" onSubmit={submit}>
          <Field label="Email">
            <Input value={email} onChange={(e) => setEmail(e.target.value)} type="email" required />
          </Field>
          <Button type="submit">Send reset link</Button>
        </form>
      )}
    </section>
  );
}
