# Git

- `main` is protected: no direct pushes; changes go through feature branches and merge/PR.
- Conventional Commits: `feat:`, `fix:`, `docs:`, `chore:`, `refactor:`, `ci:`, `test:`, `build:`.
- Atomic, meaningful commits.
- Don't commit artifacts, secrets, `.env`, `dist/`, or `clouds.yaml` with real values.
- Check `git status` before finishing and leave no stray files.
