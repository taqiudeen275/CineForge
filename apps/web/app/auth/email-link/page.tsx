import { Suspense } from "react";
import { EmailLinkCompletion } from "@/components/email-link-completion";
export default function EmailLinkPage(){return <Suspense fallback={<div className="auth-card">Completing secure sign-in…</div>}><EmailLinkCompletion/></Suspense>}
