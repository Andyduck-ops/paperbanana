# PaperBanana Design System

> A DESIGN.md for an AI-powered academic figure generation workspace.
> Grounded in PaperBanana's academic product constraints, Trellis UI spec, and the design-system structure popularized by awesome-design-md.

---

## 1. Visual Theme & Atmosphere

PaperBanana is not a gallery site or generic AI playground. It is a focused academic workspace for turning paper context into publication-ready figures.

The default mood is:
- **Warm scholarly**: paper-like surfaces, restrained contrast, editorial calm
- **Task-forward**: core inputs and the primary generate action dominate the first screen
- **Trustworthy**: UI feels methodical, explicit, and stable rather than playful or magical
- **Quietly expressive**: artistic themes may add character, but the base experience remains legible and serious

The primary reference direction is closer to warm minimal workspaces such as Notion's softness and editorial restraint than to neon AI dashboards. In practice this means:
- Soft surfaces instead of hard chrome
- Strong hierarchy instead of many competing accents
- Serif-forward headings with sober sans-serif body copy
- A single high-emphasis CTA rather than many equal-weight actions

The UI should always satisfy the PaperBanana above-the-fold rule:
- Value proposition visible immediately
- Core input fields visible immediately
- Main generate button visible immediately
- Advanced settings visually secondary and collapsed by default

---

## 2. Color Palette & Roles

### Academic Theme (Default)

| Token | Light Mode | Dark Mode |
|-------|-----------|-----------|
| Primary | `oklch(0.46 0.1 40)` warm brown | `oklch(0.65 0.08 50)` muted amber |
| Secondary | `oklch(0.82 0.05 85)` warm sand | `oklch(0.35 0.03 260)` deep blue-gray |
| Background | `oklch(0.97 0.01 90)` warm off-white | `oklch(0.15 0.02 260)` near-black |
| Foreground | `oklch(0.19 0.02 260)` deep ink | `oklch(0.92 0.01 90)` warm white |
| Accent | `oklch(0.58 0.08 150)` sage green | `oklch(0.65 0.1 150)` brighter sage |
| Muted | `oklch(0.93 0.01 90)` paper tone | `oklch(0.25 0.02 260)` elevated surface |
| Border | `oklch(0.76 0.02 70)` warm gray | `oklch(0.30 0.02 260)` subtle divider |
| Card | `rgba(255, 251, 244, 0.92)` translucent paper | `oklch(0.20 0.02 260)` elevated card |

### Status Colors (Theme-Agnostic)

| Status | Color |
|--------|-------|
| Success | `oklch(0.61 0.16 145)` |
| Warning | `oklch(0.75 0.16 85)` |
| Error | `oklch(0.59 0.19 25)` |
| Info | `oklch(0.57 0.12 245)` |
| Running | `oklch(0.66 0.13 210)` |
| Pending | `oklch(0.63 0.04 80)` |

### Semantic Roles

| Role | Purpose |
|------|---------|
| Primary | Main CTA, focused highlights, active states |
| Secondary | Background accent washes and supporting surfaces |
| Background | Page backdrop and app field |
| Foreground | Primary reading text and icons |
| Accent | Helpful emphasis, non-primary highlights |
| Muted | Secondary surfaces, pills, neutral controls |
| Border | Dividers, field outlines, card edges |
| Card | Elevated content surfaces |

### Color Rules
- Use OKLCH for all theme colors
- Never use raw hex, RGB, or stock Tailwind color utilities in components
- Every color in JSX/CSS should resolve through semantic tokens
- Primary color is reserved for the single most important action in a region
- Error, success, and info colors should be used functionally, never decoratively
- Dark mode is not an inversion; it is a deliberate recoloring with preserved hierarchy

---

## 3. Typography Rules

### Font Stack

| Role | Font | Fallback |
|------|------|----------|
| Heading | Fraunces | Noto Serif SC, serif |
| Body | IBM Plex Sans | Noto Sans SC, sans-serif |
| Mono (code, IDs, durations) | IBM Plex Mono | monospace |

### Typography Behavior

- Headings are short, intentional, and should not read like marketing slogans
- Body copy should be compact and explanatory, not verbose
- Labels must remain visible; placeholders cannot replace labels
- Serif headings are used to create editorial authority, not ornament
- Monospace is reserved for IDs, durations, statuses, and technical metadata

### Type Scale

| Level | Size | Weight | Line Height | Usage |
|-------|------|--------|-------------|-------|
| Display | `clamp(1.35rem, 1.6vw, 1.75rem)` | 700 | 1.0 | Brand name |
| H1 | `1.5rem` | 700 | 1.2 | Page titles |
| H2 | `1.25rem` | 600 | 1.3 | Section headers |
| H3 | `1.125rem` | 600 | 1.35 | Card titles |
| Body | `0.875rem` | 400 | 1.6 | Paragraphs |
| Small | `0.8125rem` | 400 | 1.5 | Descriptions |
| Caption | `0.75rem` | 500 | 1.4 | Labels, meta |
| Eyebrow | `0.68rem` | 700 | 1.0 | Uppercase section labels |

---

## 4. Component Stylings

### Buttons

| Variant | Style |
|---------|-------|
| Primary | `bg-primary text-primary-foreground` rounded-xl, prominent fill, full-width in primary forms |
| Secondary | `bg-muted text-foreground` rounded-xl, quiet support action |
| Ghost | transparent, hover uses muted/accent surface |
| Danger | `bg-status-error text-white` |

All buttons:
- `transition-all duration-200`
- active scale `0.98`
- minimum height `48px` for primary form submission
- never place two equally emphasized primary buttons in the same zone

### Cards

- Border radius: `1rem` to `1.6rem` (generous, friendly)
- Border: `1px solid var(--theme-border)`
- Background: `var(--theme-card)` with subtle transparency
- Shadow: `var(--theme-surface-shadow)` -- soft, colored by theme
- Hover: border color shifts toward primary

### Forms

- Core inputs are stacked vertically in task order
- Method/context field appears before figure brief field
- Reference upload is secondary and more compact than text inputs
- Advanced configuration stays collapsed until intentionally opened
- Validation and action errors appear inline and in plain language

### Navigation

- Header stays compact and should not visually outrank the main form
- History, settings, projects, language, and theme controls are utility actions
- Theme selection belongs in settings or a secondary control zone, never as the visual center of the header

### Inputs

- Border radius: `0.75rem` to `1rem`
- Focus: `outline: 2px solid var(--color-primary)` with `outline-offset: 2px`
- Disabled: opacity 0.6, cursor not-allowed

### Stage Cards (Pipeline Progress)

- Running: pulsing border, spinning icon
- Complete: green accent, checkmark icon, duration badge
- Error: red accent, X icon, expandable error details with retry button
- Pending: muted, initial letter icon

---

## 5. Layout Principles

### Spacing Scale

Based on `0.25rem` (4px) increments:
- `0.5rem` (8px) -- tight gaps
- `0.75rem` (12px) -- component padding
- `1rem` (16px) -- standard padding
- `1.5rem` (24px) -- section gaps
- `2rem` (32px) -- large sections
- `3rem` (48px) -- page sections

### Grid

- Header: compact utility band with strong left identity and subdued right utilities
- Workspace: single centered composition, max-width `72rem`
- Hero/form region: max-width `52rem`, centered, above the fold
- Settings: 2-column grid on desktop, single on mobile

### Above-The-Fold Rules

- The first viewport must contain the product value proposition, both core text fields, and the primary generate button
- Decorative empty states must not push the form below the fold
- Progress, results, and candidate comparison can take more space after generation begins
- On desktop, the composition should feel like a focused editorial workbench, not a dashboard mosaic

### Z-Index Hierarchy

| Layer | Z-Index | Element |
|-------|---------|---------|
| Base | 0 | Content |
| Elevated | 10 | Cards, buttons |
| Drawer backdrop | 60 | Blur overlay |
| Drawer | 70 | Settings, history panels |
| Modal | 80 | Export, wizard |
| Toast | 90 | Notifications |

---

## 6. Depth & Elevation

### Shadows

- Surface: `0 18px 60px rgba(43, 32, 20, 0.08)` -- soft, warm
- Card hover: `0 10px 24px color-mix(in srgb, var(--theme-primary) 18%, transparent)`
- Drawer: `-28px 0 80px rgba(20, 18, 16, 0.12)` -- directional
- Modal: `0 25px 50px -12px rgba(0, 0, 0, 0.25)`

### Backdrop Blur

- Mode bar: `blur(8px)`
- Drawer backdrop: `blur(10px)` with `rgba(17, 17, 17, 0.14)`
- History panel: `blur(14px)`

---

## 7. Do's and Don'ts

### Do
- Use semantic tokens (`--color-primary`, `--theme-card`) everywhere
- Prefer `color-mix()` for derived colors (hover states, borders)
- Use OKLCH for theme colors
- Add `prefers-reduced-motion` fallbacks
- Test contrast in both light and dark modes
- Use generous border-radius for cards (1rem+)
- Keep the first screen focused on the user's main job
- Collapse advanced controls by default
- Make the primary CTA the heaviest visual element in the form
- Use whitespace to separate core tasks from optional tools

### Don't
- Use raw Tailwind colors (`bg-blue-500`, `text-white`) in components
- Use `!important` in CSS
- Hardcode color values in TypeScript/JSX
- Skip focus-visible styles
- Use centered body text for paragraphs
- Add shadows to every element
- Let theme pickers, empty states, or utility controls dominate the first screen
- Put decorative illustrations ahead of task-critical inputs
- Give multiple actions the same visual priority in one section

---

## 8. Responsive Behavior

### Breakpoints

| Name | Width | Behavior |
|------|-------|----------|
| Mobile | < 640px | Single column, stacked header, full-width drawers |
| Tablet | 640px - 1024px | 2-column grids, condensed header |
| Desktop | > 1024px | Full 3-column header, 2-column settings |
| Wide | > 1400px | Max container width, centered layout |

### Mobile Adaptations
- Header collapses to 2-column grid, identity spans full width
- Theme swatches resize to `3rem`
- Workspace padding reduces to `1rem`
- Settings drawer becomes full-width
- History panel becomes full-width overlay
- Core form remains fully visible without requiring exploration of side panels first

---

## 9. Agent Prompt Guide

When implementing UI changes in PaperBanana:

1. **Always check the theme CSS files first** (`web/src/themes/*.css`) -- every color must reference a CSS custom property
2. **Use semantic class names** from `base.css` (`text-primary`, `bg-card`, `border-border`)
3. **Follow the existing component patterns** -- explicit props interfaces, relative imports, barrel exports
4. **Add i18n keys** to both `en.json` and `zh.json` for all user-facing strings
5. **Preserve the first-screen hierarchy** -- headline, two core fields, and primary CTA must remain immediately visible
6. **Test in multiple themes** -- changes must work in Qi Baishi, Rococo, Pop Anime, Night Mono, and Academic
7. **Respect `prefers-reduced-motion`** -- wrap animations in media query checks
8. **Keep components under 200 lines when practical** -- split large components into smaller pieces when refactoring
9. **Use the design token system** -- never hardcode colors, spacing, or typography values

---

## Theme Implementation Reference

Each theme file (`web/src/themes/<name>.css`) overrides these CSS custom properties:

```css
[data-theme="theme-name"] {
  --theme-primary: oklch(...);
  --theme-secondary: oklch(...);
  --theme-background: oklch(...);
  --theme-foreground: oklch(...);
  --theme-accent: oklch(...);
  --theme-muted: oklch(...);
  --theme-border: oklch(...);
  --theme-card: ...;
  --theme-muted-foreground: oklch(...);
  --theme-primary-muted: color-mix(...);
  --theme-primary-alpha: color-mix(...);
  --theme-background-rgb: r, g, b;
  --theme-surface-shadow: ...;
  --theme-status-success: oklch(...);
  --theme-status-warning: oklch(...);
  --theme-status-error: oklch(...);
  --theme-status-info: oklch(...);
  --theme-status-pending: oklch(...);
  --theme-status-running: oklch(...);
  --theme-font-heading: ...;
  --theme-font-body: ...;
  --theme-preview-primary: oklch(...);
  --theme-preview-bg: oklch(...);
  --theme-preview-accent: oklch(...);
  --theme-option-bg: oklch(...);
  --theme-option-ink: oklch(...);
  --theme-option-accent: oklch(...);
}
```

The `base.css` file defines the default (Academic) theme on `:root` and registers Tailwind v4 `@theme` tokens.
