# Contributing Guide

Thanks for your interest in contributing to Payment Gateway! This project powers a production-style, event-driven payment system built with Go and a modern frontend. Contributions of all kinds are welcome — code, docs, tests, and ideas.

## Code of Conduct

By participating in this project, you agree to uphold our standards. Please read the [Code of Conduct](CODE%20OF%20CONDUCT.md) before contributing.

## How to Contribute

- Open an issue to discuss bugs, features, or questions.
- Fork the repo and create a topic branch: `feature/...`, `fix/...`, or `chore/...`.
- Make focused changes with clear tests.
- Open a Pull Request (PR) and fill out the template.

## Development Setup

- Backend: Go (see [README](README.md) for prerequisites and running services)
- Frontend: Next.js (in `frontend/`) — use `pnpm dev`
- Databases and queues via Docker (RabbitMQ, Postgres) or your own instances

## Running Locally

- Follow steps in [README](README.md#getting-started) for environment and migrations
- Start API and workers from `cmd/...`
- Start the frontend in `frontend/`

## Style & Quality

- Go: `gofmt`, idiomatic Go, small functions, clear error handling
- Frontend: ESLint, TypeScript, consistent hooks and components
- Commits: Use Conventional Commits where possible
  - `feat: add settlement retries`
  - `fix(api): correct JWT claim expiry`
  - `docs: add DLQ quickstart`

## Testing

- Go: `go test ./...` for unit tests; integration tests with RabbitMQ use `-tags=integration`
- Frontend: basic component and route tests (if present)
- Prefer small, deterministic tests; mock external systems when feasible

## Pull Request Checklist

- Changes are scoped and easy to review
- Includes tests or clear manual steps
- Updates docs or README where relevant
- CI passes (build, tests, lint)

## Release & Changelog

- Maintainers may squash and merge
- Follow semver for API changes when applicable

## Questions

Have an idea or unsure where to start? Open an issue or propose a small PR — we’re happy to help.
