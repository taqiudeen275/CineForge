# Identity Platform setup guide

Last checked: 2026-07-10.

CineForge supports Google sign-in and passwordless email links through Google Identity Platform. The browser holds Identity Platform state only long enough to obtain an ID token. `POST /v1/auth/session` exchanges that token for CineForge-owned HttpOnly cookies, then the frontend clears Firebase client state. Authorization, device sessions, workspaces, projects, spend controls, and audit events remain server-owned.

## Local development

Use these values with the Firebase Auth emulator:

```env
NEXT_PUBLIC_FIREBASE_API_KEY=local-api-key
NEXT_PUBLIC_FIREBASE_AUTH_DOMAIN=localhost
NEXT_PUBLIC_FIREBASE_PROJECT_ID=cineforge-local
FIREBASE_PROJECT_ID=cineforge-local
FIREBASE_AUTH_EMULATOR_HOST=localhost:9099
```

Start the dependencies and application:

```powershell
docker compose -f infra/dev/compose.yaml up -d postgres redis temporal temporal-ui fake-gcs gcs-setup clamav mailpit firebase-emulator migrate
docker compose -f infra/dev/compose.yaml --profile app up --build
```

Open the web app at <http://localhost:3000>, Mailpit at <http://localhost:8025>, and the Auth Emulator UI at <http://localhost:4000>.

### Test a passwordless sign-in

1. Open `/sign-in` and choose **Continue with email**.
2. Enter any test email address. The API always returns the same generic confirmation.
3. Open the branded message in Mailpit and follow its one-time link.
4. On another browser/device, re-enter the same email when prompted; the email is never embedded in the URL.
5. The callback exchanges the ID token for a server session and redirects only to a validated `/app/**` destination.

Expired, reused, or mismatched links produce a recoverable error and never reveal whether an account existed. Mail and audit logs must not include raw tokens or complete action links.

## Production configuration

1. Enable Identity Platform on the production Google Cloud project and register a Firebase Web app on that same project.
2. Enable the Email/Password provider and email-link sign-in. The provider must remain enabled for email-link authentication even though CineForge never presents a password form.
3. Enable Google as an independent provider, configure its OAuth consent screen, and add the Identity Platform auth handler and application origins.
4. Add the production, staging, and intentional local domains to Authorized Domains.
5. Enable TOTP MFA for team-workspace owners. CineForge blocks designated high-risk owner operations until enrollment and recent authentication are satisfied.
6. Configure the API with Application Default Credentials and a least-privilege runtime identity; do not ship service-account JSON files.
7. Configure CineForge's SMTP provider, sender authentication, and branded templates. Mailpit is development-only.

Frontend configuration:

```env
NEXT_PUBLIC_API_BASE_URL=/api
NEXT_PUBLIC_FIREBASE_API_KEY=<web-app-api-key>
NEXT_PUBLIC_FIREBASE_AUTH_DOMAIN=<project-id>.firebaseapp.com
NEXT_PUBLIC_FIREBASE_PROJECT_ID=<project-id>
```

API configuration:

```env
APP_ENV=production
FIREBASE_PROJECT_ID=<project-id>
SESSION_COOKIE_SECURE=true
WEB_ORIGIN=https://app.cineforge.ai
```

Do not set `FIREBASE_AUTH_EMULATOR_HOST` in production.

## Legacy password cutover

Do not copy password hashes into CineForge. During the communicated migration window, keep the Identity Platform provider enabled and direct users to an email link. After successful email-link ownership verification and operational review, invalidate remaining legacy password credentials with an audited Identity Platform Admin migration, revoke existing refresh tokens where policy requires it, and retain Google/email-link identities. Run this as a separately approved production operation with a dry-run export, rollback plan, and support communication.

## Verification checklist

- Email-link request responses are enumeration-safe and rate-limited.
- Same-device and cross-device completion work; expired, reused, and wrong-email links fail safely.
- Google and email-link sign-ins create one CineForge server session and clear Firebase client state.
- `/app/**` redirects unauthenticated users to `/sign-in` with only a safe relative destination.
- `/v1/me` remains authoritative after refresh, expiration, revocation, or account-deletion recovery.
- Owner TOTP enrollment and recent-auth challenges protect designated high-risk operations.
- Logout clears cookies and revokes the CineForge device session.
- Logs and audits contain neither passwords, raw tokens, nor full authentication links.

## Official references

- [Firebase email-link authentication](https://firebase.google.com/docs/auth/web/email-link-auth)
- [Firebase TOTP MFA](https://firebase.google.com/docs/auth/web/totp-mfa)
- [Identity Platform Google sign-in](https://docs.cloud.google.com/identity-platform/docs/web/google)
- [Firebase Auth session cookies](https://firebase.google.com/docs/auth/admin/manage-cookies)
- [Firebase Authentication emulator](https://firebase.google.com/docs/emulator-suite/connect_auth)
- [Application Default Credentials](https://docs.cloud.google.com/docs/authentication/application-default-credentials)
