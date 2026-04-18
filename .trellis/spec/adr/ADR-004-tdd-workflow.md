---
id: ADR-004
status: Accepted
traces_to: [REQ-004]
date: 2026-04-08
---

# ADR-004: TDD Workflow Architecture

## Context

Both frontend and backend need consistent TDD practices. Current state:
- Backend: Table-driven tests exist but not consistently applied
- Frontend: Some hook tests, but no systematic TDD

## Decision

Implement **Red-Green-Refactor** workflow for both stacks:

### Backend (Go)
1. **Red**: Write failing table-driven test
2. **Green**: Implement minimal code to pass
3. **Refactor**: Clean up with confidence

Pattern:
```go
func TestService(t *testing.T) {
    tests := []struct {
        name     string
        input    Input
        want     Output
        wantErr  bool
    }{
        {name: "success", input: Input{...}, want: Output{...}},
        {name: "error case", input: Input{...}, wantErr: true},
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, err := Service(tt.input)
            if (err != nil) != tt.wantErr {
                t.Errorf("unexpected error: %v", err)
            }
            if diff := cmp.Diff(tt.want, got); diff != "" {
                t.Errorf("mismatch (-want +got):\n%s", diff)
            }
        })
    }
}
```

### Frontend (TypeScript)
1. **Red**: Write failing test with Vitest
2. **Green**: Implement minimal component/hook
3. **Refactor**: Clean up with test safety

Pattern:
```typescript
import { renderHook, waitFor } from '@testing-library/react'
import { useHook } from './useHook'

describe('useHook', () => {
  it('should return expected value', async () => {
    const { result } = renderHook(() => useHook())
    await waitFor(() => {
      expect(result.current.data).toEqual(expected)
    })
  })
})
```

## Alternatives Considered

| Option | Pros | Cons |
|--------|------|------|
| Red-Green-Refactor | Industry standard, proven | Requires discipline |
| Test-after | Faster initial dev | Lower coverage, design flaws |
| BDD (Gherkin) | Business-readable | Overhead for this project |

## Consequences

- **Positive**: Consistent quality, regression prevention
- **Negative**: Initial development slower
- **Risks**: Team adoption requires training

> **Pre-Tauri Note**: This ADR was written for the current web-app architecture. Naming conventions remain valid in the Tauri migration.
