# CineForge Product Requirements Document

**Version:** 2.0  
**Status:** Product and architecture baseline  
**Last updated:** 2026-07-06  
**Owners:** Product, Engineering, Design, Trust & Safety  
**Related documents:** [Documentation index](docs/README.md), [competitive landscape](docs/non-technical/competitive-landscape.md), [node specification](docs/node-and-workflow-spec.md), [system design](docs/system-design.md), [security](docs/security-and-trust.md), [delivery plan](docs/delivery-reliability-and-cost.md)

## 1. Executive summary

CineForge is an AI-native production studio for creating narrative video. It combines story development, a persistent Story World, a typed multimodal generation graph, collaborative review, and a full non-linear editor (NLE) in one product.

The product serves two equal primary audiences:

1. **Solo creators**, who need guided automation, good defaults, and one place to move from idea to export.
2. **Indie teams of 2–10**, who need shared canon, roles, review, versioning, cost control, and reliable handoffs.

Generic AI canvases help people connect models, but do not understand whether a character changed clothes too early, a prop reappeared after being destroyed, or an approved scene was regenerated from an obsolete reference. Existing AI filmmaking products reduce script-to-video friction, but commonly hide lineage or provide only shallow workflow control. CineForge's wedge is a **persistent, versioned production truth** that every prompt, asset, shot, approval, and edit can reference.

The product is not a one-click film vending machine. It is a creative operating system: automation accelerates the work, while creators retain authorship and explicit control over canon and final output.

## 2. Vision, mission, and promise

### Vision

Make ambitious visual storytelling accessible without reducing storytelling to disconnected AI clips.

### Mission

Give creators a coherent production environment in which their story world survives every generation, revision, collaborator, model upgrade, and edit.

### Product promise

> Develop the world, direct the shots, compare the takes, and finish the story—with every decision traceable.

### Product principles

- **Story before model.** Users work with characters, scenes, shots, and takes; provider details remain available but secondary.
- **Canon is explicit.** Approved facts and references are versioned, reviewable, and never silently overwritten.
- **AI proposes; creators decide.** Generated structure and assets remain editable, attributable, and reversible.
- **Simple first, deep when needed.** Creator Mode and Studio Mode expose different complexity over the same project model.
- **No silent substitution.** Model fallback, prompt rewriting, safety changes, and cost changes are visible.
- **Every output has lineage.** The system records inputs, canon snapshot, model capability snapshot, parameters, policy result, and transformations.
- **Spend is a production constraint.** Estimate before execution, reserve before spending, settle after completion.
- **The final cut is portable.** Users can export masters, source packages, stems, and core editorial interchange.

## 3. Problem statement

AI video production currently fragments work across writing tools, image generators, video generators, audio services, spreadsheets, review links, and NLEs. This creates five compounding problems:

1. **Continuity drift:** subjects, wardrobe, locations, props, lighting, and chronology vary across shots.
2. **Lost intent:** prompts and references are copied without preserving why an asset exists or which story beat it serves.
3. **Unrepeatable results:** model versions, parameters, hidden prompt expansion, and source files are not captured reliably.
4. **Expensive iteration:** creators discover cost only after failed generations, duplicate work, or regenerate to recover context.
5. **Broken collaboration:** feedback, approvals, and editorial changes live outside the generative workflow.

CineForge must make a multi-model workflow feel like one production system rather than a collection of API forms.

## 4. Users and jobs to be done

### 4.1 Solo creator

Examples: filmmaker, animator, writer-director, YouTube storyteller, music-video creator.

Needs:

- Turn an idea or script into scenes and a workable visual plan.
- Create reusable characters and locations without learning graph programming.
- Generate shots, dialogue, music, and effects with predictable spend.
- Assemble, edit, and export without moving through ten products.
- Reveal advanced controls only when the work demands them.

### 4.2 Indie production team

Examples: director, producer, writer, editor, art director, sound designer, reviewer.

Needs:

- Share one authoritative story world and asset library.
- Assign responsibility and control who can approve canon or spend credits.
- Work concurrently without corrupting canvas or timeline state.
- Compare versions, leave frame-accurate feedback, and record decisions.
- Export into established post-production workflows.

### 4.3 Secondary users

- Creative agencies producing narrative campaigns.
- Animation and previs teams.
- Educators and student film teams.
- Production companies evaluating AI-assisted development.

Enterprise procurement, on-premises deployment, and a public model marketplace are not launch targets.

## 5. Product modes

Modes change presentation and defaults, not the stored project format.

### Creator Mode

- Guided project setup and story intake.
- Recommended workflows expressed as tasks: “design the protagonist,” “board this scene,” “animate approved frames.”
- Curated model recommendations and safe parameter defaults.
- Template-driven canvas with optional graph reveal.
- Automatic proxy generation, naming, grouping, and timeline assembly.
- Plain-language cost estimates and recovery suggestions.

### Studio Mode

- Full node graph and provider/model controls.
- Roles, approvals, assignment, comments, and audit history.
- Workspace budgets, concurrency limits, and provider policies.
- Batch operations, reusable subgraphs, templates, comparison views, and advanced NLE controls.
- Production dashboards for blocked work, review queues, failed jobs, and spend.

A user may switch modes at any time. Switching never deletes or rewrites project state.

## 6. Core domain model

### Workspace

Billing, membership, policies, templates, shared assets, and audit boundary.

### Project

A film, episode, trailer, short, or other deliverable with a default format, visual language, content rating, and production policy.

### Story World

The persistent semantic layer containing story structure, entities, relationships, chronology, rules, and approved references.

### Canon Snapshot

An immutable version of approved Story World facts used by a run. Draft facts can change; a completed run always points to the exact snapshot it consumed.

### Scene, shot, and take

- A **scene** is a narrative unit with place, time, participants, dramatic intent, and continuity state.
- A **shot** is a planned camera/audio unit serving a scene beat.
- A **take** is a generated, imported, or edited candidate for a shot.

### Workflow graph

Typed nodes and edges describing how data and media are produced or transformed. Canvas position is presentation metadata and does not define execution order.

### Asset and asset version

An asset is a logical item; every upload, generation, transformation, and render creates an immutable version with lineage. Deletion follows retention policy and never rewrites historical accounting.

### Timeline

The editable composition of video, audio, captions, graphics, effects, and automation using a rational timebase.

## 7. Primary journeys

### Journey A: guided solo production

1. Create a project from an idea, script, image, or template.
2. Confirm format, genre, audience, visual direction, and spend ceiling.
3. Review AI-extracted characters, locations, props, scenes, and unanswered questions.
4. Approve a first Canon Snapshot and generate reference candidates.
5. Accept a suggested storyboard, adjust shots, and generate low-cost draft frames.
6. Approve keyframes, animate selected takes, and create dialogue/audio.
7. Auto-assemble a timeline, edit, review continuity warnings, and render.
8. Export a master, stems, project interchange, and provenance manifest.

### Journey B: collaborative indie production

1. A producer creates the workspace, budget, roles, and provider policy.
2. A writer imports a script; the team reviews extracted story structure.
3. Art direction approves canonical character and location references.
4. The director creates shots; artists branch and compare takes in parallel.
5. Reviewers comment and approve at asset, shot, scene, and cut levels.
6. The editor works in the shared timeline while generation continues.
7. Updated shots are conformed without losing approved cut timing.
8. The producer reviews usage, resolves blocked jobs, and approves final export.

### Journey C: model change without project breakage

1. A provider model is upgraded or deprecated.
2. Existing runs retain the old model capability snapshot and outputs.
3. New runs show a migration recommendation and any parameter differences.
4. The user explicitly chooses the new model or an approved compatible alternative.
5. The system creates a new take; it never mutates the previous take.

## 8. Functional requirements

Priority definitions: **P0** is required for paid launch, **P1** is the next committed expansion, and **P2** is exploratory.

### 8.1 Account, workspace, and project

- **PRD-001 P0:** Users can create a personal workspace without configuring team administration.
- **PRD-002 P0:** Workspaces support owner, admin, editor, reviewer, and viewer roles.
- **PRD-003 P0:** A project stores format, aspect ratio, frame rate, audio language, rating, style direction, and default cost/quality policy.
- **PRD-004 P0:** Projects and assets are private by default.
- **PRD-005 P0:** Users can duplicate a project or template without sharing mutable history.
- **PRD-006 P0:** Workspace owners can set monthly budgets, per-run approval thresholds, and member spend permissions.
- **PRD-007 P1:** Workspace-level reusable canon entities and brand/story libraries.

### 8.2 Develop and Story World

- **PRD-010 P0:** Import plain text, Markdown, Fountain, PDF, and supported screenplay formats; preserve the original as an immutable source.
- **PRD-011 P0:** Extract acts, episodes, scenes, beats, characters, locations, props, factions, relationships, and timeline events with confidence and source citations.
- **PRD-012 P0:** Never promote extracted or generated facts to canon without explicit user approval.
- **PRD-013 P0:** Entity profiles support descriptions, aliases, traits, physical features, voice, wardrobe states, reference assets, relationships, and scene-specific state.
- **PRD-014 P0:** Scene profiles support participants, location, story time, emotional intent, continuity-in/out, required props, sound notes, and shot list.
- **PRD-015 P0:** Users can diff Canon Snapshots and see affected scenes, shots, nodes, and approved takes before publishing a change.
- **PRD-016 P0:** Continuity checks detect contradictory facts, wardrobe/prop state errors, chronology conflicts, missing coverage, and references derived from obsolete canon.
- **PRD-017 P0:** AI suggestions identify uncertainty and show source or reasoning; they do not masquerade as script facts.
- **PRD-018 P1:** Relationship map, chronology view, arc tracker, and coverage report.
- **PRD-019 P1:** Lockable production rules, including visual grammar, lens language, color script, and prohibited changes.

### 8.3 Reference and asset library

- **PRD-020 P0:** Upload, generate, tag, search, filter, compare, favorite, approve, and retire assets.
- **PRD-021 P0:** Reference sets can contain multiple views, expressions, poses, wardrobe states, lighting references, voice samples, and usage notes.
- **PRD-022 P0:** Every asset version records origin, owner, checksum, media metadata, policy status, rights declaration, transformations, and parent lineage.
- **PRD-023 P0:** Approved references are addressable from prompts and nodes by stable entity reference rather than copied URLs.
- **PRD-024 P0:** Imported media is validated, malware-scanned, normalized, and proxied before processing.
- **PRD-025 P1:** Semantic and visual similarity search across authorized project assets.
- **PRD-026 P1:** Rights-expiration and consent-expiration warnings.

### 8.4 Canvas and workflow graph

- **PRD-030 P0:** Infinite canvas supports pan, zoom, minimap, multi-select, alignment, frames, groups, comments, and search.
- **PRD-031 P0:** Connections are type-checked before save and execution; invalid connections explain the expected and received type.
- **PRD-032 P0:** Users can run one node, a selected dependency slice, or an approved batch.
- **PRD-033 P0:** The UI displays queued, reserved, running, awaiting input, succeeded, partially succeeded, failed, cancelled, and blocked states.
- **PRD-034 P0:** Branching produces independent takes and preserves the source branch.
- **PRD-035 P0:** Compare/select supports side-by-side, overlay, A/B, and approval of a winning take.
- **PRD-036 P0:** Users see estimated price, quality/speed tier, provider, and expected output before a billable run.
- **PRD-037 P0:** Templates and subgraphs are versioned; existing projects keep the version they instantiated.
- **PRD-038 P0:** Runs capture immutable inputs, Canon Snapshot, node definition, provider capability, parameters, expanded prompt, policy result, price decision, and outputs.
- **PRD-039 P0:** Failed work can resume safely without duplicating settled charges or completed outputs.
- **PRD-040 P1:** Batch iterators accept text/media arrays and CSV input with explicit fan-out limits.
- **PRD-041 P1:** Storyboard view and graph view are projections of the same workflow data.
- **PRD-042 P2:** Reviewed third-party workflow packages; no arbitrary executable custom nodes in the hosted product.

### 8.5 Multi-model generation

- **PRD-050 P0:** Support curated text, image, video, speech, music, and sound capabilities through a stable CineForge model catalog.
- **PRD-051 P0:** Google models form the initial catalog; the project schema must not embed Google-specific identifiers or request shapes.
- **PRD-052 P0:** Users may select a model explicitly or use a transparent recommendation based on capability, policy, cost, and quality tier.
- **PRD-053 P0:** Fallback never occurs silently. Automatic fallback requires workspace permission and an exact compatible capability class.
- **PRD-054 P0:** Provider-specific advanced controls are stored in namespaced extensions and validated against the captured capability version.
- **PRD-055 P0:** Safety rejection, provider outage, invalid input, quota exhaustion, timeout, and internal failure are distinct user-visible outcomes.
- **PRD-056 P0:** Provider callbacks and polling are idempotent; duplicate completion cannot create duplicate assets or charges.
- **PRD-057 P1:** Add curated non-Google providers through an adapter conformance and legal/security approval process.
- **PRD-058 P1:** Workspace policy can allow or deny providers and data classes.
- **PRD-059 P2:** BYOK and public model marketplace, subject to a separate security and support design.

### 8.6 Collaboration and review

- **PRD-060 P0:** Canvas, story documents, and timeline support simultaneous editing with presence indicators.
- **PRD-061 P0:** Node and clip soft locks warn about concurrent manipulation without holding hidden server locks indefinitely.
- **PRD-062 P0:** Comments can target project, canon fact, node, asset region, shot, timeline range, or frame.
- **PRD-063 P0:** Approval states are draft, in review, changes requested, and approved, with actor and timestamp.
- **PRD-064 P0:** Only authorized roles may publish canon, approve final takes, change provider policy, or exceed budgets.
- **PRD-065 P0:** Version history supports named checkpoints and restoration as a new version, never destructive rewind.
- **PRD-066 P1:** Review links with expiration, watermark policy, download control, and optional passcode.
- **PRD-067 P1:** Assignments, due dates, review queues, and notifications.

### 8.7 Full NLE

- **PRD-070 P0:** Multitrack video, audio, caption, adjustment, and graphics tracks use rational time values and project timebase.
- **PRD-071 P0:** Editing includes select, blade, trim, ripple, roll, slip, slide, snapping, grouping, nesting, markers, and undo/redo.
- **PRD-072 P0:** Effects include transform, crop, opacity, blend, color controls, LUTs, speed, freeze, reverse, common transitions, and parameter keyframes.
- **PRD-073 P0:** Audio includes gain, pan, fades, automation, ducking, EQ presets, dialogue/SFX/music grouping, and stem export.
- **PRD-074 P0:** Captions support transcription import/generation, correction, styling, timing, and sidecar or burned-in export.
- **PRD-075 P0:** Browser editing uses proxies and waveform/thumbnail caches while retaining original media for final render.
- **PRD-076 P0:** A generated replacement take can conform into an existing shot without discarding trim intent, comments, or markers.
- **PRD-077 P0:** Final cloud render is deterministic for the captured render manifest and reports unsupported effects before starting.
- **PRD-078 P0:** Export MP4/H.264, H.265 where licensed and supported, WebM, MOV-compatible masters, WAV stems, captions, and image sequences where configured.
- **PRD-079 P0:** Export OTIO plus supported FCPXML/EDL mappings with a relink manifest; warn clearly about lossy interchange.
- **PRD-080 P1:** Scene pacing suggestions, audio sync, color match, beat alignment, and gap-fill proposals, all non-destructive.
- **PRD-081 P2:** AAF and deep round-trip of proprietary NLE effects.

### 8.8 Billing and usage

- **PRD-090 P0:** Bill workspace seats plus metered generation, processing, storage, and render usage according to published rules.
- **PRD-091 P0:** Every billable operation follows estimate → authorization → reservation → execution → settlement or reversal.
- **PRD-092 P0:** The usage ledger is append-only and reconciles provider reports, internal jobs, credits, refunds, and invoices.
- **PRD-093 P0:** Provider failures and internal failures do not consume user credits unless a documented usable output was delivered.
- **PRD-094 P0:** Users can inspect cost by project, member, capability, provider, scene, and time period.
- **PRD-095 P0:** Budget limits stop new reservations without interrupting already authorized work unless an administrator cancels it.
- **PRD-096 P1:** Cost-aware batch optimizer and draft/final quality policies.

### 8.9 Trust, safety, and provenance

- **PRD-100 P0:** Apply upload, prompt, reference, and output moderation appropriate to a commercial-safe mature-content policy.
- **PRD-101 P0:** Block sexual exploitation, CSAM, non-consensual intimate imagery, prohibited extremist content, and abusive impersonation.
- **PRD-102 P0:** Require rights/consent declarations for identifiable likeness or voice cloning and record the declaration in asset metadata.
- **PRD-103 P0:** Provide reporting, appeal, preservation, and escalation workflows with least-privilege access.
- **PRD-104 P0:** Generated and transformed assets retain internal provenance; final exports are C2PA-ready and include an optional human-readable production report.
- **PRD-105 P0:** Customer content is not used to train CineForge or provider models without separate explicit consent and compatible provider terms.
- **PRD-106 P0:** Project export and verified deletion are available to workspace owners, subject to billing, fraud, and legal retention.

## 9. Information architecture

1. **Home:** recent projects, tasks, review requests, usage, templates.
2. **Develop:** source material, structure, Story World, continuity, scenes, shots.
3. **Create:** infinite canvas, asset library, model catalog, run queue.
4. **Edit:** timeline, monitor, inspector, media bin, mixer, captions.
5. **Review:** versions, comments, approvals, compare, share.
6. **Workspace:** people, roles, providers, templates, billing, security, audit.

Creator Mode may combine Develop/Create/Edit into a guided flow; the canonical objects remain identical.

## 10. Success metrics

### North-star metric

**Approved narrative minutes produced per active workspace per month**, accompanied by quality guardrails. This measures completed creative progress rather than raw generations.

### Activation

- At least 60% of qualified new projects reach an approved Canon Snapshot.
- At least 40% reach an approved storyboard or three approved shots within seven days.
- Median time from project creation to first useful draft take under 20 minutes, excluding provider delay beyond published ranges.

### Production value

- At least 50% of generated final-cut assets retain a canon reference and complete lineage.
- At least 30% of active workspaces produce a timeline export within 30 days.
- Fewer than 5% of approved shots are regenerated primarily because the system used an obsolete reference.

### Reliability and trust

- 99.9% monthly control-plane availability at GA.
- At least 99.95% of acknowledged state mutations durably recorded.
- 100% of billable settlements linked to an immutable run and ledger reservation.
- No silent model substitution.
- Deletion and access requests completed within published policy timelines.

### Business

- Positive gross margin per paid workspace after provider, render, storage, and payment costs.
- Expansion driven by collaboration and production volume rather than opaque credit expiration.

Metrics are targets to validate, not claims about current performance.

## 11. Non-functional requirements

- **Scale:** design for 10,000 MAU and 1,000 simultaneous editing sessions in year one; horizontally scale stateless services.
- **Responsiveness:** local canvas/timeline interaction at 60 fps on supported hardware; p95 acknowledged collaborative operations below 300 ms within the primary region under target load.
- **Durability:** immutable assets in versioned object storage; database PITR; tested restoration.
- **Accessibility:** WCAG 2.2 AA for core creation, review, billing, and administration paths; media-canvas alternatives where exact parity is impractical.
- **Compatibility:** current evergreen desktop browsers; degraded read/review experience on mobile; desktop shell is later.
- **Observability:** trace every run and render across API, workflow, provider adapter, asset, and ledger identifiers.
- **Localization:** Unicode-safe data model and localization-ready UI; initial product language English.
- **Privacy:** global launch from a primary US region with transparent residency; EU residency is the first regional expansion.

## 12. Business model

- Paid workspace tiers include seats, storage allowance, collaboration features, and support level.
- Generation and heavy rendering use separately metered credits or currency-denominated usage.
- Display estimated consumption before execution and actual settlement afterward.
- Credits must not obscure provider price changes; publish conversion and expiration rules.
- Free evaluation, if offered, receives strict abuse controls, watermark policy, lower concurrency, and capped storage.

## 13. Launch acceptance

Paid launch is accepted only when:

1. Every P0 journey has an automated happy-path test and documented recovery path.
2. Canon publication, generation, asset lineage, collaboration, timeline editing, rendering, export, and billing work as one traceable flow.
3. Security review has no unresolved critical or high findings.
4. Provider adapter tests cover timeout, rejection, duplicate callback, malformed output, cancellation, and quota exhaustion.
5. Backup restore, billing reconciliation, account compromise, provider outage, and render-worker-loss exercises have passed.
6. Product, support, and trust teams have runbooks for failures users can encounter.
7. No P0 requirement depends on a preview provider capability without a documented alternative or launch gate.

## 14. Explicit non-goals for paid launch

- Training a proprietary foundation model.
- Arbitrary user-supplied executable nodes.
- Open model marketplace or BYOK.
- On-premises deployment.
- Mobile authoring parity.
- Complete replacement for high-end finishing, VFX, or DAW software.
- AAF or perfect proprietary-effect round-trip.
- Fully autonomous publication to social platforms.
- Legal determination of copyright ownership; CineForge records provenance and declarations but does not provide legal advice.

## 15. Assumptions and decisions

- Solo creators and indie teams are equal primary audiences.
- Creator Mode and Studio Mode share one domain and storage model.
- Google supplies the initial curated models, but provider adapters are a launch architecture requirement.
- Infrastructure is managed-first on GCP; portability is achieved at product contracts, not by avoiding every cloud primitive.
- The web client edits proxies; authoritative masters render in the cloud.
- Commercial-safe mature content is allowed within the detailed safety policy.
- Customer projects are private and excluded from training by default.
- Requirements will change through versioned decisions; historical runs and exports remain interpretable.

## 16. Open validation questions

These are research hypotheses, not implementation decisions left to engineers:

- Which guided workflow creates the fastest durable activation for solo creators: character-first, script-first, or storyboard-first?
- How much graph complexity should Creator Mode reveal before users become less successful?
- Which continuity warnings creators consider valuable versus intrusive?
- Which initial non-Google provider materially improves conversion or gross margin enough to justify integration?
- What interchange subset covers the majority of target users' real finishing workflows?

Product research must answer these through prototypes and beta evidence without altering the locked architectural boundaries above.
