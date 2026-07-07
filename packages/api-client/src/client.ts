export class ApiError extends Error {
  constructor(
    message: string,
    public readonly status: number,
    public readonly problem?: unknown,
  ) {
    super(message);
  }
}

export interface RequestOptions extends Omit<RequestInit, "body"> {
  body?: unknown;
  etag?: string;
  idempotencyKey?: string;
}

export async function apiFetch<T>(path: string, options: RequestOptions = {}): Promise<T> {
  const headers = new Headers(options.headers);
  headers.set("Accept", "application/json");
  if (options.body !== undefined) headers.set("Content-Type", "application/json");
  if (options.etag) headers.set("If-Match", options.etag);
  if (options.idempotencyKey) headers.set("Idempotency-Key", options.idempotencyKey);
  const csrf = typeof document === "undefined" ? undefined : readCookie("cf_csrf");
  if (csrf && !["GET", "HEAD", "OPTIONS"].includes(options.method ?? "GET")) {
    headers.set("X-CSRF-Token", csrf);
  }

  const response = await fetch(`/api${path}`, {
    ...options,
    headers,
    credentials: "include",
    body: options.body === undefined ? undefined : JSON.stringify(options.body),
  });

  if (!response.ok) {
    const problem = await response.json().catch(() => undefined);
    throw new ApiError(
      (problem as { detail?: string } | undefined)?.detail ?? "Request failed",
      response.status,
      problem,
    );
  }
  if (response.status === 204) return undefined as T;
  return response.json() as Promise<T>;
}

export async function apiUpload(path: string, file: File): Promise<void> {
  const csrf = readCookie("cf_csrf");
  const headers = new Headers({ "Content-Type": file.type || "application/octet-stream" });
  if (csrf) headers.set("X-CSRF-Token", csrf);
  const response = await fetch(`/api${path}`, {
    method: "PUT",
    headers,
    credentials: "include",
    body: file,
  });
  if (!response.ok) {
    const problem = await response.json().catch(() => undefined);
    throw new ApiError(
      (problem as { detail?: string } | undefined)?.detail ?? "Upload failed",
      response.status,
      problem,
    );
  }
}

function readCookie(name: string): string | undefined {
  return document.cookie
    .split("; ")
    .find((entry) => entry.startsWith(`${name}=`))
    ?.split("=")[1];
}
