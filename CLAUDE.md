# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Distributed locking library for MongoDB in Go. Single `lock` package providing exclusive and shared locks backed by a MongoDB collection. Originally by Square, maintained by foomo.

## Commands

Always prefer `make` targets over raw commands. Run `make help` for all targets.

- **`make check`** — full pipeline: tidy, generate, lint, test
- **`make lint`** — run golangci-lint
- **`make lint.fix`** — auto-fix lint violations
- **`make test`** — run tests with coverage (`-tags=safe`)
- **`make test.race`** — run tests with race detection
- **`make tidy`** — run `go mod tidy`
- **Single test:** `go test -tags=safe -v -run TestName ./...`

Tests require MongoDB on `localhost:27017` (override with `TEST_MONGO_URL`, `TEST_MONGO_DB` env vars).

## Architecture

Four source files, one package (`lock`):

- **lock.go** — `Client` struct wrapping a `*mongo.Collection`. Provides `XLock` (exclusive), `SLock` (shared), `Unlock`, `Renew`, `Status`, and `CreateIndexes`. Uses upserts with duplicate-key detection (code 11000) for atomic lock acquisition.
- **purge.go** — `Purger` interface that finds and removes expired locks (TTL <= 0). Purge unlocks all locks sharing a lockId with any expired lock.

Key design details:
- Resources are documents keyed by a unique `resource` field. Each document holds one exclusive lock and an array of shared locks with a count.
- `lockId` groups related locks for bulk unlock/renew. Knowing only the resource name is not enough to unlock.
- TTL is stored as `expiresAt` timestamp; -1 means no TTL, 0 means expired. Renewal requires TTL >= 1s to avoid races with the purger.
- Locks are sorted by MongoDB ObjectId (not CreatedAt) for deterministic ordering.

## Git Conventions

- **Conventional Commits** enforced by lefthook: `type(scope): message`
  - Types: `feat`, `fix`, `chore`, `docs`, `refactor`, `test`, `ci`, `build`, `perf`, `sec`, `style`, `revert`, `wip`
- **Branch names** must use `feature/*` or `fix/*` prefixes
- Build tag: always use `-tags=safe`

## Tooling

- **mise** — tool version manager (installs golangci-lint, lefthook)
- **lefthook** — git hooks (pre-commit lint/format, commit-msg validation, branch naming)
- Linter config in `.golangci.yml`
