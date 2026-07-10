import Link from "next/link";
export function Brand({ href = "/" }: { href?: string }) {
  return <Link href={href} className="brand-mark" aria-label="CineForge home"><span className="brand-glyph" aria-hidden="true"><span /></span><span>CINEFORGE</span></Link>;
}
