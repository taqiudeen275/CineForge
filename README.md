# CineForge

CineForge is an AI-native story production system. This repository currently contains the account, workspace, project, shared-library, and secure media-ingestion foundation described in the [PRD](prd.md).

## Prerequisites

- Bun 1.3.5
- Docker Desktop with the Linux engine running
- Optional host Go 1.24+; Go builds also run in containers

## Local setup

1. Copy `.env.example` to `.env` and keep it uncommitted.
2. Start infrastructure:

   ```powershell
   docker compose -f infra/dev/compose.yaml up -d postgres redis temporal temporal-ui fake-gcs gcs-setup clamav mailpit firebase-emulator migrate
   ```

3. Install frontend dependencies:

   ```powershell
   bun install
   bun run api:generate
   ```

4. Run the application either with the `app` Compose profile or in separate terminals:

   ```powershell
   docker compose -f infra/dev/compose.yaml --profile app up --build
   ```

Local services:

- Web: <http://localhost:3000>
- Control API: <http://localhost:8080/healthz>
- Mailpit: <http://localhost:8025>
- Temporal UI: <http://localhost:8233>
- Firebase Emulator UI: <http://localhost:4000>

## Quality checks

```powershell
bun run api:generate
bun run typecheck
bun run lint
bun run test
docker compose -f infra/dev/compose.yaml build control-api media-worker
```

## Identity Platform credentials

Google Identity Platform is a Google Cloud product, but its web and admin SDK surface is Firebase Authentication. In this repo the browser uses the Identity Platform/Firebase client SDK only long enough to obtain a fresh ID token. The Go API exchanges that token for server-owned HttpOnly cookies, verifies future requests server-side, and stores CineForge workspace authorization in PostgreSQL.

Local development defaults to the Firebase Auth emulator:

- `FIREBASE_AUTH_EMULATOR_HOST=localhost:9099`
- `NEXT_PUBLIC_FIREBASE_API_KEY=local-api-key`
- `NEXT_PUBLIC_FIREBASE_AUTH_DOMAIN=localhost`
- `NEXT_PUBLIC_FIREBASE_PROJECT_ID=cineforge-local`

For the complete local and production setup flow, see [Identity Platform setup](docs/identity-platform-setup.md).

For production:

1. Create or select a Google Cloud project, enable billing, and enable Identity Platform.
2. In Identity Platform, enable Email/Password and Google providers.
3. For Google sign-in, create or reuse the Google OAuth web client ID/secret, configure the OAuth consent screen, add the production app domain, and copy the web SDK setup values into the `NEXT_PUBLIC_FIREBASE_*` environment variables.
4. The browser API key and auth-domain values are public client configuration, not server secrets. Authorization still happens in the Go API.
5. The Go API should use Application Default Credentials. On Cloud Run/GKE this means an attached least-privilege service account/workload identity. For local production-like testing, prefer `gcloud auth application-default login` or service-account impersonation.
6. Avoid long-lived service-account JSON keys. If a temporary key is unavoidable, keep it outside the repo, set `GOOGLE_APPLICATION_CREDENTIALS` only for that shell/session, rotate it, and remove it as soon as possible.

Useful official docs, checked 2026-07-07:

- [Identity Platform Google provider](https://docs.cloud.google.com/identity-platform/docs/web/google)
- [Identity Platform email/password sign-in](https://docs.cloud.google.com/identity-platform/docs/sign-in-user-email)
- [Firebase Auth session cookies](https://firebase.google.com/docs/auth/admin/manage-cookies)
- [Application Default Credentials](https://docs.cloud.google.com/docs/authentication/provide-credentials-adc)
- [Local ADC setup](https://docs.cloud.google.com/docs/authentication/set-up-adc-local-dev-environment)
- [Service account key best practices](https://docs.cloud.google.com/iam/docs/best-practices-for-managing-service-account-keys)

## Security notes

- The local stack uses emulator credentials and is not suitable for Internet exposure.
- Projects and media are private by default.
- Uploaded media is quarantined, scanned, validated, and normalized before it can be attached to a library entity.
- Production deployment must replace fake GCS, Firebase emulator, Mailpit, and local Temporal with the managed services in `docs/system-design.md`.

## Documentation

Start with the [documentation index](docs/README.md). Security requirements in [security-and-trust.md](docs/security-and-trust.md) are release constraints, not optional hardening.
