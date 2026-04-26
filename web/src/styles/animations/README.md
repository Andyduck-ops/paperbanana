# PaperBanana Motion Design System

A comprehensive animation system for PaperBanana built with performance and accessibility in mind.

## Features

- ✅ **GPU-accelerated animations** - Only animate compositor-safe properties
- ✅ **Reduced motion support** - Full accessibility compliance
- ✅ **TypeScript** - Full type definitions
- ✅ **React hooks** - Powerful composition primitives
- ✅ **Pre-built components** - Ready-to-use animated components
- ✅ **Motion tokens** - Consistent easing and duration scales

## Quick Start

### 1. Import CSS

```tsx
// In your main.tsx or App.tsx
import '@/styles/animations/keyframes.css';
import '@/styles/animations/components.css';
```

### 2. Use Components

```tsx
import { FadeIn, ScaleIn, StaggerContainer, Skeleton } from '@/components/motion';

// Fade animation on scroll
<FadeIn>
  <h1>Hello World</h1>
</FadeIn>

// Scale with spring
<ScaleIn spring>
  <NotificationBadge />
</ScaleIn>

// Staggered children
<StaggerContainer>
  {items.map(item => <Card key={item.id} {...item} />)}
</StaggerContainer>

// Loading skeleton
<Skeleton variant="text" lines={3} />
```

### 3. Use Hooks

```tsx
import { useScrollReveal, useReducedMotion, useStagger } from '@/components/motion';

function MyComponent() {
  const [ref, isVisible, style] = useScrollReveal({
    direction: 'up',
    duration: 'slow',
    threshold: 0.3,
  });

  const { prefersReducedMotion } = useReducedMotion();
  const delay = useStagger({ index: 2, increment: 0.1 });

  return (
    <div ref={ref} style={{ ...style, transitionDelay: `${delay}s` }}>
      Content
    </div>
  );
}
```

## Animation Components

### FadeIn

Fade animation with optional direction.

```tsx
<FadeIn direction="up" duration="slow" triggerOnView>
  <Content />
</FadeIn>
```

**Props:**
- `direction`: `'up' | 'down' | 'left' | 'right' | 'none'`
- `duration`: `'fast' | 'base' | 'slow' | 'slower' | 'slowest' | number`
- `easing`: `'ease-out' | 'ease-in-out' | 'ease-spring' | 'ease-in'`
- `delay`: `number` (seconds)
- `triggerOnView`: `boolean`
- `threshold`: `number` (0-1)

### SlideIn

Slide animation from a direction.

```tsx
<SlideIn direction="left" duration="slower">
  <Sidebar />
</SlideIn>
```

### ScaleIn

Scale animation with optional spring physics.

```tsx
<ScaleIn spring duration="slow">
  <Modal />
</ScaleIn>
```

### StaggerContainer

Staggers animations for child elements.

```tsx
<StaggerContainer increment={0.08} maxItems={6}>
  {cards.map(card => <Card key={card.id} {...card} />)}
</StaggerContainer>
```

### Skeleton

Loading placeholder with shimmer effect.

```tsx
<Skeleton variant="text" lines={3} animated />
<Skeleton variant="circular" width={40} height={40} />
<Skeleton variant="rectangular" width="100%" height={200} />
```

### PageTransition

Page transition wrapper.

```tsx
<PageTransition pageKey={location.pathname} direction="up" duration={0.5}>
  <Outlet />
</PageTransition>
```

### Modal

Animated modal dialog.

```tsx
<Modal
  isOpen={isOpen}
  onClose={handleClose}
  title="Confirm Action"
  size="md"
  closeOnBackdrop
  closeOnEscape
>
  <p>Are you sure?</p>
</Modal>
```

### Toast

Toast notification with auto-dismiss.

```tsx
<Toast
  isOpen={showToast}
  onClose={() => setShowToast(false)}
  message="Operation successful!"
  type="success"
  duration={5000}
  position="top-right"
/>
```

### ProgressBar

Progress indicator with variants.

```tsx
<ProgressBar value={75} size="md" color="primary" striped />
<ProgressBar indeterminate size="lg" color="success" />
```

### Spinner

Loading spinner.

```tsx
<Spinner size="md" color="primary" label="Loading data..." />
```

## Motion Tokens

### Duration Scale

| Token | Value | Use Case |
|-------|-------|----------|
| `--duration-fast` | 0.15s | Micro-interactions |
| `--duration-base` | 0.3s | Standard transitions |
| `--duration-slow` | 0.6s | Content reveals |
| `--duration-slower` | 0.8s | Page transitions |
| `--duration-slowest` | 1.2s | Hero animations |

### Easing Functions

| Token | Value | Use Case |
|-------|-------|----------|
| `--ease-out` | cubic-bezier(0.16, 1, 0.3, 1) | Content entry |
| `--ease-in-out` | cubic-bezier(0.65, 0, 0.35, 1) | State changes |
| `--ease-spring` | cubic-bezier(0.34, 1.56, 0.64, 1) | Playful interactions |
| `--ease-in` | cubic-bezier(0.4, 0, 1, 1) | Progress indicators |

### CSS Usage

```css
.card {
  transition: transform var(--duration-base) var(--ease-out);
}

.card:hover {
  transform: translateY(-4px);
}
```

## Hooks API

### useReducedMotion

```tsx
const { prefersReducedMotion, getAnimationDuration, shouldAnimate } = useReducedMotion();
```

### useScrollReveal

```tsx
const [ref, isVisible, style] = useScrollReveal({
  direction: 'up',
  distance: 20,
  duration: 'slow',
  easing: 'ease-out',
  threshold: 0.2,
});

return <div ref={ref} style={style}>Content</div>;
```

### useIntersectionObserver

```tsx
const [ref, isIntersecting, entry] = useIntersectionObserver({
  threshold: 0.1,
  rootMargin: '0px',
  triggerOnce: true,
});
```

### useStagger

```tsx
const delay = useStagger({ index: 3, increment: 0.05, maxItems: 8 });
const delays = useStaggerArray({ count: 5, baseDelay: 0.1 });
```

### useAnimatedValue

```tsx
const [value, setTarget, isAnimating] = useAnimatedValue({
  target: 100,
  duration: 0.6,
  easing: 'ease-out',
  onComplete: () => console.log('Done!'),
});
```

## Reduced Motion Support

All animations respect `prefers-reduced-motion`:

- Animations are reduced to instant transitions (0.01ms)
- Parallax effects are disabled
- Infinite loops are stopped
- Opacity-only fades are allowed

```tsx
// Automatic in all components
<FadeIn> {/* Respects user preference */} </FadeIn>

// Manual check in custom animations
const { prefersReducedMotion } = useReducedMotion();
```

## Performance Guidelines

See [PERFORMANCE.md](./PERFORMANCE.md) for detailed guidelines.

### Quick Tips

1. **Only animate `transform` and `opacity`**
2. **Use `will-change` sparingly** (max 3-4 elements)
3. **Remove `will-change` after animation**
4. **Use IntersectionObserver** for scroll triggers
5. **Batch animations** with stagger
6. **Test at 60fps** with Chrome DevTools

## Browser Support

- Chrome 74+
- Firefox 63+
- Safari 10.1+
- Edge 79+
- iOS Safari 10.3+
- Android Chrome 74+

## License

Part of PaperBanana - AI Academic Paper Figure Generation Tool
