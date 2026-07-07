import Link from "next/link";
export default function VerifyEmailPage() {
  return (
    <section className="w-full max-w-sm">
      <div className="mb-5 h-10 w-10 rounded-full bg-[color-mix(in_srgb,var(--accent)_18%,transparent)]" />
      <h1 className="m-0 text-3xl font-semibold tracking-[-.04em]">Verify your email</h1>
      <p className="mt-3 text-sm leading-6 text-[var(--muted)]">
        We sent a verification link to your inbox. After verifying, sign in again to create your
        private workspace.
      </p>
      <Link
        href="/sign-in"
        className="mt-7 inline-flex min-h-10 items-center rounded-lg bg-[var(--accent)] px-4 text-sm font-medium text-white"
      >
        Return to sign in
      </Link>
    </section>
  );
}
