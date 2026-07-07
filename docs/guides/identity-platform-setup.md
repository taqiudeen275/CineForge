# Identity Platform setup guide

Last checked: 2026-07-07.

This slice uses Google Identity Platform through the Firebase Authentication SDKs. That is expected: Identity Platform is the Google Cloud product, while Firebase Auth is the client/admin SDK surface used by web apps and server code.

## How CineForge uses it

1. The browser signs in with the Firebase web SDK using email/password or Google.
2. The browser receives a short-lived ID token.
3. The browser sends that token to `POST /v1/auth/session`.
4. The Go API verifies the token with the Firebase Admin SDK / Identity Platform.
5. The Go API creates server-owned HttpOnly session cookies and a CineForge device session.
6. The browser signs out of Firebase client state so long-lived app sessions are controlled by the Go API.
7. Workspace permissions, project access, spend controls, and audit trails are enforced by CineForge, not by Firebase client state.

## Why `NEXT_PUBLIC_FIREBASE_*` is public

Next.js only exposes browser-side environment variables when they start with `NEXT_PUBLIC_`. The Firebase web SDK runs in the browser, so it needs the public Firebase app configuration there:

- `apiKey`
- `authDomain`
- `projectId`

These values identify the Firebase/Identity Platform project and web app. They are not database passwords and are not trusted for authorization. Google’s Firebase docs describe the Firebase config object as the setup object used to connect the web app to Firebase resources, and Firebase API-key docs explain that Firebase-related API keys identify the project/app rather than authorize privileged API access.

What must stay private:

- service account credentials
- OAuth client secret
- session-cookie signing/admin credentials
- database URL/passwords
- provider API keys

## Local development with the Auth emulator

Use the emulator for ordinary local testing. It keeps auth data out of production and lets you sign up test users freely.

Expected local values:

```env
NEXT_PUBLIC_FIREBASE_API_KEY=local-api-key
NEXT_PUBLIC_FIREBASE_AUTH_DOMAIN=localhost
NEXT_PUBLIC_FIREBASE_PROJECT_ID=cineforge-local
FIREBASE_PROJECT_ID=cineforge-local
FIREBASE_AUTH_EMULATOR_HOST=localhost:9099
```

Start local services:

```powershell
docker compose -f infra/dev/compose.yaml up -d postgres redis temporal temporal-ui fake-gcs gcs-setup clamav mailpit firebase-emulator migrate
```

Then run the app:

```powershell
docker compose -f infra/dev/compose.yaml --profile app up --build
```

Open:

- Web app: <http://localhost:3000>
- Firebase Emulator UI: <http://localhost:4000>
- Mailpit: <http://localhost:8025>
- Temporal UI: <http://localhost:8233>

### Local email/password test flow

1. Open <http://localhost:3000/sign-up>.
2. Create a user with email/password.
3. The app redirects to `/verify-email` because the API blocks workspace bootstrap for unverified users.
4. Watch the Firebase emulator logs for the verification URL:

   ```powershell
   docker compose -f infra/dev/compose.yaml logs -f firebase-emulator
   ```

5. Open the printed verification link in your browser.
6. Return to <http://localhost:3000/sign-in> and sign in again.
7. The app should exchange the fresh ID token, create a server session, bootstrap the personal workspace, and send you to `/app`.

Mailpit is not used for Firebase Auth verification or password reset in local emulator mode. Mailpit is currently used for CineForge workspace invitation emails.

### Local Google sign-in

For the emulator, Google sign-in may be simulated by the emulator UI or by SDK flows depending on the emulator image/runtime. For the most predictable local test of this slice, use email/password first. Use real Google OAuth only after configuring a real Identity Platform project.

## Production Identity Platform setup

### 1. Create or select a Google Cloud project

Use a real Google Cloud project for production, usually the same project where Cloud Run, Cloud SQL, GCS, and Secret Manager will live.

Recommended baseline:

- billing enabled
- production and non-production projects separated
- least-privilege deployment/service accounts
- no long-lived service account JSON keys in the repo

### 2. Enable Identity Platform

In Google Cloud Console:

1. Go to Identity Platform.
2. Enable Identity Platform for the project.
3. Confirm the project ID; this becomes `FIREBASE_PROJECT_ID` and `NEXT_PUBLIC_FIREBASE_PROJECT_ID`.

### 3. Add Firebase to the same Google Cloud project

Because the web SDK setup/config is exposed through Firebase tooling:

1. Open Firebase Console.
2. Add Firebase to the existing Google Cloud project.
3. Register a Web app.
4. Copy the web config object values into:

   ```env
   NEXT_PUBLIC_FIREBASE_API_KEY=...
   NEXT_PUBLIC_FIREBASE_AUTH_DOMAIN=...
   NEXT_PUBLIC_FIREBASE_PROJECT_ID=...
   ```

For this app, we only need Firebase Auth client config. We are not using Firestore, Firebase Storage, or Firebase Hosting as the source of truth.

### 4. Configure authorized domains

In Firebase Authentication / Identity Platform settings, add the domains that are allowed to complete auth redirects:

- local dev if testing against real Identity Platform: `localhost`
- staging domain, for example `staging.cineforge.ai`
- production domain, for example `app.cineforge.ai`

Do not rely on `localhost` being automatically present in new projects; check the authorized-domain list explicitly.

### 5. Enable email/password provider

In Identity Platform:

1. Go to Providers / Sign-in method.
2. Add or enable Email/Password.
3. Keep email enumeration protection and abuse controls in mind for production.
4. Decide whether self-service sign-up is open, invite-gated, or waitlist-gated. The current slice supports open sign-up, but product policy can later add gating before workspace creation.

For production email deliverability and branding:

- configure sender/from branding in Identity Platform/Firebase templates
- set the app/action URL to your production domain
- test verification and password-reset emails on real inboxes
- align copy with CineForge’s security wording

### 6. Enable Google provider

In Identity Platform:

1. Add Google as a provider.
2. Create or select a Google OAuth web client from Google Cloud APIs & Services.
3. Configure the OAuth consent screen.
4. Add the authorized JavaScript origins:

   - `https://app.cineforge.ai`
   - staging origin if applicable
   - `http://localhost:3000` only for development/testing

5. Add authorized redirect URIs required by the provider setup. Identity Platform/Firebase commonly uses a handler under the auth domain, for example:

   ```text
   https://<project-id>.firebaseapp.com/__/auth/handler
   ```

6. Copy the OAuth client ID and secret into the provider configuration in Identity Platform. The OAuth client secret is not a frontend environment variable.

### 7. Configure the Go API credentials

Production should use Application Default Credentials:

- Cloud Run: attach a least-privilege service account.
- GKE: use Workload Identity.
- Local production-like test: use:

  ```powershell
  gcloud auth application-default login
  ```

Avoid service account JSON keys. If one is temporarily unavoidable:

1. Store it outside the repository.
2. Set `GOOGLE_APPLICATION_CREDENTIALS` only in your local shell/session.
3. Rotate/delete it after use.
4. Never put it in `.env`, Docker images, logs, or commits.

### 8. Production environment variables

Frontend:

```env
NEXT_PUBLIC_API_BASE_URL=/api
NEXT_PUBLIC_FIREBASE_API_KEY=<web-app-api-key>
NEXT_PUBLIC_FIREBASE_AUTH_DOMAIN=<project-id>.firebaseapp.com
NEXT_PUBLIC_FIREBASE_PROJECT_ID=<project-id>
```

Go API:

```env
APP_ENV=production
FIREBASE_PROJECT_ID=<project-id>
SESSION_COOKIE_SECURE=true
WEB_ORIGIN=https://app.cineforge.ai
```

Do not set `FIREBASE_AUTH_EMULATOR_HOST` in production.

## Verification checklist

- Email/password sign-up creates an unverified user.
- Unverified user cannot bootstrap a workspace.
- Verification email link works.
- Verified user can sign in and reaches `/app`.
- Google sign-in reaches `/app`.
- `POST /v1/auth/session` sets HttpOnly session cookies.
- Firebase client state is cleared after session exchange.
- Server-side `/v1/me` works after refresh.
- Workspace/project calls fail without valid server session.
- Logout clears cookies and revokes the CineForge device session.
- Password reset sends a real email in production.
- Audit logs never contain raw tokens, passwords, or full auth links.

## Official references

- [Identity Platform authentication concepts](https://docs.cloud.google.com/identity-platform/docs/concepts-authentication)
- [Sign in users with Google using Identity Platform](https://docs.cloud.google.com/identity-platform/docs/web/google)
- [Sign in with email/password using Identity Platform](https://docs.cloud.google.com/identity-platform/docs/sign-in-user-email)
- [Firebase web setup and config object](https://firebase.google.com/docs/web/setup)
- [Firebase API keys](https://firebase.google.com/docs/projects/api-keys)
- [Firebase Auth session cookies](https://firebase.google.com/docs/auth/admin/manage-cookies)
- [Firebase Authentication emulator](https://firebase.google.com/docs/emulator-suite/connect_auth)
- [Application Default Credentials](https://docs.cloud.google.com/docs/authentication/application-default-credentials)
- [Local ADC setup](https://docs.cloud.google.com/docs/authentication/set-up-adc-local-dev-environment)
- [Service account key best practices](https://docs.cloud.google.com/iam/docs/best-practices-for-managing-service-account-keys)
