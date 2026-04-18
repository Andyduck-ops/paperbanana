# System UI Optimization with DESIGN.md

## Goal

Systematically optimize the PaperBanana web frontend using the DESIGN.md concept from awesome-design-md (inspired by RunwayML's cinematic creative-tool UI). Create a unified, maintainable design system that fixes current UI architecture problems.

## Requirements

### R1: Create PaperBanana DESIGN.md
- Create `web/DESIGN.md` following the Stitch 9-section format
- Base aesthetic on RunwayML (cinematic dark UI, media-rich layout) adapted for PaperBanana's art-theme system
- Define color palette, typography, component stylings, layout principles, depth/elevation, responsive behavior
- Include Agent Prompt Guide section for future AI coding sessions

### R2: Unify Theme System
- Merge the three parallel theme systems (art theme, dark/light mode, dynamic theme loading) into one coherent system
- `data-theme` controls art direction (hue/brand personality)
- `data-color-scheme` controls light/dark luminance
- Eliminate CSS variable conflicts and priority confusion
- Ensure no FOUC (flash of unstyled content) on page load

### R3: Refactor App.tsx
- Extract routing logic to a dedicated Router component
- Extract providers/wrappers to AppProviders
- Reduce App.tsx from 500+ lines to under 150 lines
- Keep the custom "router" pattern (no React Router dependency per current architecture) but make it clean

### R4: Clean Component Architecture
- Decide on Atomic Design: either commit to it (use atoms in organisms) or abandon the directory structure
- Move orphaned components from `components/` root into proper directories
- Remove/deprecate unused legacy components
- Ensure barrel exports (`index.ts`) are consistent

### R5: Sync Design Documentation
- Update or replace `web/design-system/paperbanana/MASTER.md` to match actual code
- Update `.trellis/spec/ui-design/design-principles.md` with new decisions

## Acceptance Criteria

- [ ] `web/DESIGN.md` exists and follows 9-section Stitch format
- [ ] Theme system uses exactly 2 attributes: `data-theme` + `data-color-scheme`, no conflicts
- [ ] App.tsx is under 150 lines, with clear separation of concerns
- [ ] All components in `components/` root are moved to proper subdirectories
- [ ] `MASTER.md` reflects actual design tokens in code
- [ ] `npm run typecheck` (or equivalent) passes in `web/`
- [ ] `npm run lint` passes in `web/`
- [ ] No visual regressions in core Workspace and generation flow

## Technical Notes

- Tailwind CSS v4 is used (`@theme` directive in `themes/base.css`)
- OKLCH color space is already used for dark mode; preserve this
- Art themes: academic, qi-baishi (default), pop-anime, rococo, japanese-bw
- Font loading in `index.html` needs optimization (currently loads 5 Google Fonts)
- Keep i18n intact; update strings in both `en.json` and `zh.json` if new copy is added
- Follow existing frontend spec: explicit Props interfaces, barrel exports, no path aliases
