# Selector Fix Report

## Summary

Fixed 4 failing tests across 2 test files by updating multi-element selectors to use more specific patterns.

## Tasks Completed

### SELECTOR-001: GD-UI-005.test.tsx

**Issue**: Tests used `getByText(/Failed/i)` and `getByText('checkmark')` which matched multiple elements or no elements.

**Fixes Applied**:
1. Line 109: Changed `screen.getByText(/Failed/i)` to `screen.getByText(/^✗ Failed$/)`
2. Line 159: Changed `screen.getByText(/Failed/i)` to `screen.getByText(/^✗ Failed$/)`
3. Line 296: Changed `screen.getByText('✓')` to `screen.getByText(/^✓ Successful$/)`
4. Line 320: Changed `screen.getByText('✗')` to `screen.getByText(/^✗ Failed$/)`

**Result**: All 15 tests pass.

### SELECTOR-002: GD-BATCH.test.tsx

**Issue**: Test used `getByText(/Failed/i)` which matched both the status text and the error message.

**Fix Applied**:
1. Line 119: Changed `screen.getByText(/Failed/i)` to `screen.getByText(/^✗ Failed$/)`

**Result**: All 20 tests pass.

## Root Cause Analysis

The component `BatchProgressPanel` displays status text that includes Unicode symbols:
- Completed: `✓ Successful`
- Failed: `✗ Failed`
- Running: `◆`

The error message paragraph also contains "Failed" (e.g., "Failed at planner stage", "Generation failed"), causing regex `/Failed/i` to match multiple elements.

## Solution Pattern

Use exact match patterns with anchors (`^` and `$`) to target specific text:
```typescript
// Before
screen.getByText(/Failed/i)

// After
screen.getByText(/^✗ Failed$/)
```

This ensures we match only the status badge text, not the error message below it.

## Files Modified

1. `web/src/test/golden/GD-UI-005.test.tsx` - 4 selector fixes
2. `web/src/test/golden/GD-BATCH.test.tsx` - 1 selector fix

## Verification

```bash
# GD-UI-005.test.tsx: 15 tests passed
# GD-BATCH.test.tsx: 20 tests passed
# Total: 35 tests passed
```
