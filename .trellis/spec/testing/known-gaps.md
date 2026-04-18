# Known Testing Gaps

## Security Middleware Untested

Router security middleware is defined but untested and currently unused:

- Authentication middleware
- CORS middleware
- Rate limiting middleware
- Metrics middleware

These exist in the codebase but have no test coverage and are not wired into
the active router.

## Backend Test Suite Red

`go test ./...` currently fails in two packages:

1. **`internal/application/agents/critic`** -- metadata value mismatch
   (`reused_artifact` expected `"false"`, code produces `"true"`) and
   prompt fixture drift
2. **`internal/config`** -- default provider and `EnableWAL` behavior has
   diverged from test expectations

## Frontend Tests Not Runnable from Clean Checkout

Running `cd web && npm run test:run` from a clean checkout fails because
`node_modules` is not committed. After `npm install`, 31 tests still fail
due to the issues documented in [frontend-testing.md](frontend-testing.md).

## E2E Coverage Stale Relative to Current UI

- Playwright config only matches `test-ui.spec.ts`
- Specs under `web/e2e/` are not included in the default Playwright run
- E2E specs have not been updated to reflect current UI changes
  (semantic tokens, component refactoring, prompt label changes)

## No Integration Tests for Key Flows

The following flows lack integration test coverage:

- **Refine flow** -- no integration test for the refine/polish pipeline
- **Asset URL resolution** -- no test verifying asset URLs are constructed
  correctly end-to-end
- **Legacy popover flows** -- no coverage for the old history popover pattern

## No E2E Coverage for Edge Cases

- **Multi-tab scenarios** -- no E2E test for concurrent browser tabs
- **Stale localStorage config** -- no E2E test for what happens when
  localStorage contains outdated provider configuration
- **Network failure recovery** -- no E2E test for backend unavailability

## ESLint Declared but Not Installed

`web/package.json` includes a `lint` script referencing ESLint, but
`eslint` is not installed as a dependency. Running `npm run lint` fails
immediately.

## Vitest Does Not Exclude Playwright Specs

`web/vite.config.ts` does not exclude `*.spec.ts` files from the Vitest
discovery glob. This causes Vitest to attempt executing Playwright test
files, which immediately fail because Playwright's `test()` and
`test.describe()` globals are not available under Vitest.

**Recommended fix:** Add an exclude pattern to the Vitest configuration:

```typescript
// web/vite.config.ts
test: {
  exclude: [
    '**/node_modules/**',
    '**/dist/**',
    '**/e2e/**',
    '**/*.spec.ts',
  ],
}
```
