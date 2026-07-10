"use client";

import { getApp, getApps, initializeApp } from "firebase/app";
import {
  browserLocalPersistence,
  connectAuthEmulator,
  getAuth,
  inMemoryPersistence,
  setPersistence,
  type Auth,
} from "firebase/auth";

const config = {
  apiKey: process.env.NEXT_PUBLIC_FIREBASE_API_KEY,
  authDomain: process.env.NEXT_PUBLIC_FIREBASE_AUTH_DOMAIN,
  projectId: process.env.NEXT_PUBLIC_FIREBASE_PROJECT_ID,
};

let identity: Auth | undefined;
let emulatorConnected = false;

export function getIdentity(): Auth {
  if (typeof window === "undefined") throw new Error("Identity is only available in the browser");
  if (!config.apiKey || !config.projectId) throw new Error("Identity Platform is not configured");
  if (!identity) {
    const app = getApps().length ? getApp() : initializeApp(config);
    identity = getAuth(app);
    if (!emulatorConnected && config.authDomain === "localhost") {
      connectAuthEmulator(identity, "http://localhost:9099", { disableWarnings: true });
      emulatorConnected = true;
    }
  }
  return identity;
}

export async function configureEphemeralIdentityPersistence() {
  const auth = getIdentity();
  try {
    await setPersistence(auth, inMemoryPersistence);
  } catch {
    await setPersistence(auth, browserLocalPersistence);
  }
}
