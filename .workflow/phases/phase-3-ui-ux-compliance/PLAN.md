# Phase 3: UI/UX Compliance - Plan

**Created:** 2026-03-24
**Status:** Ready for execution

## Goals

1. Implement CSS tokens for all status colors
2. Add 100ms click feedback and hover states
3. Implement accessibility (a11y) features
4. Reduce cognitive load with simplified error messages

## Tasks

### Task 3.1: Visual Design Consistency

**Priority:** P1
**Files:**
- `web/src/themes/base.css` - Add status color tokens
- `web/src/themes/pop-art.css` - Add theme-specific status colors

**Implementation:**
1. Add CSS tokens for status colors:
   - `--status-success`, `--status-warning`, `--status-error`, `--status-info`
   - `--status-pending`, `--status-running`, `--status-completed`

2. Standardize spacing tokens:
   - Already defined in base.css
   - Add `--radius-*` tokens for border radius

3. Add dark mode support:
   - Create `pop-art-dark.css` with dark theme variants

**Acceptance Criteria:**
- All status colors use CSS tokens
- Components use standardized spacing
- Dark mode theme available

### Task 3.2: Interaction Feedback

**Priority:** P1
**Files:**
- `web/src/themes/base.css` - Add interaction tokens
- `web/src/components/` - Update interactive components

**Implementation:**
1. Add 100ms click feedback:
   - `--transition-fast: 100ms`
   - Add `:active` states with scale transform

2. Implement hover states:
   - Add `:hover` states for all buttons, links, cards
   - Use `--transition-medium: 200ms`

3. Add keyboard navigation support:
   - Add `:focus-visible` styles
   - Ensure all interactive elements are focusable

**Acceptance Criteria:**
- Click feedback feels responsive (< 150ms)
- All interactive elements have hover states
- Keyboard navigation works

### Task 3.3: Accessibility (a11y)

**Priority:** P1
**Files:**
- `web/src/components/ProgressPanel.tsx` - Add aria labels
- `web/src/components/StageCard.tsx` - Add live regions
- `web/src/themes/base.css` - Add reduced motion support

**Implementation:**
1. Screen reader announcements:
   - Add `aria-live="polite"` for status updates
   - Add `aria-label` for icon buttons

2. WCAG AA contrast compliance:
   - Verify all text/background combinations
   - Adjust colors as needed

3. `prefers-reduced-motion` support:
   - Add media query to disable animations
   - Reduce or remove transitions for users who prefer reduced motion

**Acceptance Criteria:**
- Screen reader announces progress changes
- All text meets WCAG AA contrast ratio (4.5:1)
- Animations respect user preferences

### Task 3.4: Cognitive Load Optimization

**Priority:** P1
**Files:**
- `web/src/components/ResultPanel.tsx` - Simplify error display
- `web/src/components/ProgressPanel.tsx` - Add elapsed time

**Implementation:**
1. Simplify error messages:
   - Show action items clearly
   - Hide stack traces behind "details" toggle

2. Add elapsed time display:
   - Show time since start
   - Update every second during execution

3. Implement undo for destructive actions:
   - Add confirmation for delete/cancel
   - Consider toast notifications for undo

**Acceptance Criteria:**
- Error messages show clear action items
- Elapsed time visible during processing
- Destructive actions require confirmation

## Verification

- All components use CSS tokens
- Keyboard navigation works for all interactive elements
- Screen reader announces progress
- WCAG AA contrast compliance verified
- Reduced motion preferences respected

## Files Changed

| File | Change |
|------|--------|
| `web/src/themes/base.css` | Add status colors, interaction tokens, reduced motion |
| `web/src/themes/pop-art.css` | Add status color values |
| `web/src/themes/pop-art-dark.css` | New - Dark theme variant |
| `web/src/components/ProgressPanel.tsx` | Add aria, elapsed time |
| `web/src/components/StageCard.tsx` | Add live regions |
| `web/src/components/ResultPanel.tsx` | Simplify error display |
