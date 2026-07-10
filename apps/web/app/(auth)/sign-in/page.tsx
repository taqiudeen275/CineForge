import { Suspense } from "react";
import { AuthForm } from "@/components/auth-form";
export default function SignInPage() {
  return (
    <Suspense fallback={<section className="auth-card">Opening secure sign-in…</section>}>
      <AuthForm />
    </Suspense>
  );
}
