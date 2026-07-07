import { afterEach, expect, test } from "bun:test";
import { ApiError, apiFetch, apiUpload } from "./client";

const originalFetch = globalThis.fetch;

afterEach(() => {
  globalThis.fetch = originalFetch;
  Reflect.deleteProperty(globalThis, "document");
});

test("apiFetch maps JSON bodies, CSRF, ETag, and idempotency headers", async () => {
  let seenUrl = "";
  let seenInit: RequestInit | undefined;
  Object.defineProperty(globalThis, "document", {
    configurable: true,
    value: { cookie: "cf_csrf=csrf-token; theme=dark" },
  });
  globalThis.fetch = ((url: URL | RequestInfo, init?: RequestInit) => {
    seenUrl = String(url);
    seenInit = init;
    return Promise.resolve(
      new Response(JSON.stringify({ ok: true }), {
        status: 200,
        headers: { "Content-Type": "application/json" },
      }),
    );
  }) as typeof fetch;

  const result = await apiFetch<{ ok: boolean }>("/v1/workspaces", {
    method: "POST",
    etag: "3",
    idempotencyKey: "a".repeat(16),
    body: { name: "Story lab" },
  });

  const headers = new Headers(seenInit?.headers);
  expect(result.ok).toBe(true);
  expect(seenUrl).toBe("/api/v1/workspaces");
  expect(seenInit?.credentials).toBe("include");
  expect(seenInit?.body).toBe(JSON.stringify({ name: "Story lab" }));
  expect(headers.get("Accept")).toBe("application/json");
  expect(headers.get("Content-Type")).toBe("application/json");
  expect(headers.get("If-Match")).toBe("3");
  expect(headers.get("Idempotency-Key")).toBe("a".repeat(16));
  expect(headers.get("X-CSRF-Token")).toBe("csrf-token");
});

test("apiFetch raises ApiError using problem detail", async () => {
  globalThis.fetch = (() =>
    Promise.resolve(
      new Response(JSON.stringify({ detail: "Workspace access denied", code: "forbidden" }), {
        status: 403,
        headers: { "Content-Type": "application/problem+json" },
      }),
    )) as unknown as typeof fetch;

  try {
    await apiFetch("/v1/workspaces/not-yours");
    throw new Error("expected apiFetch to throw");
  } catch (error) {
    expect(error).toBeInstanceOf(ApiError);
    expect((error as ApiError).message).toBe("Workspace access denied");
    expect((error as ApiError).status).toBe(403);
  }
});

test("apiUpload sends raw file content with CSRF only", async () => {
  let seenInit: RequestInit | undefined;
  Object.defineProperty(globalThis, "document", {
    configurable: true,
    value: { cookie: "cf_csrf=upload-csrf" },
  });
  globalThis.fetch = ((_: URL | RequestInfo, init?: RequestInit) => {
    seenInit = init;
    return Promise.resolve(new Response(null, { status: 204 }));
  }) as typeof fetch;

  const file = new File(["hello"], "reference.png", { type: "image/png" });
  await apiUpload("/v1/media/uploads/asset-id/content", file);

  const headers = new Headers(seenInit?.headers);
  expect(seenInit?.method).toBe("PUT");
  expect(seenInit?.credentials).toBe("include");
  expect(seenInit?.body).toBe(file);
  expect(headers.get("Content-Type")).toBe("image/png");
  expect(headers.get("X-CSRF-Token")).toBe("upload-csrf");
});
