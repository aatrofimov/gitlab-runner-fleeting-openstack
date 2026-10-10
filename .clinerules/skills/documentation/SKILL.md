---
name: documentation
description: Use when writing, updating, or organizing documentation in this repository (docs/), creating ADRs, or deciding where a new markdown file belongs. Applies the Diátaxis framework and the no-duplication rule.
---
# Documentation

Documentation in this repo follows Diátaxis and lives in `docs/`. Follow these rules.

## Categories (choose exactly one per page)
- `docs/getting-started/` — tutorials: step-by-step from zero.
- `docs/how-to/` — how-to: solve one concrete task.
- `docs/reference/` — reference: facts (variables, roles, architecture). No narrative.
- `docs/explanation/` — explanation: why/how and background; `docs/adr/` for decisions.

## Process
1. Check `docs/README.md` (the index) — keep it updated when adding pages.
2. Choose the single correct category; never duplicate content across categories.
3. Link, don't copy: if a fact already exists (variable, ADR, reference page), link to it.
4. Keep each page focused on one topic.

## ADRs
- Significant decisions go to `docs/adr/NNNN-slug.md` (next number, kebab-case slug).
- Format: Context, Decision, Consequences. Keep concise.
- Never rewrite an existing ADR's decision; add a new ADR that supersedes it.

## Style
- Markdown, `markdownlint`-clean; short sections; code fences with language tags.
