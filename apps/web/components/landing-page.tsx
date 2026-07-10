import Image from "next/image";
import Link from "next/link";
import Script from "next/script";
import {
  ArrowRight,
  Check,
  Clapperboard,
  Fingerprint,
  Layers3,
  Play,
  Sparkles,
  UsersRound,
} from "lucide-react";
import heroImage from "../../../hero-cineforge-canvas.png";
import workflowImage from "../../../feature-node-workflow.png";
import { Brand } from "@/components/brand";
import { ThemeToggle } from "@/components/theme-toggle";

const journey = [
  [
    "01",
    "Develop the world",
    "Shape characters, places, chronology, and rules into a versioned Story World.",
  ],
  [
    "02",
    "Direct every shot",
    "Turn story beats into visual plans, references, and controlled generative workflows.",
  ],
  [
    "03",
    "Compare the takes",
    "Branch ideas, preserve lineage, review together, and choose what becomes canon.",
  ],
  [
    "04",
    "Finish the story",
    "Conform approved work into a shared timeline and export with production history intact.",
  ],
];

export function LandingPage() {
  const productData = {
    "@context": "https://schema.org",
    "@type": "SoftwareApplication",
    name: "CineForge",
    applicationCategory: "MultimediaApplication",
    operatingSystem: "Web",
    description: "A story-first AI production studio for narrative video.",
  };
  return (
    <main className="landing">
      <Script id="cineforge-product-data" type="application/ld+json">
        {JSON.stringify(productData)}
      </Script>
      <header className="public-nav">
        <Brand />
        <nav aria-label="Main navigation">
          <a href="#product">Product</a>
          <a href="#workflow">Workflow</a>
          <a href="#creators">Creators</a>
          <a href="#teams">Teams</a>
          <a href="#trust">Trust</a>
        </nav>
        <div className="public-actions">
          <ThemeToggle />
          <Link href="/sign-in" className="nav-signin">
            Sign in
          </Link>
          <Link href="/sign-in" className="button button-primary">
            Start creating
          </Link>
        </div>
      </header>

      <section className="hero-section" id="product">
        <Image
          className="hero-art"
          src={heroImage}
          alt="Concept illustration of a connected cinematic production canvas"
          priority
          sizes="100vw"
        />
        <div className="hero-shade" />
        <div className="hero-copy">
          <div className="eyebrow">
            <span /> Early access · Foundation available now
          </div>
          <h1>
            Every story deserves
            <br />
            <em>one production truth.</em>
          </h1>
          <p>
            Develop the world, direct the shots, and finish the story—with every creative decision
            private, versioned, and traceable.
          </p>
          <div className="hero-actions">
            <Link href="/sign-in" className="button button-primary button-large">
              Start creating <ArrowRight size={17} />
            </Link>
            <a href="#workflow" className="button button-glass button-large">
              <Play size={15} fill="currentColor" /> See how it works
            </a>
          </div>
          <div className="hero-proof">
            <span>
              <Check size={14} /> Private by default
            </span>
            <span>
              <Check size={14} /> No training without consent
            </span>
            <span>
              <Check size={14} /> Built for production
            </span>
          </div>
        </div>
      </section>

      <section className="marquee" aria-label="CineForge capabilities">
        <span>STORY WORLD</span>
        <i>◆</i>
        <span>SHOT DESIGN</span>
        <i>◆</i>
        <span>GENERATIVE WORKFLOWS</span>
        <i>◆</i>
        <span>REVIEW</span>
        <i>◆</i>
        <span>FULL NLE</span>
        <i>◆</i>
        <span>PROVENANCE</span>
      </section>

      <section className="section workflow-section" id="workflow">
        <div className="section-intro">
          <span className="section-number">01 / WORKFLOW</span>
          <h2>
            From first idea to final cut,
            <br />
            <em>without losing the thread.</em>
          </h2>
          <p>
            CineForge is designed as a creative operating system—not a pile of disconnected
            generation forms.
          </p>
        </div>
        <div className="journey-grid">
          {journey.map(([n, title, copy]) => (
            <article key={n}>
              <span>{n}</span>
              <h3>{title}</h3>
              <p>{copy}</p>
              <ArrowRight size={17} />
            </article>
          ))}
        </div>
        <figure className="workflow-visual">
          <Image
            src={workflowImage}
            alt="Concept art showing a prompt flowing through a generation node into a video clip"
            sizes="(max-width: 900px) 100vw, 86vw"
          />
          <figcaption>
            Product vision · Workflow execution is planned for a later delivery phase.
          </figcaption>
        </figure>
      </section>

      <section className="section audience-grid">
        <article id="creators" className="audience-card creator-card">
          <div className="icon-orb">
            <Sparkles />
          </div>
          <span>FOR SOLO CREATORS</span>
          <h2>
            Move at the speed
            <br />
            of imagination.
          </h2>
          <p>
            Guided creation, production-ready defaults, and one coherent place to build an ambitious
            story.
          </p>
          <ul>
            <li>Outcome-oriented Creator Mode</li>
            <li>Reusable characters and locations</li>
            <li>Clear cost and quality choices</li>
          </ul>
        </article>
        <article id="teams" className="audience-card team-card">
          <div className="icon-orb">
            <UsersRound />
          </div>
          <span>FOR INDIE TEAMS</span>
          <h2>
            Share the vision.
            <br />
            Keep control.
          </h2>
          <p>
            Roles, approvals, budgets, and traceable handoffs for production teams of two to ten.
          </p>
          <ul>
            <li>Workspace roles and permissions</li>
            <li>Shared canon and review history</li>
            <li>Spend policy and oversight</li>
          </ul>
        </article>
      </section>

      <section className="section trust-section" id="trust">
        <div className="section-intro">
          <span className="section-number">02 / TRUST</span>
          <h2>
            Your unreleased story
            <br />
            <em>stays yours.</em>
          </h2>
        </div>
        <div className="trust-grid">
          <Trust
            icon={<Fingerprint />}
            title="Private by default"
            copy="Projects, assets, and decisions begin inside your workspace boundary."
          />
          <Trust
            icon={<Layers3 />}
            title="Lineage, not mystery"
            copy="Outputs retain their source inputs, versions, model decisions, and transformations."
          />
          <Trust
            icon={<Clapperboard />}
            title="Creator in control"
            copy="AI proposes. You approve canon, winning takes, spending, and the final cut."
          />
        </div>
      </section>

      <section className="final-cta">
        <span>YOUR STORY IS WAITING</span>
        <h2>
          Build the world.
          <br />
          <em>Make it move.</em>
        </h2>
        <p>Start with a private workspace and shape the production from there.</p>
        <Link href="/sign-in" className="button button-primary button-large">
          Enter CineForge <ArrowRight size={17} />
        </Link>
      </section>
      <footer className="public-footer">
        <Brand />
        <p>CineForge is in early access. Product-vision features are delivered in phases.</p>
        <div>
          <a href="#product">Product</a>
          <a href="#trust">Trust</a>
          <Link href="/sign-in">Sign in</Link>
        </div>
        <small>© {new Date().getFullYear()} CineForge</small>
      </footer>
    </main>
  );
}

function Trust({ icon, title, copy }: { icon: React.ReactNode; title: string; copy: string }) {
  return (
    <article>
      {icon}
      <h3>{title}</h3>
      <p>{copy}</p>
    </article>
  );
}
