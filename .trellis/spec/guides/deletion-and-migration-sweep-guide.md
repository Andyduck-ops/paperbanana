# Deletion & Migration Sweep Guide

> **Purpose**: When you delete files, rename APIs, or collapse N→M variants, sweep the *whole repo* — not just the surfaces the DoD lists.

---

## The Problem

**Deletions leak into non-code surfaces.**

Code compiles cleanly because the deleted file is no longer imported, but stale references survive in places the type-checker never reads:

- Spec docs (`.trellis/spec/**/*.md`)
- Architecture docs (`web/DESIGN.md`, `web/design-system/**/MASTER.md`)
- E2E tests (`web/e2e/*.spec.ts`) that hit DOM via selectors, not imports
- i18n locale files (`web/src/i18n/locales/*.json`) — JSON keys aren't validated by TS
- README, CHANGELOG, `.trellis/tasks/*/research/*.md`
- Gitignored design-system docs (invisible to PR review)
- Inline comments referencing the dead identifier

These rot silently. The next agent reads the spec, believes a deleted file still exists, and writes broken code.

---

## Before Declaring "Deletion Complete"

### Step 1: Enumerate Every Identifier You Killed

When you delete files or fields, list **every** symbol the change removed:
- File paths (e.g., `themes/qi-baishi.css`)
- Module exports (e.g., `useTheme`, `ThemeSelector`)
- Attributes (e.g., `data-theme`)
- Enum values (e.g., `'qi-baishi' | 'pop-anime'`)
- i18n keys (e.g., `theme.options.qi-baishi.label`)

### Step 2: Run the Sweep

```bash
# Build a grep alternation of every dead identifier
DEAD='qi-baishi|pop-anime|rococo|japanese-bw|themes/base\.css|useTheme|data-theme'

# Sweep the repo (NOT just src/)
grep -rEn "$DEAD" \
  .trellis/spec/ \
  web/DESIGN.md \
  web/design-system/ \
  web/e2e/ \
  web/src/i18n/locales/ \
  web/src/ \
  README.md 2>/dev/null
```

### Step 3: Triage Each Hit

| Kind of hit | Action |
|-------------|--------|
| Real reference (still wired) | The deletion is incomplete — fix it |
| Stale doc / spec mention | Patch the doc |
| Dead i18n key | Delete the key + any test asserting on it |
| Dead e2e selector test | Delete the test |
| Explicit "Forbidden:" notice | Keep it — it's a guardrail |

### Step 4: Re-grep to Zero

After patching, re-run the same grep. The only hits remaining should be **explicit forbidden-pattern notices** (e.g., a spec line that says *"`data-theme` is forbidden"*).

---

## Surfaces Often Missed

| Surface | Why missed | How to catch |
|---------|------------|--------------|
| `.trellis/spec/<layer>/*.md` | DoD usually lists 1-2 specs, not all | Grep `.trellis/spec/` for every dead identifier |
| `web/DESIGN.md` | Top-level design doc, no import graph | Always grep the file by name |
| `web/design-system/**/MASTER.md` | Often gitignored, invisible to PR review | Grep manually before declaring done |
| `web/e2e/*.spec.ts` | Playwright specs aren't run in default vitest | Grep e2e/ explicitly |
| `web/src/i18n/locales/*.json` | TS doesn't type-check JSON keys | Grep locale files; check for orphan key tests in `i18n/index.test.ts` |
| Task `research/*.md` | Generated artifacts, not refreshed | Grep `.trellis/tasks/`, decide per-task whether to patch or annotate |
| Comments with `// removed X` | Backwards-compat hacks linger | Grep for the deleted identifier in `// ` comments |

---

## The Sweep Is Part of "Done"

A deletion is **not done** when:
- Code compiles ✓
- Tests pass ✓
- The spec sweep is incomplete ✗

Add this to your task DoD:

> **DoD-N**: After deletion, `grep -rEn "<dead-identifier-alternation>" .trellis/ web/ README.md` returns only explicit forbidden-pattern notices.

---

## Case Study: themes-light-dark-only (2026-04-26 → 04-27)

**Context**: Collapsed 14 art-theme stylesheets to 2 anchors (`claude-light.css` + `linear-dark.css`). Removed `data-theme` attribute, `useTheme` hook, `ThemeSelector` component, 12 CSS files.

**What the DoD listed for spec patches**: `design-principles.md`, `references/README.md`. Two files.

**What `/trellis-check` actually found drifted** (next day):
- `frontend/state-management.md` — appStore still described persisting `theme` field
- `frontend/component-patterns.md` — entire CSS Theme System section described 14 files
- `frontend/directory-structure.md` — themes/ listing still showed deleted files; `useTheme.ts` listed; `useAppStoreAdapter.ts` listed despite deletion
- `frontend/conventions.md` — Import Organization example imported `themes/base.css` + `themes/workspace.css`
- `web/DESIGN.md` — full dual-axis architecture, 14-theme table, Fraunces/IBM Plex typography (real fonts are Source Serif 4 / Inter Variable)
- `web/design-system/paperbanana/MASTER.md` — gitignored, missed by every PR review
- `web/e2e/full-ui.spec.ts` — dead test asserting on `data-theme` attribute and 5 ThemeSelector buttons
- `web/src/i18n/locales/{en,zh}.json` + `i18n/index.test.ts` — orphan `theme.options.qi-baishi/pop-anime/rococo/japanese-bw/academic` keys

**Lesson**: The DoD asked for 2 files. Reality needed 6 + 1 dead test + i18n cleanup. The `/trellis-check` skill caught it because it grep-swept `.trellis/spec/frontend/` for forbidden tokens — but only because someone ran the check.

**Adopted**: this guide. Future deletion tasks should run the sweep *as part of finishing*, not as cleanup the next day.

---

## Checklist

Before marking a deletion/migration task complete:

- [ ] Listed every dead identifier (files, exports, attributes, enum values, i18n keys)
- [ ] Ran the grep sweep across `.trellis/spec/`, `web/DESIGN.md`, `web/design-system/`, `web/e2e/`, `web/src/i18n/locales/`, `web/src/`, top-level docs
- [ ] Triaged each hit (real ref / stale doc / dead i18n / dead test / explicit forbidden notice)
- [ ] Re-greped after patching → zero hits except explicit forbidden notices
- [ ] If any orphan tests/keys survive, opened a follow-up task at appropriate priority
