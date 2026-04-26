# PaperBanana Design System

> Two anchor stylesheets — Claude (light) and Linear (dark). No third style.
> Authoritative spec: [`.trellis/spec/ui-design/design-principles.md`](../.trellis/spec/ui-design/design-principles.md) §0 §7.

---

## 1. Visual Theme & Atmosphere

### Mood
Editorial, focused, gallery-like. The UI exists to showcase generated visuals — it should feel like a creative studio or scholarly journal, not a dashboard.

### Density
Medium-low density. Generous whitespace around the canvas. Controls are grouped and collapsible. The generated image is always the hero.

### Design Philosophy
- **Content-first**: Generated artifacts occupy the visual center; chrome recedes.
- **Two anchors only**: Claude/Anthropic light (parchment + Source Serif 4 + terracotta) and Linear dark (marketing-black + Inter Variable + indigo). Adding a third anchor requires updating `design-principles.md` §0 first.
- **Single axis**: Switch is driven by `data-color-scheme="light" | "dark"` only. The legacy `data-theme` attribute is forbidden.
- **Motion with purpose**: Transitions guide attention; no decorative animation.

---

## 2. Color Palette & Roles

Tokens are defined in `web/src/themes/`:
- `tokens.css` — Tailwind v4 `@theme` bridge mapping `--theme-*` to `--color-*`
- `claude-light.css` — `:root` and `[data-color-scheme="light"]` overrides
- `linear-dark.css` — `[data-color-scheme="dark"]` overrides

### Light Anchor (Claude / Anthropic) — `[data-color-scheme="light"]`

| Token | OKLCH | Role |
|-------|-------|------|
| `--theme-background` | `oklch(0.96 0.01 95)` | Parchment canvas (#f5f4ed) |
| `--theme-card` | `oklch(0.98 0.005 95)` | Cards, panels |
| `--theme-muted` | `oklch(0.93 0.01 90)` | Muted backgrounds |
| `--theme-foreground` | `oklch(0.20 0.005 80)` | Primary text (#141413) |
| `--theme-muted-foreground` | `oklch(0.46 0.01 80)` | Secondary text |
| `--theme-primary` | `oklch(0.56 0.16 40)` | Terracotta CTA (#c96442) |
| `--theme-secondary` | `oklch(0.82 0.05 85)` | Secondary surfaces |
| `--theme-accent` | `oklch(0.58 0.08 150)` | Highlights, progress |
| `--theme-border` | `oklch(0.93 0.01 90)` | Dividers |
| `--theme-ring` | `oklch(0.85 0.01 80)` | Ring shadow (no drop shadow) |
| Focus blue | `#3898ec` | Form-field focus ring (only allowed cool color) |

### Dark Anchor (Linear) — `[data-color-scheme="dark"]`

| Token | OKLCH | Role |
|-------|-------|------|
| `--theme-background` | `oklch(0.10 0.005 264)` | Marketing Black canvas (#08090a) |
| `--theme-card` | `oklch(0.13 0.005 264)` | Cards, panels |
| `--theme-muted` | `oklch(0.16 0.005 264)` | Muted backgrounds |
| `--theme-foreground` | `oklch(0.97 0.005 264)` | Primary text |
| `--theme-muted-foreground` | `oklch(0.62 0.01 260)` | Secondary text |
| `--theme-primary` | `oklch(0.58 0.16 285)` | Brand Indigo CTA (#5e6ad2) |
| `--theme-secondary` | `oklch(0.35 0.03 260)` | Secondary surfaces |
| `--theme-accent` | `oklch(0.65 0.10 150)` | Highlights, progress |
| `--theme-border` | `rgba(255,255,255,0.08)` | Semi-transparent white edges |
| `--theme-border-subtle` | `rgba(255,255,255,0.05)` | Hairline separators |

### Status Tokens (both anchors)

| Token | Light | Dark |
|-------|-------|------|
| `--theme-status-success` | `oklch(0.61 0.16 145)` | `oklch(0.70 0.14 145)` |
| `--theme-status-warning` | `oklch(0.75 0.16 85)` | `oklch(0.78 0.14 85)` |
| `--theme-status-error` | `oklch(0.59 0.19 25)` | `oklch(0.68 0.16 25)` |
| `--theme-status-info` | `oklch(0.57 0.12 245)` | `oklch(0.68 0.10 245)` |

### Usage Rules
- **Never** hardcode hex values in components. Use `var(--theme-*)` or Tailwind v4 `bg-primary` / `text-foreground` (mapped via `tokens.css`).
- **Never** reference `data-theme` — it is forbidden. Only `data-color-scheme="light" | "dark"` exists.
- Adding a third stylesheet requires editing `design-principles.md` §0 first; the gating procedure is in that file.

---

## 3. Typography Rules

### Font Stack

| Anchor | Heading | Body | Mono |
|--------|---------|------|------|
| Light (Claude) | `Source Serif 4` (single weight 500) | `Inter Variable` | `JetBrains Mono` |
| Dark (Linear) | `Inter Variable` (weight 510, `cv01` + `ss03`) | `Inter Variable` (weight 510) | `JetBrains Mono` |

### Type Scale

| Token | Size | Weight (Light / Dark) | Line-Height | Letter-Spacing | Usage |
|-------|------|----------------------|-------------|----------------|-------|
| `text-display` | 2.5rem (40px) | 500 / 510 | 1.10 | -0.01em / -0.022em | Hero titles |
| `text-h1` | 1.75rem (28px) | 500 / 510 | 1.10 | -0.01em / -0.022em | Page headings |
| `text-h2` | 1.375rem (22px) | 500 / 510 | 1.20 | 0 / -0.022em | Section headings |
| `text-h3` | 1.125rem (18px) | 500 / 510 | 1.30 | 0 | Card titles |
| `text-body` | 0.9375rem (15px) | 400 / 510 | 1.6 | 0 | Body text |
| `text-small` | 0.8125rem (13px) | 400 / 510 | 1.5 | 0.01em | Captions, meta |
| `text-label` | 0.75rem (12px) | 500 | 1.4 | 0.02em | Buttons, badges, labels |

### Typography Rules
- Light anchor headings use **Source Serif 4 weight 500 only** (Anthropic single-weight rule).
- Dark anchor sets `font-variation-settings: "wght" 510` and `font-feature-settings: "cv01" on, "ss03" on, "cv11" on` on `[data-color-scheme="dark"]`.
- Chinese fallbacks: `Noto Serif SC` for headings, `Noto Sans SC` for body.
- Minimum readable size: 12px.

---

## 4. Component Stylings

### Button

| Variant | Background | Text | Hover | Active |
|---------|-----------|------|-------|--------|
| Primary | `var(--theme-primary)` | `var(--theme-primary-foreground)` | lighten 8% | scale(0.98) |
| Secondary | transparent | `var(--theme-foreground)` | `var(--theme-card)` bg | scale(0.98) |
| Ghost | transparent | `var(--theme-muted-foreground)` | `var(--theme-foreground)` text | — |
| Destructive | `var(--theme-status-error)` | `#fff` | lighten 8% | scale(0.98) |

- **Height**: 40px (standard), 48px (hero CTA), 32px (compact).
- **Radius**: 8px (standard), 9999px (pill/tag).

### Card / Panel

```css
background: var(--theme-card);
border: 1px solid var(--theme-border);
border-radius: 12px;
padding: 20px;
box-shadow: var(--theme-surface-shadow);
```

### Input / Textarea

```css
background: var(--theme-card);
border: 1px solid var(--theme-border);
border-radius: 10px;
padding: 12px 14px;
color: var(--theme-foreground);
```

- Light focus: `box-shadow: 0 0 0 1px var(--theme-border), 0 0 0 3px color-mix(in srgb, #3898ec 30%, transparent);`
- Dark focus: indigo ring via `--theme-primary-alpha`.
- Error state: `border-color: var(--theme-status-error)`.

### Drawer / Modal

- **Drawer**: slides from left/right, `backdrop-filter: blur(12px)`.
- **Modal**: centered, max-width 560px, ring shadow only (no drop shadows).
- **Animation**: 200ms ease-out transform + 150ms opacity.

---

## 5. Layout Principles

### Spacing Scale

| Token | Value |
|-------|-------|
| `space-1` | 4px |
| `space-2` | 8px |
| `space-3` | 12px |
| `space-4` | 16px |
| `space-5` | 20px |
| `space-6` | 24px |
| `space-8` | 32px |
| `space-10` | 40px |
| `space-12` | 48px |
| `space-16` | 64px |

### Grid
- Workspace max-width: 1200px, centered.
- Core input area max-width: 720px.
- Sidebar (history/settings): 360px width.
- Gap between major regions: `space-6` (24px).

### Z-Index Hierarchy

| Layer | Z-Index | Element |
|-------|---------|---------|
| Base | 0 | Content |
| Sticky | 10 | Sticky headers |
| Overlay | 40 | Drawers |
| Modal | 50 | Modals, dialogs |
| Toast | 60 | Toast notifications |
| Tooltip | 70 | Tooltips, popovers |

---

## 6. Depth & Elevation

Both anchors use **ring shadows** (no drop shadows), per the Anthropic and Linear references.

### Shadow Tokens

| Token | Light | Dark |
|-------|-------|------|
| `--theme-shadow-whisper` | `0 4px 24px rgba(0,0,0,0.05)` | `0 0 0 1px rgba(0,0,0,0.20)` |
| `--theme-shadow-ring` | `0 0 0 1px var(--theme-ring)` | `0 0 0 1px var(--theme-border)` |
| `--theme-surface-shadow` | `var(--theme-shadow-whisper)` | `var(--theme-shadow-whisper)` |

---

## 7. Do's and Don'ts

### Do
- Use `data-color-scheme="light"` or `"dark"` on `<html>`. That is the only switch.
- Use `var(--theme-*)` tokens instead of raw colors.
- Keep generated images as the visual focal point.
- Use OKLCH for new color definitions.
- Respect `prefers-reduced-motion`.

### Don't
- Don't reference `data-theme` — it is forbidden.
- Don't import `themes/base.css`, `themes/qi-baishi.css`, `themes/pop-anime.css`, `themes/rococo.css`, `themes/japanese-bw.css`, or `themes/workspace.css`. They have been deleted.
- Don't hardcode Tailwind color utilities like `bg-green-500`. Use `bg-primary`, `text-foreground`, etc.
- Don't add a third anchor without updating `design-principles.md` §0 first.
- Don't animate layout properties (width, height, top, left) — use transform.

---

## 8. Responsive Behavior

### Breakpoints

| Name | Width | Behavior |
|------|-------|----------|
| Mobile | < 640px | Single column, drawers become full-screen, stacked layout |
| Tablet | 640–1024px | Two-column where appropriate, drawers slide |
| Desktop | > 1024px | Full layout, side-by-side panels |

### Touch Targets
- Minimum: 44×44px for all interactive elements.

---

## 9. Agent Prompt Guide

### Quick Reference

```
Light primary:   #c96442 (terracotta)
Light bg:        #f5f4ed (parchment)
Light heading:   Source Serif 4, weight 500

Dark primary:    #5e6ad2 (indigo)
Dark bg:         #08090a (marketing black)
Dark heading:    Inter Variable, weight 510, cv01 + ss03
```

### Ready-to-Use Prompts

**Create a card:**
> "Create a card using `var(--theme-card)` background, 12px radius, `1px solid var(--theme-border)`, 20px padding, and `box-shadow: var(--theme-surface-shadow)`."

**Create a primary button:**
> "Create a primary button with `var(--theme-primary)` background, `var(--theme-primary-foreground)` text, 8px radius, 40px height. Active state scales to 0.98."

**Switch color scheme:**
> "Set `document.documentElement.dataset.colorScheme = 'dark'` to switch to the Linear dark anchor. Do not write `data-theme` — the legacy attribute is forbidden."

---

*Last updated: 2026-04-27 (post `themes-light-dark-only` task)*
