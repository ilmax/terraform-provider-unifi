# AGENTS.md

This repository is scaffolded by Codex. The following instructions are mandatory for all Codex runs in this repo.

## Repository Rules

- Use **local git** only. Do not use external MCP servers or network fetches unless explicitly approved.
- Create a **plan** and request approval before starting multi-step work.
- Break work into **small tasks** and **commit after each task** with **semantic-version commit messages** (e.g., `chore: ...`, `feat: ...`, `docs: ...`, `test: ...`, `ci: ...`).
- Use the **Terraform Plugin Framework** (plugin model) and follow provider best practices.
- Prefer **smaller, focused resources** over monolithic ones.
- The provider must support **api_key authentication only** (no username/password).
- Implement **tests** for new provider functionality.
- Use the **latest Go version** available in toolchains and workflows.

## Project Scope (from owner)

- Scaffold a Terraform provider for UniFi based on `github.com/ilmax/unifi-client-go@v0.1.0`.
- Add initial resources: `device`, `network`, `wifi`, `firewall`, `wan`.
- Add documentation for provider and resources.
- Add a GitHub workflow to **release** the provider.

## Collaboration

- If anything is unclear, ask for clarification **before** making assumptions.
