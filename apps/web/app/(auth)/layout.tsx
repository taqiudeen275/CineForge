import { Brand } from "@/components/brand";

export default function AuthLayout({ children }: { children: React.ReactNode }) {
  return (
    <main className="grid min-h-screen grid-rows-[auto_1fr] p-6 sm:p-10">
      <Brand />
      <div className="grid place-items-center py-12">{children}</div>
    </main>
  );
}
