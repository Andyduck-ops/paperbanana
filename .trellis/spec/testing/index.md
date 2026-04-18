# Testing Guidelines

Index of all testing specifications for the PaperBanana project.

## Documents

| Document | Description |
|----------|-------------|
| [strategy.md](strategy.md) | Testing pyramid, framework choices, and overall strategy |
| [backend-testing.md](backend-testing.md) | Go testing patterns, assertions, mocking, and fixtures |
| [frontend-testing.md](frontend-testing.md) | Vitest + Testing Library patterns, TDD workflow, and known issues |
| [known-gaps.md](known-gaps.md) | Current test reliability issues, untested areas, and tooling gaps |

## Quick Reference

- **Backend runner:** `go test ./...`
- **Frontend runner:** `cd web && npm run test:run`
- **E2E runner:** `cd web && npx playwright test -c playwright.config.ts`
- **No CI pipeline detected** -- all testing is local
- **No enforced coverage thresholds**
