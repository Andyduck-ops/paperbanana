# Learnings

Cross-task knowledge accumulated during Golden Data batch implementation.

## Phase Learnings

### Test Organization

- Group tests by Golden Data case prefix (e.g., GD-001-M-*) for easy mapping to YAML specifications
- Use nested describe blocks to organize related test cases within a file
- Keep test file names aligned with the case IDs they test

### Test Patterns

- Use `getAllByText` when checking for status icons that may appear multiple times
- Use `{ container }` pattern for class-based assertions
- Mock hooks (useLanguage) consistently across all test files

### Golden Data Categories

1. **Happy Path (GD-001)**: Standard successful execution
2. **Failure Handling (GD-002)**: Stage failure attribution
3. **Resume Correctness (GD-003)**: Snapshot restoration
4. **UI States (GD-UI-001 to GD-UI-005)**: UI component behavior

### Common Anti-Patterns to Test

- Generic error messages without context
- Hidden failures (silent partial completion)
- Progress indicators that lie (stuck at 99%)
- Missing stage attribution in errors
- Resume that doesn't show previous progress

## Verification Learnings

### Test Selector Patterns

- Use `getAllByText` when testing for status text that appears in multiple elements (badge + error message)
- Test for icons/symbols using regex: `/Failed/i` matches both "Failed" and "failed"
- Container-based assertions (`container.querySelector`) work well for class-based checks

### Test-Component Interface Validation

- Before writing tests, verify the actual component interface:
  1. Check required vs optional props
  2. Identify which hooks are used internally
  3. Check for composed components (like `DualInputPanel`)
- Tests should match the actual component API, not an assumed one

### Verification Report Quality

- Include pass rate percentage for quick assessment
- Categorize failures by root cause for actionable insights
- Map test failures to contract requirements for compliance tracking

## Technical Notes

- Vitest with @testing-library/react works well for component testing
- Tests can run without backend (mock-based approach)
- Stage state types should match component interfaces exactly
