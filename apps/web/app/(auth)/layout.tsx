import { Brand } from "@/components/brand";
import { ThemeToggle } from "@/components/theme-toggle";

export default function AuthLayout({ children }: { children: React.ReactNode }) {
  return (
    <main className="auth-layout">
      <div className="auth-ambient" />
      <header>
        <Brand />
        <ThemeToggle />
      </header>
      <div className="auth-stage">
        <aside>
          <span>ONE PRODUCTION TRUTH</span>
          <blockquote>
            “Develop the world, direct the shots, and finish the story—with every decision
            traceable.”
          </blockquote>
          <p>Private by default · No password required</p>
        </aside>
        {children}
      </div>
    </main>
  );
}
