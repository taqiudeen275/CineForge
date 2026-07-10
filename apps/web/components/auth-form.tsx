"use client";
import { useState } from "react";
import Link from "next/link";
import { useRouter, useSearchParams } from "next/navigation";
import { GoogleAuthProvider, signInWithPopup, signOut } from "firebase/auth";
import { apiFetch } from "@cineforge/api-client";
import { ArrowRight, Mail } from "lucide-react";
import { configureEphemeralIdentityPersistence, getIdentity } from "@/lib/firebase";
import { resolveMfa } from "@/lib/mfa";
import { Button, Field, Input } from "@/components/ui";

export function AuthForm() {
  const router = useRouter(); const params = useSearchParams();
  const [email,setEmail]=useState(""); const [sent,setSent]=useState(false); const [error,setError]=useState(""); const [busy,setBusy]=useState(false);
  const next = safeNext(params.get("next"));
  async function exchange(user: import("firebase/auth").User) { const token=await user.getIdToken(); await ensureCsrf(); const result=await apiFetch<{requiresVerification:boolean}>("/v1/auth/session",{method:"POST",body:{idToken:token,deviceLabel:browserLabel()}}); await signOut(getIdentity()); router.replace(result.requiresVerification?"/verify-email":next); router.refresh(); }
  async function requestLink(e:React.FormEvent){ e.preventDefault(); setBusy(true); setError(""); try{ await ensureCsrf(); await apiFetch("/v1/auth/email-link",{method:"POST",body:{email,next}}); localStorage.setItem("cf-email-for-sign-in",email.trim().toLowerCase()); localStorage.setItem("cf-auth-next",next); setSent(true); }catch(e){setError(e instanceof Error?e.message:"Could not send the sign-in link")}finally{setBusy(false)} }
  async function google(){setBusy(true);setError("");try{await configureEphemeralIdentityPersistence();let credential;try{credential=await signInWithPopup(getIdentity(),new GoogleAuthProvider())}catch(e){credential=await resolveMfa(e)}await exchange(credential.user)}catch(e){setError(e instanceof Error?e.message:"Google sign-in failed")}finally{setBusy(false)}}
  return <section className="auth-card"><div className="auth-kicker">YOUR PRIVATE STUDIO</div><h1>{sent?"Check your inbox":"Welcome to CineForge"}</h1><p>{sent?<>We sent a secure sign-in link to <strong>{email}</strong>. It expires automatically.</>:"Sign in or create an account. No password required."}</p>{sent?<div className="auth-sent"><Mail/><div><strong>Open the link on this device</strong><span>You can also continue elsewhere by confirming your email.</span></div><Button variant="quiet" onClick={()=>setSent(false)}>Use another email</Button></div>:<><form onSubmit={requestLink} className="auth-form"><Field label="Email address"><Input value={email} onChange={e=>setEmail(e.target.value)} type="email" autoComplete="email" placeholder="you@example.com" required /></Field>{error?<p role="alert" className="auth-error">{error}</p>:null}<Button disabled={busy} type="submit" className="w-full">{busy?"Sending…":<>Continue with email <ArrowRight size={16}/></>}</Button></form><div className="auth-divider"><span/>OR<span/></div><Button className="w-full" variant="quiet" onClick={google} disabled={busy}><GoogleMark/>Continue with Google</Button></>}<p className="auth-terms">By continuing, you agree to use CineForge responsibly. Your projects are private by default.</p><Link href="/" className="auth-back">← Back to CineForge</Link></section>;
}
export async function ensureCsrf(){const response=await fetch("/api/v1/auth/csrf",{credentials:"include"});if(!response.ok)throw new Error("Could not establish a secure session")}
export function safeNext(value:string|null){return value==="/app"||value?.startsWith("/app/")||value?.startsWith("/invite?token=")?value:"/app"}
function browserLabel(){return `${navigator.platform||"Web"} · ${navigator.userAgent.includes("Firefox")?"Firefox":navigator.userAgent.includes("Edg")?"Edge":navigator.userAgent.includes("Chrome")?"Chrome":"Browser"}`}
function GoogleMark(){return <span className="google-mark" aria-hidden="true">G</span>}
