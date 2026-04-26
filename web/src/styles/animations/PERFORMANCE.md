# PaperBanana Animation Performance Guidelines

## Quick Reference

```
ANIMATE:  transform, opacity, filter
AVOID:    width, height, top, left, margin, padding, color, background-color
BUDGET:   will-change on max 3-4 elements, remove after use
TARGET:   60fps = 16.67ms per frame
```

## GPU-Accelerated Properties

### ✅ Safe Properties (Compositor Thread)

| Property | Examples | Use Case |
|----------|----------|----------|
| `transform` | `translate()`, `scale()`, `rotate()` | Position, size, rotation changes |
| `opacity` | `0` to `1` | Fade in/out, overlays |
| `filter` | `blur()`, `brightness()` | Visual effects |
| `backdrop-filter` | `blur()`, `brightness()` | Glass morphism |

### ❌ Unsafe Properties (Trigger Layout/Paint)

| Property | Impact | Alternative |
|----------|--------|-------------|
| `width`/`height` | Layout | `transform: scale()` |
| `top`/`left` | Layout | `transform: translate()` |
| `margin`/`padding` | Layout | Inner element with transform |
| `color` | Paint | Overlay with opacity |
| `background-color` | Paint | Pseudo-element with opacity |

## will-change Budget

```css
/* ❌ Wrong: Permanent will-change */
.card {
  will-change: transform;
}

/* ✅ Correct: Add before, remove after */
.card.will-animate {
  will-change: transform;
}
```

**Rules:**
- Max 3-4 elements with `will-change` simultaneously
- Remove `will-change` after animation completes
- Never use `will-change: auto` on collections
- Use explicit properties: `will-change: transform, opacity`

## Frame Budget

| Metric | Target | Budget |
|--------|--------|--------|
| Frame rate | 60fps | 16.67ms per frame |
| Style + Layout | < 5ms | ~30% of frame budget |
| Paint + Composite | < 5ms | ~30% of frame budget |
| JavaScript | < 5ms | ~30% of frame budget |
| Idle buffer | ~1.67ms | Headroom for GC |

## Height Animation Trick

Use CSS Grid for smooth height animations without triggering layout:

```css
.expandable {
  display: grid;
  grid-template-rows: 0fr;
  transition: grid-template-rows var(--duration-slow) var(--ease-out-quart);
}

.expandable.open {
  grid-template-rows: 1fr;
}

.expandable > .content {
  overflow: hidden;
}
```

## Perceived Performance

| Technique | Effect |
|-----------|--------|
| **Preemptive starts** | Begin on `pointerdown` not `click` (saves ~80-120ms) |
| **Early completion** | Visual feedback can finish before operation completes |
| **Ease-in for progress** | Compresses perceived wait time |
| **Ease-out for entrances** | Natural deceleration feels "settled" |

## Intersection Observer Pattern

```typescript
// ✅ Correct: Use IntersectionObserver for scroll triggers
const observer = new IntersectionObserver((entries) => {
  entries.forEach(entry => {
    if (entry.isIntersecting) {
      entry.target.classList.add('is-visible');
      observer.unobserve(entry.target); // One-shot
    }
  });
}, { threshold: 0.2, rootMargin: '0px 0px -100px 0px' });

// ❌ Wrong: Scroll event listeners
window.addEventListener('scroll', () => {
  // Triggers on every scroll frame - expensive!
});
```

## requestAnimationFrame Throttling

```typescript
// ✅ Correct: RAF with ticking guard
let ticking = false;
window.addEventListener('scroll', () => {
  if (!ticking) {
    requestAnimationFrame(() => {
      // Do animation work
      ticking = false;
    });
    ticking = true;
  }
});
```

## Stagger Budget

- Max 8 items in a stagger sequence
- Total stagger time: max 0.8s
- Formula: `delay = base_delay + (index * stagger_increment)`
- For >8 items: batch into groups of 4-6

## Testing Performance

### Chrome DevTools Performance Panel

1. Open DevTools → Performance tab
2. Click Record (⏺️)
3. Trigger animation
4. Stop recording

**Look for:**
- 🟣 Purple "Layout" bars during animation = problem
- 🟢 Green "Paint" bars during animation = problem
- 🔴 Red frame markers = dropped frames (>16.67ms)
- ⚠️ "Forced reflow" warnings = layout thrashing

### Frame Rate Targets

| Scenario | Minimum FPS | Average FPS |
|----------|-------------|-------------|
| Page transitions | 45 | 60 |
| Scroll animations | 50 | 60 |
| Micro-interactions | 55 | 60 |
| Complex choreography | 40 | 55 |

## Reduced Motion

Always respect user preferences:

```css
@media (prefers-reduced-motion: reduce) {
  *,
  *::before,
  *::after {
    animation-duration: 0.01ms !important;
    animation-iteration-count: 1 !important;
    transition-duration: 0.01ms !important;
    scroll-behavior: auto !important;
  }
}
```

```typescript
// JavaScript detection
const prefersReducedMotion = window.matchMedia('(prefers-reduced-motion: reduce)');
if (prefersReducedMotion.matches) {
  // Disable parallax, springs, infinite loops
}
```

## Common Pitfalls

| Pitfall | Solution |
|---------|----------|
| Animating layout properties | Use `transform` instead |
| Too many simultaneous animations | Batch and stagger |
| Not removing `will-change` | Add via JS, remove on transitionend |
| Scroll-linked animations without RAF | Use `requestAnimationFrame` with ticking guard |
| Infinite loops without reduced-motion check | Detect and disable for reduced-motion |
| Heavy filters on many elements | Limit to 1-2 hero elements |

## File Organization

```
web/src/styles/animations/
├── motion-tokens.json      # Token definitions
├── keyframes.css           # @keyframes library
├── components.css          # Component animation styles
└── PERFORMANCE.md          # This file

web/src/components/motion/
├── index.ts                # Public exports
├── types.ts                # TypeScript definitions
├── AnimatedContainer.tsx   # Generic animated wrapper
├── FadeIn.tsx              # Fade animation
├── SlideIn.tsx             # Slide animation
├── ScaleIn.tsx             # Scale animation
├── StaggerContainer.tsx    # Stagger children
├── Skeleton.tsx            # Loading placeholder
├── PageTransition.tsx      # Page transition wrapper
├── Modal.tsx               # Modal dialog
├── Toast.tsx               # Toast notification
├── ProgressBar.tsx         # Progress indicator
├── Spinner.tsx             # Loading spinner
└── hooks/                  # Animation hooks
    ├── index.ts
    ├── useReducedMotion.ts
    ├── useIntersectionObserver.ts
    ├── useScrollReveal.ts
    ├── useStagger.ts
    └── useAnimatedValue.ts
```

## Animation Presets

| Animation | Duration | Easing | Use Case |
|-----------|----------|--------|----------|
| `fade-up` | 0.6s | ease-out | Content reveals |
| `scale-spring` | 0.6s | ease-spring | Notifications, modals |
| `slide-in-right` | 0.8s | ease-out | Page transitions |
| `modal-enter` | 0.3s | ease-spring | Dialog open |
| `toast-enter` | 0.6s | ease-out | Notifications |
