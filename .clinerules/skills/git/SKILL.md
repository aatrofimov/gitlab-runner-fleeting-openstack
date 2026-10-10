---
name: git
description: Use when committing, branching, merging, or reviewing changes in this repository, or when deciding commit messages and workflow.
---
# Git

## Workflow
- `main` is protected; no direct push. Work on feature branches, merge via PR/merge request.
- Conventional Commits: `feat:`, `fix:`, `docs:`, `chore:`, `refactor:`, `ci:`, `test:`.
- Atomic, meaningful commits.

## Before finishing
- `git status` clean of stray files; respect `.gitignore`.
- Lint and tests pass (`make lint`, `make test`).
