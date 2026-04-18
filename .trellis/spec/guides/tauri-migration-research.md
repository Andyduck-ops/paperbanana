# Tauri Migration Research

> Research findings for migrating PaperBanana from web-app (Go + React) to Tauri desktop application with Go sidecar.

**Date:** 2026-04-17
**Status:** Research Complete

---

## 1. Executive Summary

**Recommendation: Proceed with Tauri v2 + Go sidecar.** The migration is low-risk because:

- All Go backend code works as-is (HTTP API unchanged)
- SSE streaming over localhost works in Tauri webviews (requires CSP configuration)
- React frontend can be preserved (no Svelte rewrite needed)
- Sidecar lifecycle is well-supported in Tauri v2

**Alternative considered:** Wails (Go-native desktop framework) offers simpler Go-to-frontend IPC but has a smaller ecosystem, less mature tooling, and would require restructuring the entire Go backend into Wails' binding model. Tauri's sidecar approach preserves the existing HTTP architecture with minimal changes.

---

## 2. Architecture: Tauri + Go Sidecar

### Current Architecture

```
Browser (React SPA, :5173)
    ↓ HTTP/SSE (proxy → :8080)
Go Server (Gin + GORM + SQLite)
    ↓ File I/O + HTTP
SQLite + Assets + LLM Providers
```

### Target Architecture

```
┌─────────────────────────────────────┐
│  Tauri Desktop App (single binary)  │
├─────────────────────────────────────┤
│  React Frontend (webview)           │
│  - Same components, same hooks      │
│  - API base URL = localhost:auto    │
└──────────────┬──────────────────────┘
               │ HTTP/SSE (localhost:auto-port)
┌──────────────▼──────────────────────┐
│  Go Backend (sidecar binary)        │
│  - Same Gin router, same handlers   │
│  - Auto-selects available port      │
│  - Prints port to stdout on startup │
└──────────────┬──────────────────────┘
               │ File I/O + HTTP
         SQLite + Assets + LLM Providers
```

---

## 3. Integration Patterns

### 3.1 Sidecar Configuration

In `tauri.conf.json`:
```json
{
  "bundle": {
    "externalBin": ["binaries/paperbanana-server"]
  }
}
```

Binary naming convention (REQUIRED by Tauri):
```
src-tauri/binaries/
├── paperbanana-server-x86_64-pc-windows-msvc.exe    # Windows x64
├── paperbanana-server-x86_64-apple-darwin            # macOS Intel
├── paperbanana-server-aarch64-apple-darwin           # macOS Apple Silicon
├── paperbanana-server-x86_64-unknown-linux-gnu       # Linux x64
└── paperbanana-server-aarch64-unknown-linux-gnu      # Linux ARM64
```

### 3.2 Spawning the Sidecar from Rust

```rust
use tauri::Manager;
use tauri_plugin_shell::ShellExt;
use tauri_plugin_shell::process::CommandEvent;

#[tauri::command]
async fn start_backend(app: tauri::AppHandle) -> Result<u16, String> {
    let (rx, mut child) = app
        .shell()
        .sidecar("paperbanana-server")
        .map_err(|e| e.to_string())?
        .spawn()
        .map_err(|e| e.to_string())?;

    // Read port from stdout
    let port = tokio::select! {
        Some(event) = rx.recv() => {
            match event {
                CommandEvent::Stdout(line) => {
                    let output = String::from_utf8_lossy(&line);
                    // Parse "BACKEND_PORT=8080" from Go server stdout
                    output.trim()
                        .strip_prefix("BACKEND_PORT=")
                        .and_then(|p| p.parse::<u16>().ok())
                        .ok_or("Failed to parse backend port")?
                }
                CommandEvent::Terminated(status) => {
                    return Err(format!("Backend exited with status: {:?}", status));
                }
                _ => continue,
            }
        }
    };

    Ok(port)
}
```

### 3.3 Go Sidecar Changes

```go
// cmd/server/main.go additions

func findAvailablePort(start int) int {
    for port := start; port < start+100; port++ {
        if listener, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", port)); err == nil {
            listener.Close()
            return port
        }
    }
    return start
}

// In main():
port := findAvailablePort(8080)
address := fmt.Sprintf("127.0.0.1:%d", port)

// Signal port to Tauri (printed to stdout, parsed by Rust)
fmt.Printf("BACKEND_PORT=%d\n", port)
```

### 3.4 IPC Mechanism Decision

| Mechanism | Pros | Cons | Verdict |
|-----------|------|------|---------|
| HTTP over localhost | Zero code changes, SSE works, same API | Slightly higher latency than IPC | **USE THIS** |
| Tauri invoke + events | Lower latency, native feel | Requires Rust→Go bridge, breaks SSE model | Too much rework |
| WebSocket | Bidirectional, real-time | Requires rewriting SSE to WS | Unnecessary |
| Stdin/stdout pipes | Direct | Unreliable for complex payloads | Not recommended |

**Decision: Keep HTTP over localhost.** The existing SSE streaming model works perfectly. There is no need to introduce a Rust→Go IPC bridge when localhost HTTP achieves the same result with zero business logic changes.

---

## 4. SSE Streaming Considerations

### 4.1 Does SSE Work in Tauri Webviews?

**Yes, with configuration.** SSE over localhost works in all three Tauri webview engines:
- **Windows**: WebView2 (Chromium-based) -- full SSE support
- **macOS**: WKWebView (WebKit) -- full SSE support
- **Linux**: WebKitGTK -- full SSE support

### 4.2 CSP Configuration Required

Tauri v2 does NOT enforce CSP by default, but you MUST configure it to allow localhost connections:

```json
// tauri.conf.json
{
  "app": {
    "security": {
      "csp": "default-src 'self'; connect-src 'self' http://127.0.0.1:* http://localhost:* ipc: http://ipc.localhost; style-src 'self' 'unsafe-inline'"
    }
  }
}
```

Key directives:
- `connect-src http://127.0.0.1:*` -- allows fetch() and EventSource to localhost
- `style-src 'unsafe-inline'` -- needed for Tailwind CSS dynamic classes

### 4.3 Gotchas

1. **CORS**: The Go server must allow requests from the Tauri webview origin. Add `AllowOrigins: []string{"tauri://localhost", "https://tauri.localhost"}` to Gin CORS config.

2. **WebKitGTK buffering**: Some WebKitGTK versions buffer SSE events. Mitigation: use `X-Accel-Buffering: no` and `Cache-Control: no-cache` headers on the SSE endpoint.

3. **Connection timeout**: Default webview timeout may be shorter than long-running generation. The existing heartbeat mechanism in `sse.ts` should handle this.

4. **No native SSE plugin**: Tauri doesn't have a built-in SSE plugin (GitHub issue #14551 is open). HTTP-based SSE is the standard approach.

---

## 5. Sidecar Lifecycle Management

### 5.1 Lifecycle Flow

```
App Start
    ↓
Rust: spawn Go sidecar via shell().sidecar().spawn()
    ↓
Rust: read BACKEND_PORT from stdout
    ↓
Rust: poll GET /health until 200 OK (with retries)
    ↓
Rust: store port in Tauri state
    ↓
Frontend: query get_backend_port() Tauri command
    ↓
Frontend: set API_BASE = http://127.0.0.1:{port}/api/v1
    ↓
App Ready
    ↓
App Close
    ↓
Rust: send SIGTERM to Go process
    ↓
Go: graceful shutdown (drain in-flight requests, 30s timeout)
    ↓
Rust: force kill if not exited after timeout
```

### 5.2 Health Check

```rust
async fn wait_for_health(port: u16, max_retries: u32) -> Result<(), String> {
    let client = reqwest::Client::new();
    for i in 0..max_retries {
        if let Ok(resp) = client
            .get(format!("http://127.0.0.1:{}/health", port))
            .timeout(Duration::from_secs(2))
            .send()
            .await
        {
            if resp.status().is_success() {
                return Ok(());
            }
        }
        tokio::time::sleep(Duration::from_millis(500)).await;
    }
    Err("Backend health check failed".into())
}
```

### 5.3 Graceful Shutdown

Go sidecar already has graceful shutdown via `cmd/server/main.go`. Tauri will clean up child processes on app exit. For explicit control:

```rust
app.on_window_event(|window, event| {
    if let tauri::WindowEvent::CloseRequested { .. } = event {
        // Send SIGTERM to sidecar
        if let Some(mut child) = state.sidecar_child.lock().unwrap().take() {
            let _ = child.kill();
        }
    }
});
```

---

## 6. Frontend Framework Decision: React vs Svelte

### Decision Matrix

| Criterion | React (Keep) | Svelte (Rewrite) | Weight |
|-----------|--------------|-------------------|--------|
| Existing codebase | Full (30+ components, 10+ hooks) | None | **HIGH** |
| Migration effort | Minimal (URL detection change) | Full rewrite (weeks) | **HIGH** |
| Bundle size | ~150KB (gzipped) | ~15KB (gzipped) | LOW |
| Tauri community examples | Many | Fewer | MEDIUM |
| DX with Tauri | Good | Good | LOW |
| Future maintenance | Proven, large ecosystem | Growing, smaller | MEDIUM |
| Startup time impact | <100ms difference | <50ms difference | LOW |

### Recommendation: **Keep React**

Rationale:
- The existing React codebase has 30+ components, 10+ hooks, complete i18n, and theme system
- Svelte's bundle size advantage (~135KB savings) is negligible in a desktop app context (no download concern)
- The startup time difference is imperceptible in a Tauri app (both render in <100ms)
- A full Svelte rewrite would take weeks and introduce regressions
- React has better TypeScript tooling (existing project uses TS 5.9)

The only scenario where Svelte makes sense is if the project is planning a complete frontend redesign from scratch. Given that the Go backend is being preserved as-is, there's no reason to rewrite a working frontend.

---

## 7. Cross-Platform Build Pipeline

### 7.1 Go Sidecar Cross-Compilation

```bash
# Windows x64
GOOS=windows GOARCH=amd64 go build -o src-tauri/binaries/paperbanana-server-x86_64-pc-windows-msvc.exe ./cmd/server

# macOS Intel
GOOS=darwin GOARCH=amd64 go build -o src-tauri/binaries/paperbanana-server-x86_64-apple-darwin ./cmd/server

# macOS Apple Silicon
GOOS=darwin GOARCH=arm64 go build -o src-tauri/binaries/paperbanana-server-aarch64-apple-darwin ./cmd/server

# Linux x64
GOOS=linux GOARCH=amd64 go build -o src-tauri/binaries/paperbanana-server-x86_64-unknown-linux-gnu ./cmd/server

# Linux ARM64
GOOS=linux GOARCH=arm64 go build -o src-tauri/binaries/paperbanana-server-aarch64-unknown-linux-gnu ./cmd/server
```

### 7.2 GitHub Actions Workflow

```yaml
name: Build Tauri App
on:
  push:
    tags: ['v*']

jobs:
  build:
    strategy:
      matrix:
        include:
          - platform: macos-latest
            target: aarch64-apple-darwin
          - platform: macos-latest
            target: x86_64-apple-darwin
          - platform: ubuntu-22.04
            target: x86_64-unknown-linux-gnu
          - platform: windows-latest
            target: x86_64-pc-windows-msvc

    runs-on: ${{ matrix.platform }}
    steps:
      - uses: actions/checkout@v4

      # Build Go sidecar
      - uses: actions/setup-go@v5
        with: { go-version: '1.23' }
      - run: |
          mkdir -p src-tauri/binaries
          GOOS=${TARGET%%-*} GOARCH=${TARGET%%-*} go build -o src-tauri/binaries/paperbanana-server-${{ matrix.target }}${{ matrix.platform == 'windows-latest' && '.exe' || '' }} ./cmd/server

      # Build Tauri app
      - uses: actions/setup-node@v4
        with: { node-version: 22 }
      - run: cd web && npm install
      - uses: dtolnay/rust-toolchain@stable
      - uses: tauri-apps/tauri-action@v0
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
        with:
          tagName: ${{ github.ref_name }}
          releaseName: 'PaperBanana ${{ github.ref_name }}'
```

### 7.3 Code Signing

- **macOS**: Requires Apple Developer certificate + notarization. Use `rcodesign` for CI notarization.
- **Windows**: Requires code signing certificate. Use `signtool` or Azure Key Vault signing.
- **Linux**: No code signing required. Distribute as AppImage or .deb.

---

## 8. Auto-Update Strategy

### 8.1 Tauri Updater Plugin

Tauri v2 has a built-in updater plugin that:
- Queries configured endpoints for version checks
- Verifies cryptographic signatures (mandatory)
- Downloads and installs the new app package

### 8.2 Sidecar Update Limitation

**The Tauri updater does NOT update sidecar binaries.** It only updates the main application package.

Workaround options:
1. **Bundle sidecar in app package** -- The sidecar is already bundled inside the Tauri app. When the app updates, the entire package (including sidecar) is replaced. This is the default behavior.
2. **Self-updating sidecar** -- Add a `/api/v1/update` endpoint to the Go server that checks for and downloads a newer sidecar binary. Restart the Go process after update.
3. **Version mismatch detection** -- On startup, compare Go sidecar version with the expected version. If mismatch, prompt user to download the full app update.

**Recommendation: Option 1 (default behavior).** Since the sidecar is bundled inside the Tauri app package, updating the app automatically updates the sidecar. This is the simplest and most reliable approach.

### 8.3 Database Migration During Update

When the app updates and the Go sidecar starts with a new schema version, GORM's `AutoMigrate` will apply schema changes automatically. For breaking changes:
- Add a migration version check in `bootstrap.go`
- Log migration results
- Keep backward-compatible column additions (GORM handles this)

---

## 9. Migration Path

### Step-by-Step Plan

1. **Create Tauri scaffold** (`src-tauri/` directory with Cargo.toml, tauri.conf.json, main.rs)
2. **Configure Go sidecar** (modify `cmd/server/main.go` for port auto-detection and stdout port signaling)
3. **Implement sidecar spawning in Rust** (spawn, read port, health check, store in state)
4. **Update frontend API base URL** (add Tauri environment detection to `web/src/lib/api.ts`)
5. **Configure CSP** (allow localhost connections in tauri.conf.json)
6. **Add CORS to Go server** (allow tauri://localhost origin)
7. **Test SSE streaming** (verify in all three webview engines)
8. **Set up cross-platform builds** (GitHub Actions workflow)
9. **Code signing** (Apple Developer cert, Windows cert)
10. **Configure auto-updater** (Tauri updater plugin + JSON manifest endpoint)

### Files to Create

```
src-tauri/
├── Cargo.toml
├── tauri.conf.json
├── icons/
├── binaries/              # Go sidecar binaries (platform-specific)
└── src/
    ├── main.rs            # Tauri entry point
    ├── sidecar.rs         # Sidecar spawn + port detection
    └── commands.rs        # Tauri commands (get_backend_port, etc.)
```

### Files to Modify

```
cmd/server/main.go          # Port auto-detection, BACKEND_PORT stdout
web/src/lib/api.ts          # Tauri-aware API base URL
web/vite.config.ts          # Tauri dev plugin, proxy update
web/package.json            # Add @tauri-apps/api, @tauri-apps/cli
.gitignore                  # Add src-tauri/target/, binaries/
```

### Files NOT Changing

All `internal/` Go packages (domain, application, infrastructure) and all `web/src/components/`, `web/src/hooks/` (except api.ts URL logic) remain unchanged.

---

## 10. Risks and Open Questions

### Risks

| Risk | Probability | Impact | Mitigation |
|------|-------------|--------|------------|
| SSE buffering on WebKitGTK | Medium | High | Add `X-Accel-Buffering: no` header, test on Linux early |
| CSP blocking localhost | Low | High | Configure CSP explicitly in tauri.conf.json |
| Sidecar port collision | Low | Medium | Scan 8080-8180 range |
| macOS code signing delay | High | Medium | Start certificate process early |
| Sidecar crash on startup | Low | High | Health check with retries, error UI in frontend |
| Binary size bloat (Go + Rust) | Low | Low | Use `go build -ldflags="-s -w"`, UPX compression |

### Open Questions

1. **Apple Silicon universal binary** -- Should we ship a universal binary or separate arm64/x64? Universal is simpler for users but doubles the macOS binary size.
2. **Database location** -- Should we use Tauri's `app_data_dir()` for the SQLite database? This is the standard location on each platform.
3. **Asset storage location** -- Similar to database, move to `app_data_dir()/assets/` for platform convention compliance.
4. **Docker support** -- Do we continue supporting Docker deployment alongside Tauri? The web-app deployment path can remain as a separate build target.
5. **Multiple instances** -- Should the Tauri app prevent multiple instances? If so, use single-instance plugin.

---

## 11. References

- [Tauri v2 Sidecar Documentation](https://v2.tauri.app/develop/sidecar/)
- [Tauri v2 CSP Documentation](https://v2.tauri.app/security/csp/)
- [Tauri v2 GitHub Actions Pipeline](https://v2.tauri.app/distribute/pipelines/github/)
- [Tauri v2 Updater Plugin](https://v2.tauri.app/plugin/updater/)
- [Evil Martians: Rust + Tauri + Sidecar](https://evilmartians.com/chronicles/making-desktop-apps-with-revved-up-potential-rust-tauri-sidecar)
- [Tauri SSE Feature Request (GitHub #14551)](https://github.com/tauri-apps/tauri/issues/14551)
- [Wails vs Tauri Comparison](https://dev.to/arashgl/taurirust-vs-wailsgo-4pd6)
- [Tauri v2 Python Sidecar Example](https://github.com/dieharders/example-tauri-v2-python-server-sidecar)

---

*Research completed 2026-04-17. This document should be updated as implementation proceeds and new findings emerge.*
