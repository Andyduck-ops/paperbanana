# PaperBanana Design System

> Inspired by RunwayML's cinematic dark UI and creative-tool precision. Adapted for PaperBanana's multi-art-theme scientific visualization workflow.

---

## 1. Visual Theme & Atmosphere

### Mood
Cinematic, focused, gallery-like. The UI exists to showcase generated visuals — it should feel like a darkroom or editing suite, not a dashboard.

### Density
Medium-low density. Generous whitespace around the canvas. Controls are grouped and collapsible. The generated image is always the hero.

### Design Philosophy
- **Content-first**: Generated artifacts occupy the visual center; chrome recedes.
- **Theme-aware**: The base UI is neutral and adaptive. Art themes (academic, Qi Baishi, pop-anime, Rococo, Night Mono) inject personality through color, typography, and texture — but never compromise readability.
- **Dark-first**: Default dark mode is optimized for long creative sessions. Light mode is supported but secondary.
- **Motion with purpose**: Transitions guide attention; no decorative animation.

---

## 2. Color Palette & Roles

### Base Tokens (Dark Mode — Default)

| Token | Hex / OKLCH | Role |
|-------|-------------|------|
| `--color-background` | `#0a0a0f` | App canvas, deepest layer |
| `--color-surface` | `#13131a` | Cards, panels, drawers |
| `--color-surface-elevated` | `#1a1a24` | Modals, popovers, dropdowns |
| `--color-border` | `#2a2a3a` | Dividers, outlines |
| `--color-border-subtle` | `#1e1e2a` | Hairline separators |
| `--color-foreground` | `#f0f0f5` | Primary text |
| `--color-foreground-muted` | `#8a8a9a` | Secondary text, placeholders |
| `--color-foreground-dim` | `#5a5a6a` | Tertiary text, disabled |
| `--color-primary` | `#e8d5b5` | CTA buttons, active states, accent (warm champagne) |
| `--color-primary-hover` | `#f0e0c8` | Primary hover |
| `--color-primary-active` | `#dcc0a0` | Primary pressed |
| `--color-accent` | `#c4a45a` | Highlights, badges, progress indicators |
| `--color-accent-glow` | `rgba(196, 164, 90, 0.25)` | Focus rings, selection glow |
| `--color-status-success` | `#6abe6a` | Success states |
| `--color-status-error` | `#e05a5a` | Errors, failures |
| `--color-status-warning` | `#e0b85a` | Warnings |
| `--color-status-info` | `#6a9ae0` | Info, running states |

### Light Mode

| Token | Hex | Role |
|-------|-----|------|
| `--color-background` | `#f6f2ea` | Warm paper-like canvas |
| `--color-surface` | `#fcfaf5` | Cards, panels |
| `--color-surface-elevated` | `#ffffff` | Modals, popovers |
| `--color-border` | `#e0ddd5` | Dividers |
| `--color-foreground` | `#1a1a20` | Primary text |
| `--color-foreground-muted` | `#6a6a75` | Secondary text |
| `--color-primary` | `#2a4a7a` | Primary actions (deep academic blue) |
| `--color-accent` | `#c44a3a` | Accent (ink red) |

### Art Theme Overrides

Art themes override a **subset** of base tokens to inject personality. They never redefine functional colors (status, error, success).

| Theme | Override Character |
|-------|-------------------|
| `academic` | Deep blues, scholarly neutrals, serif headings |
| `qi-baishi` | Warm rice-paper white, ink red accents, traditional serif |
| `pop-anime` | Vibrant cream, bold red/blue accents, playful sans |
| `rococo` | Soft blush white, gold accents, ornamental serif |
| `japanese-bw` | True black, high-contrast white, monospace |

### Usage Rules
- **Never** hardcode hex values in components. Always use CSS custom properties.
- Art themes set `data-theme`; luminance mode sets `data-color-scheme`.
- Both attributes coexist: `<html data-theme="qi-baishi" data-color-scheme="dark">`.

---

## 3. Typography Rules

### Font Stack

| Role | Font | Fallback | Usage |
|------|------|----------|-------|
| Heading | `Fraunces` | `Noto Serif SC`, serif | Page titles, section headers, brand |
| Body | `IBM Plex Sans` | `Noto Sans SC`, sans-serif | UI text, labels, body copy |
| Mono | `IBM Plex Mono` | `monospace` | Code, timestamps, metadata |

### Type Scale

| Token | Size | Weight | Line-Height | Letter-Spacing | Usage |
|-------|------|--------|-------------|----------------|-------|
| `text-display` | 2.5rem (40px) | 600 | 1.1 | -0.02em | Hero titles |
| `text-h1` | 1.75rem (28px) | 600 | 1.2 | -0.01em | Page headings |
| `text-h2` | 1.375rem (22px) | 600 | 1.3 | 0 | Section headings |
| `text-h3` | 1.125rem (18px) | 500 | 1.4 | 0 | Card titles |
| `text-body` | 0.9375rem (15px) | 400 | 1.6 | 0 | Body text |
| `text-small` | 0.8125rem (13px) | 400 | 1.5 | 0.01em | Captions, meta |
| `text-label` | 0.75rem (12px) | 500 | 1.4 | 0.02em | Buttons, badges, labels |
| `text-mono` | 0.8125rem (13px) | 400 | 1.5 | 0 | Timestamps, IDs |

### Typography Rules
- Headings use `font-heading`; body uses `font-body`.
- Chinese text always falls back to `Noto Serif SC` / `Noto Sans SC`.
- Uppercase + wide tracking only for `text-label` in English.
- Minimum readable size: 12px.

---

## 4. Component Stylings

### Button

| Variant | Background | Text | Border | Hover | Active |
|---------|-----------|------|--------|-------|--------|
| Primary | `var(--color-primary)` | `var(--color-background)` | none | lighten 8%, lift shadow | darken 5%, scale(0.98) |
| Secondary | transparent | `var(--color-foreground)` | `1px solid var(--color-border)` | `var(--color-surface)` bg | scale(0.98) |
| Ghost | transparent | `var(--color-foreground-muted)` | none | `var(--color-foreground)` text | — |
| Destructive | `var(--color-status-error)` | `#fff` | none | lighten 8% | darken 5% |

- **Height**: 40px (standard), 48px (hero CTA), 32px (compact).
- **Radius**: 8px (standard), 9999px (pill/tag).
- **Icon + Text**: 8px gap, icon 16×16.

### Card / Panel

```
background: var(--color-surface);
border: 1px solid var(--color-border);
border-radius: 12px;
padding: 20px;
```

- Hover: `border-color` shifts to `var(--color-border-hover)` (if defined).
- No box-shadow by default in dark mode; subtle shadow in light mode.

### Input / Textarea

```
background: var(--color-surface-elevated);
border: 1px solid var(--color-border);
border-radius: 10px;
padding: 12px 14px;
color: var(--color-foreground);
```

- Focus: `outline: 2px solid var(--color-accent-glow); border-color: var(--color-accent);`
- Placeholder: `var(--color-foreground-dim)`.
- Error state: `border-color: var(--color-status-error)`.

### Drawer / Modal

- **Drawer**: slides from left/right, `backdrop-filter: blur(12px)`, bg `rgba(10,10,15,0.85)`.
- **Modal**: centered, `box-shadow: 0 24px 80px rgba(0,0,0,0.5)`, max-width 560px.
- **Animation**: 200ms ease-out transform + 150ms opacity.

### Progress / Status

- Progress line: 2px height, `var(--color-accent)` fill, `var(--color-border)` track.
- Stage indicators: 8px dots, colored by status token.

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

### Shadow System

| Token | Dark Mode | Light Mode |
|-------|-----------|------------|
| `shadow-sm` | none | `0 1px 2px rgba(0,0,0,0.05)` |
| `shadow-md` | `0 4px 16px rgba(0,0,0,0.3)` | `0 4px 16px rgba(0,0,0,0.08)` |
| `shadow-lg` | `0 12px 40px rgba(0,0,0,0.4)` | `0 12px 40px rgba(0,0,0,0.1)` |
| `shadow-xl` | `0 24px 80px rgba(0,0,0,0.5)` | `0 24px 80px rgba(0,0,0,0.12)` |

### Surface Hierarchy

1. **Background**: deepest, no shadow.
2. **Surface**: cards, panels — `shadow-sm` in light.
3. **Surface Elevated**: modals, dropdowns — `shadow-md` or higher.
4. **Overlay**: backdrops with blur + semi-transparent black.

---

## 7. Do's and Don'ts

### Do
- Use `data-theme` + `data-color-scheme` together.
- Use semantic tokens (`--color-primary`, `--color-surface`) instead of raw colors.
- Keep generated images as the visual focal point.
- Use OKLCH for color definitions where possible.
- Respect `prefers-reduced-motion`.
- Use `font-heading` for titles, `font-body` for everything else.

### Don't
- Don't use pure black (`#000`) or pure white (`#fff`) as backgrounds.
- Don't hardcode Tailwind color utilities like `bg-green-500`.
- Don't place decorative elements above generated content.
- Don't use more than 2 font families on the same screen.
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
- Buttons in mobile: full-width where appropriate.

### Collapsing Strategy
- History drawer: full-screen overlay on mobile, 360px panel on desktop.
- Settings drawer: same as history.
- Batch candidate grid: 1 column mobile, 2 tablet, 3+ desktop.

---

## 9. Agent Prompt Guide

### Quick Reference

```
Primary:   #e8d5b5 (warm champagne)
Accent:    #c4a45a (gold)
Surface:   #13131a (dark panel)
Background:#0a0a0f (deep void)
Text:      #f0f0f5 (soft white)
Muted:     #8a8a9a (gray)
```

### Ready-to-Use Prompts

**Create a card component:**
> "Create a PaperBanana card using `var(--color-surface)` background, 12px radius, 1px `var(--color-border)` border, 20px padding. Hover should brighten the border."

**Create a primary button:**
> "Create a PaperBanana primary button with `var(--color-primary)` background, `var(--color-background)` text, 8px radius, 40px height. Active state scales to 0.98."

**Apply a theme:**
> "Apply the `qi-baishi` art theme by setting `data-theme='qi-baishi'` and using warm rice-paper tones with ink-red accents."

**Dark mode section:**
> "Build a dark-mode-first section with `var(--color-background)` canvas, `var(--color-surface)` cards, and `var(--color-foreground)` text. Light mode should invert gracefully via `data-color-scheme`."

---

*Last updated: 2026-04-18*
*Format: Stitch DESIGN.md v1*
