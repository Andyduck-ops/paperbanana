# Interface Fix Report: GD-001-M-empty-prompt

## Task: INTERFACE-001

**Status**: Completed
**Test File**: `web/src/test/golden/GD-001-M-empty-prompt.test.tsx`
**Component**: `web/src/components/GeneratePanel.tsx`

## Interface Mismatch Analysis

### Original Test Assumptions (INCORRECT)

1. **GeneratePanel has internal `useGenerate` hook** - NO, it receives `onGenerate` prop
2. **Direct `<textarea>` with `getByRole('textbox')`** - NO, it uses `DualInputPanel` with two textareas (method and caption)
3. **Button with name matching `/generate/i`** - PARTIALLY CORRECT, but text is from i18n (`t('generate.submit')` = "Generate")
4. **No props required for GeneratePanel** - NO, `onGenerate` is a required prop

### Actual Component Interface

```typescript
interface GeneratePanelProps {
  onGenerate: (prompt: string, options?: GenerateOptions) => void;  // REQUIRED
  isGenerating?: boolean;
  visualizerNodes?: string[];
  onNavigateToSettings?: () => void;
}
```

Key behaviors:
- Uses `DualInputPanel` with `methodContent` and `caption` textareas
- Button text: `t('generate.submit')` = "Generate" or `t('generate.generating')` = "Generating..."
- Validation: `hasContent = methodContent.trim() || caption.trim()` - button is disabled when empty
- When submitted with empty content, `onGenerate` is NOT called (due to `if (combinedPrompt)` check)

## Changes Made

### Mock Updates

1. **Removed `useGenerate` mock** - Component doesn't use this hook internally
2. **Added `useLanguage` mock** - Component uses i18n for button text
3. **Added `useProviders` mock** - Component fetches provider list

### Test Updates

1. **Pass `onGenerate` prop** - Required for component to function
2. **Use `getAllByRole('textbox')`** - Two textareas exist (method and caption)
3. **Check button disabled state** - Primary validation mechanism is disabled button
4. **Added `Valid Content Submission` tests** - Ensure positive path also works

### Files Modified

- `web/src/test/golden/GD-001-M-empty-prompt.test.tsx`

## Test Results

```
Test Files  1 passed (1)
Tests       10 passed (10)
Duration    10.21s
```

All tests now pass:
- Empty Prompt Detection (2 tests)
- Error Message Quality (2 tests)
- Pipeline Not Started (2 tests)
- No Artifact Created (1 test)
- Validation Before Pipeline (1 test)
- Valid Content Submission (2 tests)

## Lessons Learned

1. Always read the actual component implementation before writing tests
2. Component props are the contract - tests should verify against the actual interface
3. UI validation can manifest as disabled buttons rather than error messages
4. i18n keys affect button text matching in tests
