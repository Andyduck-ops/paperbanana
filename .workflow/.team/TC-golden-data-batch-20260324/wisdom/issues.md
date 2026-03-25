# Issues

Blocking issues encountered during Golden Data batch implementation.

## Open Issues

### Issue 1: Test-Component Interface Mismatch (GD-001-M-empty-prompt)

**Severity**: Medium (not blocking - tests need update)
**Files Affected**: `web/src/test/golden/GD-001-M-empty-prompt.test.tsx`
**Description**: Test file expects `GeneratePanel` with different interface:
- Expects standalone component with `useGenerate` hook
- Actual component requires `onGenerate` prop
- Uses `DualInputPanel` for input, not direct textarea
**Remediation**: Rewrite tests to match actual component API or create integration tests

### Issue 2: Multiple Element Selector Pattern (GD-UI-005, GD-BATCH)

**Severity**: Low (simple fix)
**Files Affected**:
- `web/src/test/golden/GD-UI-005.test.tsx`
- `web/src/test/golden/GD-BATCH.test.tsx`
**Description**: Tests use `getByText` for elements that appear multiple times
**Remediation**: Use `getAllByText` pattern or more specific selectors
