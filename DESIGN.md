# FanAPI DESIGN.md

## Purpose

This document is the highest-priority frontend and UI contract for FanAPI.
All new frontend work must follow this file before introducing new pages,
patterns, colors, components, or layout behavior.

## Product Direction

FanAPI is a multi-role AI platform with four frontend surfaces:

- User
- Admin
- Agent
- Vendor

The product should feel:

- Clear
- Professional
- Calm
- Consistent
- Dense enough for operational work without feeling crowded

Visual direction:

- Base reference: Cal.com
- Secondary influence: Linear-level precision in spacing, states, and motion
- Explicitly avoid flashy gradients, inconsistent colors, improvised cards, and AI-slop layouts

## Design Principles

1. Clarity over novelty.
2. Consistency over page-level creativity.
3. Shared system before role-specific exceptions.
4. Information density without visual pressure.
5. Components before custom markup.
6. States and edge cases are part of the design, not afterthoughts.

## Experience Model

FanAPI is one product with four role contexts, not four separate design systems.

- User: product-facing, guided, slightly warmer
- Admin: operational, structured, high-signal
- Agent: relationship and workflow-focused
- Vendor: supply-side management, stats and controls

All four surfaces must share:

- The same color tokens
- The same typography system
- The same spacing and radius scale
- The same primitives for forms, tables, cards, dialogs, filters, and empty states

## Visual Tone

Use a restrained SaaS-console aesthetic:

- Neutral surfaces
- Strong typography hierarchy
- Mild shadows
- Clear borders
- Small, intentional accents
- Precise spacing

Do not use:

- Random bright gradients
- Multiple competing accent colors
- Heavy glassmorphism
- Over-rounded playful shapes
- Dashboard cards with inconsistent heights and internal spacing

## Color Rules

Use semantic tokens only. Never hardcode page-level raw colors for UI styling.

Required semantic groups:

- background
- foreground
- muted
- muted-foreground
- card
- card-foreground
- popover
- popover-foreground
- border
- input
- primary
- primary-foreground
- secondary
- secondary-foreground
- accent
- accent-foreground
- destructive
- warning
- success
- ring

Rules:

- One primary accent only
- Status colors are reserved for status meaning
- Border contrast must remain visible in both light and dark mode
- Table row hover, active navigation, selection, and focus must be distinguishable from each other

## Typography Rules

Typography should be stable and practical.

- Headings: compact and confident
- Body text: readable and neutral
- Labels: smaller, supportive, never louder than primary content
- Table text: compact but legible
- Helper text: muted, never low-contrast to the point of being hidden

Rules:

- Use a single font stack for the product
- Limit heading levels used in-app
- Avoid oversized marketing typography in product pages
- Use numeric alignment and tabular styles where data scanning matters

## Spacing, Radius, and Elevation

Use one shared spacing scale across all surfaces.

- Tight spacing for dense admin/data views
- Medium spacing for forms and standard content
- Large spacing only for authentication, onboarding, or empty states

Rules:

- Cards of the same level use the same radius
- Form controls use the same control height scale
- Dialogs, sheets, and dropdowns use consistent corner radius and border treatment
- Shadows stay subtle; borders do most of the separation work

## Layout Rules

Shared layout model:

- Persistent left navigation for desktop role consoles
- Stable top header area
- Predictable page header pattern
- Controlled content width by page type

Page types:

- Console pages: full-width fluid with structured content bands
- Form pages: constrained readable width
- Docs/reference pages: wider readable content layout
- Auth pages: simplified centered layout

Rules:

- No page invents its own shell
- No page invents its own header structure
- Filters, actions, and page titles always appear in the same relationship
- Mobile behavior must be deliberate, not accidental collapse

## Component Rules

All core UI should be built from shadcn/ui primitives and approved wrappers.

Mandatory standardized patterns:

- Page header
- Filter bar
- Data table
- Stat card
- Section card
- Form section
- Empty state
- Loading state
- Error state
- Confirmation dialog
- Side panel / sheet
- Detail view

Rules:

- Reuse component variants before creating new ones
- Wrap raw shadcn primitives when the product needs repeatable behavior
- Prefer composition over one-off page CSS
- No bespoke button styles per page
- No custom table styling per page
- No custom modal styling per page

## Forms

Forms must be uniform across all roles.

Rules:

- Shared field spacing
- Shared label placement
- Shared validation message style
- Shared disabled/loading/submitting patterns
- Shared destructive action treatment

## Data Views

Tables, logs, transactions, channels, and task lists are core product surfaces.

Rules:

- Prioritize scanability
- Keep filters consistent across pages
- Standardize status badges
- Standardize row actions
- Standardize empty, loading, and error rows

## Motion

Motion should support clarity, not spectacle.

Allowed:

- Fast hover transitions
- Subtle dialog and sheet entrance
- Small feedback transitions for selection, focus, and loading

Avoid:

- Large animated gradients
- Excessive staggered entrances in console views
- Decorative animation without UX value

## Dark Mode

Dark mode is first-class, not an afterthought.

Rules:

- All semantic tokens must support dark mode
- Avoid pure black backgrounds
- Preserve border visibility and content hierarchy
- Do not patch dark mode per page with ad hoc overrides

## Engineering Guardrails

Frontend implementation rules:

- New frontend stack: React + Vite + Tailwind CSS + shadcn/ui
- Use the installed shadcn skill and project `components.json` once initialized
- Prefer semantic utility usage over raw color values
- Prefer shared wrappers and feature modules over page-local UI invention
- One frontend app serves all four role surfaces

Forbidden:

- Shipping new Vue pages in the replacement frontend
- Adding a second visual system
- Mixing Element Plus with the new frontend
- Direct page-level hardcoded design decisions that bypass shared tokens

## Delivery Standard

Work is not complete until it is:

- Implemented in the new frontend
- Aligned with this document
- Verified against real backend APIs
- Covered by end-to-end checks for critical flows
- Ready for handoff so later frontend work can follow the same system without redesigning from scratch
