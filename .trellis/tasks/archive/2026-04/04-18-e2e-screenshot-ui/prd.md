# E2E Screenshot Testing & UI Adjustment

## Goal

Run full-stack verification: backend tests, frontend tests, start real services, run Playwright E2E tests with screenshots, analyze visual output, and adjust UI based on findings.

## Requirements

### R1: Backend Test Run
- Run `go test ./...` for all backend packages
- Document any new failures

### R2: Frontend Test Run
- Run `npm run test:run` in `web/`
- Document any new failures caused by previous refactoring

### R3: Start Real Services
- Start Go backend server (`go run ./cmd/server --config ./configs/config.yaml`)
- Start Vite frontend dev server (`npm run dev` in `web/`)
- Verify both are healthy

### R4: E2E Screenshot Testing
- Run Playwright E2E tests: `npx playwright test e2e/`
- Capture screenshots on failure
- Review screenshot artifacts in `web/playwright-report/` or `web/test-results/`

### R5: UI Adjustment Based on Screenshots
- Analyze E2E results and screenshots
- Fix any visual regressions from previous refactoring (theme system, Button migration, App.tsx refactor)
- Adjust spacing, colors, or layout issues visible in real browser rendering

## Acceptance Criteria

- [ ] Backend tests run and pass (or failures documented)
- [ ] Frontend tests run and pass (or failures documented)
- [ ] Both services start successfully
- [ ] E2E tests produce screenshots
- [ ] Any visual regressions from previous changes are fixed

## Technical Notes

- Backend port: 8080, Frontend port: 5173
- Playwright baseURL: `http://localhost:5173`
- Database: SQLite (`.paperbanana/paperbanana.db`)
- Use `scripts/start-all.ps1` on Windows or manual background processes
