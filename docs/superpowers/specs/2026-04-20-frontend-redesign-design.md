# FanAPI Frontend Redesign Design

## Goal

Replace the current AI-generated Vue frontend with a production-grade React frontend that is visually consistent, operationally clear, and maintainable by non-frontend specialists after handoff.

## Current Problem

The existing frontend under `web/user` mixes multiple role surfaces into one Vue app with ad hoc styling and inconsistent page patterns. The result is:

- inconsistent colors and hierarchy
- uneven spacing and card patterns
- duplicated layout ideas
- weak component discipline
- low confidence for future collaborative frontend work

The product owner explicitly wants a real rebuild, not iterative polishing.

## Approved Direction

- Build a new frontend in parallel at `web/app`
- Use `React + Vite + Tailwind CSS + shadcn/ui`
- Keep one app for user/admin/agent/vendor
- Keep the first phase compatible with the current backend API
- Use `DESIGN.md` as the highest-priority design contract
- Base visual direction on Cal.com with restrained Linear-style precision

## Success Criteria

The redesign is successful when:

1. A new frontend app runs independently from the old Vue app.
2. Core auth, routing, and role shells work against the current Go backend.
3. User and admin core flows are migrated to the new frontend.
4. Shared tokens, layouts, and components are stable enough that agent/vendor can be built on top without changing the system.
5. The result passes end-to-end verification for critical journeys and can be shipped.

## Non-Goals

- Rewriting backend APIs in phase 1
- Preserving the current Vue visual language
- Designing separate visual systems per role
- Introducing multiple frontend frameworks

## Product Surfaces

### User

Primary jobs:

- authenticate
- explore models/channels
- create and manage API keys
- review billing and logs
- use playground and task-based AI generation

### Admin

Primary jobs:

- monitor platform state
- manage channels, users, vendors, settings, billing, logs, and tasks

### Agent

Primary jobs:

- operate within a narrower workflow shell
- review assigned business-facing information

### Vendor

Primary jobs:

- review key usage and supplier-side stats

## Architecture Decision

Create a new frontend app at `web/app` with these layers:

- `src/app`: app bootstrap, providers, router root
- `src/routes`: route trees and guards
- `src/layouts`: user/admin/agent/vendor shells
- `src/features`: business features by domain
- `src/components/ui`: shadcn-generated primitives
- `src/components/shared`: product-level reusable wrappers
- `src/lib`: API clients, auth, formatting, helpers
- `src/styles`: Tailwind entry, token CSS, globals

This keeps one deployment target while preserving role-specific composition boundaries.

## Migration Strategy

### Phase 0: Design and foundation

- define `DESIGN.md`
- define route and layout strategy
- initialize React/Vite/Tailwind/shadcn app
- establish auth, API, theme, and shell foundations

### Phase 1: Shared product system

- shared page header
- filter bar
- table pattern
- card pattern
- form pattern
- dialog/sheet pattern
- loading/empty/error patterns
- light/dark mode tokens

### Phase 2: User + Admin migration

User core:

- login
- register / password recovery
- dashboard/home
- models/channels
- API keys
- billing/orders
- profile
- logs/tasks/playground as priority features

Admin core:

- login
- dashboard
- channels
- users
- billing
- tasks
- logs
- settings
- vendors

### Phase 3: Agent + Vendor migration

- reuse existing shells and components
- add only role-specific feature modules

### Phase 4: Launch readiness

- feature parity checks against legacy flows
- backend compatibility checks
- end-to-end flow verification
- build, preview, and deployment readiness

## Design System Requirements

The design system must standardize:

- theme tokens
- typography
- layout rhythm
- data density rules
- form behavior
- table behavior
- overlays
- navigation
- empty/loading/error states

No business page may bypass the shared system to invent its own visual language.

## Developer Handoff Requirement

The final output must be safe for future implementation by a backend engineer with AI assistance. That means:

- the design rules are written down
- the component library is explicit
- role shells are already established
- page templates already exist
- critical flows already demonstrate the intended standard

## Risks

### Scope sprawl

Rebuilding four role surfaces can expand indefinitely.

Mitigation:

- phase delivery
- user/admin first
- agent/vendor inherit the shared system

### API mismatch

Existing backend shapes may not map cleanly into the new frontend.

Mitigation:

- phase 1 preserves current APIs
- document API improvement suggestions separately after stabilization

### Design drift after handoff

Future contributors may reintroduce ad hoc patterns.

Mitigation:

- `DESIGN.md` as hard contract
- shadcn component discipline
- strong shared wrappers

## Verification Standard

Before declaring launch-ready:

- app builds successfully
- critical routes render correctly
- auth and role guards work
- user/admin critical flows pass end-to-end checks
- visual consistency holds across shared patterns
- dark mode remains usable

## Deliverables

- `DESIGN.md`
- new frontend app under `web/app`
- migrated user and admin core flows
- reusable role shells
- stable shared component system
- end-to-end verification evidence

## Decision

Proceed with a parallel rebuild using React + Vite + Tailwind CSS + shadcn/ui, with user and admin as phase-1 implementation targets and agent/vendor as phase-2 extensions of the same system.
