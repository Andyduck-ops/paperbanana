# Backend Testing

## Runner and Assertions

**Runner:** `go test ./...`

**Assertion library:** `github.com/stretchr/testify`

- Use `require` for setup and fatal assertions (stops test immediately)
- Use `assert` for detailed checks (reports failure but continues)
- Use `cmp.Diff` for structured comparison of complex types

```go
func TestService(t *testing.T) {
    // require for setup -- test cannot proceed if this fails
    cfg, err := Load()
    require.NoError(t, err)

    // assert for individual checks -- see all failures
    assert.Equal(t, "expected", cfg.Value)
    assert.Nil(t, cfg.Extra)
}
```

## Test File Organization

- Tests live next to implementation as `_test.go` files
- Same-package tests (no `_test` suffix) for white-box access

```text
internal/api/handlers/generate.go
internal/api/handlers/generate_test.go
internal/application/orchestrator/runner.go
internal/application/orchestrator/runner_test.go
```

## Patterns

### Table-Driven Tests with t.Run

The standard pattern for multi-case tests:

```go
func TestGenerateHandler(t *testing.T) {
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

### HTTP Handler Testing with httptest

Use real Gin routers with `httptest.NewRecorder()`:

```go
func TestGenerateHandlerRunsPipeline(t *testing.T) {
    router := gin.New()
    router.POST("/generate", handler.Generate)

    req := httptest.NewRequest(http.MethodPost, "/generate", bytes.NewReader(body))
    rec := httptest.NewRecorder()
    router.ServeHTTP(rec, req)

    require.Equal(t, http.StatusOK, rec.Code)
}
```

For transport-layer integration tests, use `httptest.NewServer()`:

```go
func TestTransportLayer(t *testing.T) {
    srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(http.StatusOK)
    }))
    defer srv.Close()
    // Test against srv.URL ...
}
```

### In-Memory SQLite and TempDir

For persistence-layer tests, use in-memory SQLite or `t.TempDir()`:

```go
func TestRepository(t *testing.T) {
    db := setupInMemoryDB(t)   // Helper that creates :memory: SQLite
    repo := NewRepository(db)
    // Test repo operations...
}
```

### Environment Variable Testing

Use `t.Setenv()` for environment-scoped tests:

```go
func TestConfigFromEnv(t *testing.T) {
    t.Setenv("PAPERBANANA_API_KEY", "test-key")
    cfg, err := Load()
    require.NoError(t, err)
}
```

## Mocking

### Use Handwritten Fakes, NOT Mock Frameworks

Backend tests use local fake structs defined inside `_test.go` files:

```go
type mockRunner struct {
    startFn  func(context.Context, domainagent.AgentInput) (RunHandle, error)
    resumeFn func(context.Context, domainagent.AgentInput) (RunHandle, error)
}

func (m *mockRunner) Start(ctx context.Context, input domainagent.AgentInput) (RunHandle, error) {
    return m.startFn(ctx, input)
}

func (m *mockRunner) Resume(ctx context.Context, input domainagent.AgentInput) (RunHandle, error) {
    return m.resumeFn(ctx, input)
}
```

### What to Mock

- LLM clients when unit-testing orchestration logic
- Repository interfaces when testing application services
- Runner interfaces when testing API handlers

### What NOT to Mock

- **Gin request routing** -- use real routers and recorders in handler tests
  (see `internal/api/handlers/generate_test.go`)
- **Repository behavior** when in-memory or TempDir-backed implementations
  are available -- prefer real storage over mocked behavior
- **HTTP transport** when the package contract IS the transport layer --
  use `httptest.NewServer()` instead

## Fixtures

- Inline in test files (preferred for small fixtures)
- `testdata/` directory for larger fixture files
- Factory helpers as local functions inside `_test.go` files

```go
func newMockProjectRepository() *mockProjectRepository {
    return &mockProjectRepository{projects: make(map[string]*workspace.Project)}
}
```

## Known Failures

As of 2026-03-25, `go test ./...` has two failing packages:

### internal/application/agents/critic

- `TestCriticRounds` expects `output.Metadata["reused_artifact"] == "false"`,
  but current code in `internal/application/agents/critic/agent.go` produces `"true"`
- Prompt fixture parity checks against `diagram_system.txt` and `plot_system.txt`
  differ by an added blank line before the JSON block

### internal/config

- `config_test.go` expects default provider `gemini`, strict API-key validation,
  and `EnableWAL == false`, but `config.go` now defaults `enable_wal` to `true`
  and current loading behavior differs from test assumptions

## Async Testing Pattern

```go
ctx, cancel := context.WithTimeout(context.Background(), time.Second)
defer cancel()

handle, err := runner.Start(ctx, testAgentInput())
require.NoError(t, err)
_, err = handle.Wait()
require.NoError(t, err)
```

## Error Testing Pattern

```go
cfg, err := Load()
require.Error(t, err)
assert.Nil(t, cfg)
assert.ErrorContains(t, err, "provider gemini missing api_key")
```
