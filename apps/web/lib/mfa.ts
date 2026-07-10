import {
  getMultiFactorResolver,
  TotpMultiFactorGenerator,
  type MultiFactorError,
  type UserCredential,
} from "firebase/auth";
import { getIdentity } from "@/lib/firebase";
export async function resolveMfa(error: unknown): Promise<UserCredential> {
  if ((error as { code?: string }).code !== "auth/multi-factor-auth-required") throw error;
  const resolver = getMultiFactorResolver(getIdentity(), error as MultiFactorError);
  const hint = resolver.hints.find((v) => v.factorId === TotpMultiFactorGenerator.FACTOR_ID);
  if (!hint) throw new Error("This account requires an unsupported second factor");
  const code = prompt("Enter the 6-digit code from your authenticator app");
  if (!code) throw new Error("Multi-factor authentication was cancelled");
  return resolver.resolveSignIn(TotpMultiFactorGenerator.assertionForSignIn(hint.uid, code.trim()));
}
