# CineForge Product and Architecture Documentation

This directory is the decision baseline for CineForge. The documents describe a production product, not implemented software. If documents conflict, the more specific normative specification controls; unresolved product conflicts are decided in the PRD and recorded below.

## Document map

| Document | Purpose | Status |
|---|---|---|
| [Product Requirements Document](../prd.md) | Vision, users, journeys, requirements, priorities, success and launch acceptance | Canonical |
| [Competitive Landscape](non-technical/competitive-landscape.md) | First-party research, feature matrix, differentiation and strategic recommendations | Research baseline |
| [Node and Workflow Specification](node-and-workflow-spec.md) | Typed graph, run evidence, provider adapters, billing and execution semantics | Normative |
| [System Design](system-design.md) | Stack, components, deployment, storage, collaboration, NLE and scaling | Target architecture |
| [Security and Trust](security-and-trust.md) | Threat model, privacy, provider governance, safety and verification | Launch baseline |
| [Delivery, Reliability and Cost](delivery-reliability-and-cost.md) | Stages, staffing, SLOs, capacity, FinOps, tests and risk register | Execution baseline |
| [Identity Platform Setup](guides/identity-platform-setup.md) | Passwordless email, Google, TOTP, emulator, and production setup | Implementation guide |

## Core product sentence

**CineForge is an AI-native story production system in which a persistent, versioned Story World drives multimodal generation, continuity, collaboration, editing, and export.**

## Locked decisions

| ID | Decision | Rationale |
|---|---|---|
| D-001 | Solo creators and 2–10 person indie teams are equal primary audiences | Avoids treating collaboration or accessible automation as an afterthought |
| D-002 | Creator Mode and Studio Mode share one project/domain model | Prevents incompatible simple/pro formats and allows users to grow into depth |
| D-003 | Story World canon is separate from canvas presentation | Narrative truth must survive layout changes and appear in multiple views |
| D-004 | Published Canon Snapshots and run inputs are immutable | Historical output must remain explainable after story or model changes |
| D-005 | Google is the initial model source, not an exclusive architecture | Speeds initial delivery while preserving supplier choice and resilience |
| D-006 | Models integrate through reviewed provider adapters and a capability catalog | Prevents provider shapes/IDs from contaminating projects |
| D-007 | No unrestricted marketplace, BYOK, or executable custom nodes at launch | Controls security, support, policy and reproducibility risk |
| D-008 | Go runs the control/workflow/media orchestration plane | Efficient concurrency and operational simplicity; FFmpeg handles media computation |
| D-009 | Managed-first GCP, with Temporal Cloud for durable workflows | Minimizes operations while supporting long-running recoverable work |
| D-010 | Web-first hybrid: browser proxy editing, authoritative cloud render | Gives broad access without trusting browser codecs for master output |
| D-011 | Full launch NLE with core interchange; deep proprietary round-trip later | Keeps the end-to-end promise while bounding compatibility scope |
| D-012 | Live collaboration uses CRDTs, but privileged domain actions use the API | Concurrent editing must not bypass authorization, canon or billing rules |
| D-013 | Billing is seat plus usage with append-only reservation/settlement ledger | Aligns recurring team value and variable supplier cost with auditable charging |
| D-014 | Projects are private and excluded from training by default | Required trust posture for unreleased creative IP |
| D-015 | Commercial-safe mature policy | Supports serious stories while bounding payments, legal and abuse risk |
| D-016 | Initial global service uses a primary US region; EU is first residency expansion | Matches year-one scale without pretending active-active residency exists |
| D-017 | Year-one design envelope is 10k MAU / ~1k live sessions | Provides a concrete load and cost target without premature hyperscale architecture |
| D-018 | Internal lineage is always retained; C2PA-ready exports add portable provenance | External credentials complement rather than replace product evidence |

## Shared glossary

- **Asset:** A logical media or document item.
- **Asset Version:** Immutable bytes/structured value plus metadata and lineage.
- **Canon:** User-approved Story World facts and reference choices.
- **Canon Snapshot:** Immutable published canon version captured by runs.
- **Capability Snapshot:** Immutable description of a provider/model's supported operation, limits, price version and policy class.
- **Creator Mode:** Guided outcome-oriented interface over the common engine.
- **Entity:** Character, location, prop, faction, creature, vehicle, motif, or other story-world subject.
- **Graph:** Typed nodes and edges describing production computation.
- **Lineage:** Directed history of source versions, runs and transformations leading to an asset.
- **Node Definition:** Versioned registered contract for inputs, outputs, parameters and execution behavior.
- **Provider Adapter:** Isolated translation and lifecycle implementation for an approved external model provider.
- **Run:** Immutable authorized execution attempt for a node/graph slice.
- **Scene:** Narrative unit with time, place, participants, intent and continuity.
- **Shot:** Planned camera/audio unit serving a scene beat.
- **Studio Mode:** Full graph, collaboration, review, policy and production controls.
- **Take:** Candidate asset or composition for a shot.
- **Timeline Version:** Immutable editorial snapshot used for review or render.

## Requirement and terminology rules

- Requirement IDs live in `prd.md`; supporting documents reference rather than redefine priority.
- “Must” denotes normative behavior. “Should” denotes a preferred design that can change through a recorded decision. “May” denotes optional behavior.
- “Latest” is a UI convenience. Any run, approval, render or export resolves an immutable version.
- “Delete” means follow the documented retention/deletion workflow; it never means silently rewrite financial or security evidence.
- “Provider” and “model” are not interchangeable: one provider can expose many versioned capabilities.
- “Consistency” must name its dimension: identity, style, scene, temporal, narrative, or editorial.
- Marketing claims from competitors are labeled as such and are not architecture evidence.

## Change process

1. Update the canonical requirement or locked decision first.
2. Identify affected schemas, runs, assets, security controls, costs, tests, migrations, and historical compatibility.
3. Add or modify a dated decision record in this table when a locked decision changes.
4. Update cross-document links and the requirement traceability table.
5. Never rewrite competitor observations as current facts without rechecking the first-party source and access date.

## Review cadence

- Product requirements and delivery risks: monthly during build.
- Provider capabilities, terms, data policy, and prices: before enablement and at least quarterly.
- Threat model: for each new trust boundary/capability and quarterly.
- SLOs, capacity, and unit economics: each release stage and monthly after paid beta.
- Disaster recovery, provider outage, and billing reconciliation exercises: at the cadence defined in the delivery plan.
