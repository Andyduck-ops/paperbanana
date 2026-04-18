---
id: ADR-002
status: Accepted
traces_to: [REQ-002]
date: 2026-04-08
---

# ADR-002: OpenAPI Generation Strategy

## Context

We need a reliable way to generate OpenAPI specifications from Go code that:
- Stays synchronized with code changes
- Requires minimal maintenance
- Integrates well with Gin framework

## Decision

Use **swaggo/swag** for OpenAPI generation:

1. **Annotations**: Add swaggo comments to Gin handlers
2. **DTOs**: Define request/response structs with json tags
3. **Generation**: Run `swag init` in CI to regenerate spec
4. **Validation**: Validate spec completeness in PR checks

Example annotation:
```go
// @Summary Generate figure
// @Tags generation
// @Accept json
// @Produce json
// @Param request body GenerateRequest true "Generation parameters"
// @Success 200 {object} GenerateResponse
// @Router /api/v1/generate [post]
func (h *Handler) Generate(c *gin.Context) { ... }
```

## Alternatives Considered

| Option | Pros | Cons |
|--------|------|------|
| swaggo | Native Gin support, widely used | Requires annotations |
| Manual spec | Full control | Maintenance burden, drift |
| go-swagger | Feature rich | Complex, steeper learning curve |

## Consequences

- **Positive**: Automated, integrates well, industry standard
- **Negative**: Requires discipline to maintain annotations
- **Risks**: Missing annotations lead to incomplete spec

> **Pre-Tauri Note**: This ADR was written for the current web-app architecture. Naming conventions remain valid in the Tauri migration.
