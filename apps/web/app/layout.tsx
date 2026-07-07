import type { Metadata } from "next";
import "./globals.css";

export const metadata: Metadata = {
  title: { default: "CineForge", template: "%s · CineForge" },
  description: "Develop worlds, direct shots, and finish stories.",
};

export default function RootLayout({ children }: Readonly<{ children: React.ReactNode }>) {
  return (
    <html lang="en">
      <body>{children}</body>
    </html>
  );
}
