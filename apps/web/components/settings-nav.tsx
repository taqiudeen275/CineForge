import Link from "next/link";
export function SettingsNav({ slug }: { slug: string }) {
  return (
    <nav className="mb-8 flex gap-5 overflow-x-auto border-b border-[var(--line)] text-sm">
      <Link className="pb-3" href={`/app/${slug}/settings/general`}>
        General
      </Link>
      <Link className="pb-3" href={`/app/${slug}/settings/members`}>
        Members
      </Link>
      <Link className="pb-3" href={`/app/${slug}/settings/spend`}>
        Spend policy
      </Link>
    </nav>
  );
}
