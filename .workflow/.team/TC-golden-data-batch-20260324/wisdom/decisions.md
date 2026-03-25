# Decisions

Key decisions made during Golden Data batch implementation.

## Architecture Decisions

### Test File Organization
- **Decision**: Create one test file per category of Golden Data cases
- **Rationale**: Balances file size with logical grouping, makes maintenance easier
- **Alternatives Considered**: One test per case (too many files), all tests in one file (too large)

### Mock Strategy
- **Decision**: Use vi.mock for hooks, render real components
- **Rationale**: Tests component behavior without backend dependencies
- **Trade-off**: Doesn't test hook integration, but that's covered by integration tests

### Assertion Patterns
- **Decision**: Test both positive requirements and anti-patterns
- **Rationale**: Golden Data cases specify both what must happen and what must not happen
- **Implementation**: Use `getByText` for required elements, `queryByText` for forbidden elements

## Test Coverage Decisions

### Phase 2 (Core Pipeline Hardening)
- Focus on error handling, timeouts, and boundary conditions
- Cover all retrieval modes (auto, manual, random, none)
- Test timeout messages are actionable

### Phase 3 (UI/UX Compliance)
- Visual consistency across all stage states
- Cognitive load optimization (progress visibility, no overwhelming detail)
- Responsive design for different viewports

### Phase 4-6
- Resume correctness with metadata preservation
- Batch processing with per-task status visibility
- Intent handling for different modes and constraints
