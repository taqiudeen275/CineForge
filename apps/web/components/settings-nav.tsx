"use client";
import Link from "next/link";
import { usePathname } from "next/navigation";
export function SettingsNav({ slug }: { slug: string }) {
  const path = usePathname();
  const links = [
    ["general", "General"],
    ["members", "People & access"],
    ["spend", "Spend policy"],
  ];
  return (
    <nav className="settings-nav" aria-label="Workspace settings">
      {links.map(([route, label]) => (
        <Link
          key={route}
          className={path.endsWith(`/${route}`) ? "active" : ""}
          href={`/app/${slug}/settings/${route}`}
        >
          {label}
        </Link>
      ))}
    </nav>
  );
}
