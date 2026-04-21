# FanAPI Frontend Redesign Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build a production-ready replacement frontend in `web/app` using React, Vite, Tailwind CSS, and shadcn/ui, then migrate FanAPI's core user and admin flows and verify launch readiness.

**Architecture:** A single React application serves all four product surfaces while sharing one design system, one component library, and one API/auth foundation. The delivery is phased so the shared system and the highest-value surfaces are stabilized before secondary roles are migrated.

**Tech Stack:** React, Vite, TypeScript, React Router, Tailwind CSS, shadcn/ui, Playwright

---

### Task 1: Create the new frontend foundation

**Files:**
- Create: `web/app/*`
- Create: `web/app/package.json`
- Create: `web/app/vite.config.ts`
- Create: `web/app/tsconfig.json`
- Create: `web/app/src/app/*`
- Create: `web/app/src/styles/*`

- [ ] Initialize the new app in `web/app` with React, Vite, and TypeScript.
- [ ] Add Tailwind CSS and configure the global CSS entry for semantic design tokens.
- [ ] Initialize shadcn/ui and generate `components.json`.
- [ ] Install the core shared primitives needed for phase 1.
- [ ] Verify the new app starts independently in development.

### Task 2: Build the shared platform layer

**Files:**
- Create: `web/app/src/lib/api/*`
- Create: `web/app/src/lib/auth/*`
- Create: `web/app/src/lib/utils/*`
- Create: `web/app/src/routes/*`
- Create: `web/app/src/components/shared/*`

- [ ] Implement API clients compatible with the current Go backend.
- [ ] Implement token storage and role-aware auth helpers.
- [ ] Implement route guards for user/admin/agent/vendor.
- [ ] Create shared product wrappers for page headers, filter bars, tables, cards, dialogs, and empty states.
- [ ] Verify role routing and authenticated fetch behavior against the existing backend.

### Task 3: Implement the shared visual system

**Files:**
- Modify: `DESIGN.md`
- Create: `web/app/src/styles/theme.css`
- Create: `web/app/src/components/shared/theme/*`
- Create: `web/app/src/components/shared/layout/*`

- [ ] Encode the approved semantic tokens into the new frontend theme files.
- [ ] Implement shared layout shells for user and admin.
- [ ] Implement common density, spacing, and typography utilities.
- [ ] Add light and dark mode support using one semantic token system.
- [ ] Verify visual consistency across sample screens before feature migration.

### Task 4: Migrate user core flows

**Files:**
- Create: `web/app/src/features/auth/*`
- Create: `web/app/src/features/user-dashboard/*`
- Create: `web/app/src/features/channels/*`
- Create: `web/app/src/features/apikeys/*`
- Create: `web/app/src/features/billing/*`
- Create: `web/app/src/features/profile/*`
- Create: `web/app/src/pages/user/*`

- [ ] Rebuild login, register, and password recovery in the new auth system.
- [ ] Rebuild the user shell and dashboard/home.
- [ ] Rebuild channels/models, API keys, billing/orders, and profile.
- [ ] Migrate the highest-priority support screens such as logs, tasks, or playground based on backend compatibility and shared patterns.
- [ ] Verify the user core flow end to end against the current backend.

### Task 5: Migrate admin core flows

**Files:**
- Create: `web/app/src/features/admin-auth/*`
- Create: `web/app/src/features/admin-dashboard/*`
- Create: `web/app/src/features/admin-channels/*`
- Create: `web/app/src/features/admin-users/*`
- Create: `web/app/src/features/admin-billing/*`
- Create: `web/app/src/features/admin-tasks/*`
- Create: `web/app/src/features/admin-logs/*`
- Create: `web/app/src/features/admin-settings/*`
- Create: `web/app/src/features/admin-vendors/*`
- Create: `web/app/src/pages/admin/*`

- [ ] Rebuild the admin login and admin shell.
- [ ] Rebuild dashboard, channels, users, billing, tasks, logs, settings, and vendors using shared page patterns.
- [ ] Standardize all admin list views, filters, and actions under one data-view pattern.
- [ ] Verify the admin core flow against the current backend.

### Task 6: Extend the system to agent and vendor

**Files:**
- Create: `web/app/src/layouts/agent/*`
- Create: `web/app/src/layouts/vendor/*`
- Create: `web/app/src/features/agent/*`
- Create: `web/app/src/features/vendor/*`
- Create: `web/app/src/pages/agent/*`
- Create: `web/app/src/pages/vendor/*`

- [ ] Build agent and vendor shells by reusing the shared system.
- [ ] Port the current agent and vendor core screens with minimal new visual patterns.
- [ ] Verify these roles inherit the same design language and auth strategy.

### Task 7: Add launch-readiness verification

**Files:**
- Create: `web/app/playwright.config.ts`
- Create: `web/app/tests/e2e/*`
- Modify: `web/app/package.json`
- Modify: `README.md`

- [ ] Add end-to-end tests for the highest-risk user and admin flows.
- [ ] Add build and preview scripts required for verification.
- [ ] Document how to run the new frontend locally and how it connects to the existing backend.
- [ ] Run build, smoke checks, and end-to-end verification.
- [ ] Produce a final launch-readiness checklist.
