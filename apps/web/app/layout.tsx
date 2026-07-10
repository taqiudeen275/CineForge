import type { Metadata } from "next";
import "./globals.css";

export const metadata: Metadata = {
  metadataBase: new URL(process.env.NEXT_PUBLIC_WEB_ORIGIN ?? "http://localhost:3000"),
  title: { default: "CineForge — Story-first AI filmmaking", template: "%s · CineForge" },
  description: "Develop the world, direct the shots, and finish the story with a private, traceable AI production studio.",
  openGraph: { title: "CineForge — Story-first AI filmmaking", description: "One production truth from first idea to final cut.", type: "website", images: [{ url: "/opengraph-image" }] },
  twitter: { card: "summary_large_image" },
};

export default function RootLayout({ children }: Readonly<{ children: React.ReactNode }>) {
  return <html lang="en" suppressHydrationWarning><body>{children}</body></html>;
}
