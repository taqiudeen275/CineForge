import Link from "next/link";

export function Brand() {
  return (
    <Link href="/app" className="inline-flex items-center gap-3 font-semibold tracking-[.16em]">
      <span className="grid h-8 w-8 place-items-center rounded-lg bg-[var(--accent)] text-sm text-white">
        C
      </span>
      <span>CINEFORGE</span>
    </Link>
  );
}
