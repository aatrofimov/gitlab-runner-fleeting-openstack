---
name: gitlab
description: Use when deploying, configuring, backing up, or operating the self-hosted GitLab, its container registry, or GitLab Runners and CI in this project.
---
# GitLab

Self-hosted GitLab runs via docker compose on a dedicated server (managed by the `gitlab` role).

## Deployment
- Compose file: `docker/gitlab/docker-compose.yml` (or role template).
- Data on named volumes (config/logs/data); never lose them.
- Registry is GitLab's built-in registry (enabled in `gitlab.rb`).
- TLS via reverse proxy (nginx + certbot) + Let's Encrypt.

## Operations
- Backups: `gitlab-backup` plus volumes; schedule it.
- Upgrades: read release notes, back up first, follow the upgrade path.

## Runners
- Build servers are ephemeral, powerful VMs registered as GitLab Runners (`docker` executor).
- Register via Ansible (`gitlab_runner` role); use tags to route jobs.
