# Paperbanana Security Audit Report

**Audit Date**: 2026-03-25
**Scope**: Deep security vulnerability scan of paperbanana-clean codebase
**OWASP Top 10 Reference**: Full coverage analysis

---

## Executive Summary

This security audit identified **8 vulnerabilities** across the codebase:

| Severity | Count | Description |
|----------|-------|-------------|
| **CRITICAL** | 1 | Arbitrary Code Execution via Python exec() |
| **HIGH** | 2 | Encryption key management, CORS misconfiguration |
| **MEDIUM** | 3 | Input validation gaps, logging of sensitive data, development key warning |
| **LOW** | 2 | Missing security headers, rate limiting implementation gaps |

---

## Vulnerability Details

### CRITICAL-001: Arbitrary Code Execution via Python exec()

**OWASP Category**: A03:2021 - Injection

**Location**: `internal/application/agents/visualizer/plot_executor.go:37-82`

**Description**:
The `pythonPlotExecutor.Execute()` function passes user-provided code directly to Python's `exec()` function without any sandboxing or code validation. This is a classic Remote Code Execution (RCE) vulnerability.

**Vulnerable Code**:
```go
// Line 37-82
cmd := exec.CommandContext(ctx, e.command, "-c", plotExecutorScript)
cmd.Stdin = strings.NewReader(cleaned)  // 'cleaned' is user code

// Python script (line 68-91)
exec(code, namespace)  // Direct execution of arbitrary Python code
```

**Attack Vector**:
1. User sends malicious Python code through the `/api/v1/generate` endpoint
2. The code is passed to Python via stdin
3. `exec()` executes the code with the privileges of the server process
4. Attacker gains full system access

**Exploitation Example**:
```python
# User input could contain:
import os
os.system("rm -rf /")  # or any malicious command
```

**Impact**:
- Full system compromise
- Data exfiltration
- Lateral movement within infrastructure
- Complete service disruption

**Remediation**:
1. **Immediate**: Implement a restricted Python environment:
   - Use `RestrictedPython` or similar sandboxing library
   - Whitelist allowed Python modules (matplotlib, numpy only)
   - Remove access to os, sys, subprocess, socket modules

2. **Medium-term**: Containerization:
   - Run Python execution in isolated containers (Docker)
   - Implement resource limits (CPU, memory, time)
   - Use seccomp/AppArmor profiles

3. **Long-term**: Code validation:
   - Parse and validate AST before execution
   - Implement code signing for trusted templates
   - Add user permission levels for code execution

**References**:
- OWASP A03:2021 - Injection
- CWE-95: Improper Neutralization of Directives in Dynamically Evaluated Code

---

### HIGH-001: Development Encryption Key Exposure

**OWASP Category**: A02:2021 - Cryptographic Failures

**Location**: `internal/infrastructure/crypto/aesgcm/service.go:117-143`

**Description**:
When `PAPERBANANA_ENCRYPTION_KEY` environment variable is not set, the system creates and persists a development encryption key in `.paperbanana/dev-encryption.key`. This file:
1. Is created with file permissions `0600` but may be committed to version control
2. Uses a warning printed to stdout (not structured logging)
3. Creates the same key for all instances in a multi-instance deployment

**Vulnerable Code**:
```go
// Line 117-143
func loadOrCreateDevKey() (string, error) {
    path := os.Getenv("PAPERBANANA_ENCRYPTION_KEY_FILE")
    if strings.TrimSpace(path) == "" {
        path = defaultDevKeyPath  // ".paperbanana/dev-encryption.key"
    }
    // ...
    fmt.Printf("WARNING: PAPERBANANA_ENCRYPTION_KEY not set, using persisted development key at %s\n", path)
```

**Impact**:
- If `.paperbanana/` is committed, encryption keys are exposed
- Multiple instances share the same encryption key
- Keys may be backed up to CI/CD systems

**Remediation**:
1. Require `PAPERBANANA_ENCRYPTION_KEY` in production builds (fail-fast)
2. Add `.paperbanana/` to `.gitignore` (already present)
3. Use structured logging instead of `fmt.Printf`
4. Implement key rotation mechanism
5. Add environment detection (fail in production mode)

---

### HIGH-002: Permissive CORS Configuration

**OWASP Category**: A01:2021 - Broken Access Control

**Location**: `internal/api/middleware/cors.go:21-30`

**Description**:
The default CORS configuration allows all origins (`*`) which enables any website to make cross-origin requests to the API.

**Vulnerable Code**:
```go
// Line 21-30
func DefaultCORSConfig() CORSConfig {
    return CORSConfig{
        AllowedOrigins:   []string{"*"},
        // ...
        AllowCredentials: false,  // Good: prevents credential leakage
    }
}
```

**Impact**:
- Any malicious website can call the API
- CSRF-like attacks possible on unauthenticated endpoints
- Data exfiltration through victim's browser

**Remediation**:
1. Implement strict origin allowlist in production
2. Read allowed origins from configuration
3. Different CORS policies for different environments
4. Consider implementing CSRF tokens for state-changing operations

---

### MEDIUM-001: Insufficient Input Validation

**OWASP Category**: A03:2021 - Injection

**Location**: `internal/api/handlers/generate.go:187-205`

**Description**:
The `validateGenerateRequest()` function only validates basic constraints:
- Resume requires session ID
- At least one of prompt/content/visual_intent is required
- Critic rounds limited to 0-5

Missing validations:
- No length limits on prompt content
- No character encoding validation
- No URL validation in content references
- No model name validation against allowed list

**Impact**:
- Large payload attacks
- Potential injection vectors
- Resource exhaustion

**Remediation**:
1. Add maximum length limits for all text fields
2. Validate against allowed model list
3. Sanitize special characters
4. Add content-type validation for embedded data

---

### MEDIUM-002: Potential Sensitive Data in Logs

**OWASP Category**: A09:2021 - Security Logging and Monitoring Failures

**Location**: Multiple files using zap logger

**Description**:
While the codebase uses structured logging (zap), there are potential areas where sensitive data might be logged:

1. Error messages may contain API keys or user content:
   - `internal/api/handlers/generate.go:176`: `h.logger.Error("generation failed", zap.Error(err))`
   - Error messages from external services may contain sensitive data

2. Request metadata logging in audit middleware:
   - `internal/api/middleware/audit.go`: Logs query parameters which may contain sensitive data

**Observations**:
- API keys are properly masked in responses (`MaskAPIKey()`)
- No direct logging of API key plaintext values found
- Request bodies are not logged (good)

**Remediation**:
1. Implement sensitive data scrubbing in error handlers
2. Add PII detection for log output
3. Consider log retention policies
4. Add audit logging for security-relevant events

---

### MEDIUM-003: Hardcoded Python Command

**OWASP Category**: A06:2021 - Vulnerable and Outdated Components

**Location**: `internal/application/agents/visualizer/plot_executor.go:27-28`

**Description**:
The Python interpreter path is hardcoded as `"python3"`:

```go
func NewPlotExecutor() PlotExecutor {
    return pythonPlotExecutor{command: "python3"}
}
```

**Impact**:
- Fails if `python3` is not in PATH
- Cannot use virtual environments or specific Python versions
- No flexibility for deployment scenarios

**Remediation**:
1. Make Python path configurable via environment variable
2. Add Python version validation
3. Implement fallback paths (`python3.11`, `python3.10`, etc.)
4. Validate required Python packages are installed

---

### LOW-001: Missing Security Headers

**OWASP Category**: A05:2021 - Security Misconfiguration

**Location**: `internal/api/router.go`

**Description**:
The API router does not set recommended security headers:
- `X-Content-Type-Options: nosniff`
- `X-Frame-Options: DENY`
- `Content-Security-Policy`
- `Strict-Transport-Security` (for HTTPS)

**Remediation**:
Add a security headers middleware:
```go
func SecurityHeaders() gin.HandlerFunc {
    return func(c *gin.Context) {
        c.Header("X-Content-Type-Options", "nosniff")
        c.Header("X-Frame-Options", "DENY")
        c.Header("X-XSS-Protection", "1; mode=block")
        c.Next()
    }
}
```

---

### LOW-002: In-Memory Rate Limiting

**OWASP Category**: A07:2021 - Identification and Authentication Failures

**Location**: `internal/api/middleware/ratelimit.go`

**Description**:
Rate limiting is implemented using in-memory token buckets. This has limitations:
- Does not work across multiple instances
- Lost on restart
- Can be bypassed by distributed attacks

**Positive Observations**:
- Uses constant-time comparison for API keys
- Properly implements token bucket algorithm
- Has cleanup mechanism for old buckets

**Remediation**:
1. Use Redis-based distributed rate limiting (already have Redis client)
2. Consider IP-based rate limiting with X-Forwarded-For validation
3. Add rate limit bypass for internal services

---

## Positive Security Findings

### Good Practices Identified

1. **Path Traversal Protection** (`internal/infrastructure/assets/localstore/store.go`):
   - Uses `filepath.IsLocal()` for path validation
   - Double-check with absolute path comparison
   - Proper error handling for traversal attempts

2. **SQL Injection Prevention**:
   - All database queries use parameterized queries via GORM
   - No string concatenation in SQL queries
   - Proper use of `?` placeholders

3. **API Key Security**:
   - Keys are encrypted at rest using AES-256-GCM
   - Argon2id key derivation with appropriate parameters
   - Proper key masking in API responses
   - No plaintext keys in logs

4. **Authentication**:
   - Constant-time comparison for API key validation
   - Proper error handling (generic error messages)
   - Support for Bearer token authentication

5. **Encryption Implementation**:
   - AES-256-GCM with proper nonce handling
   - Nonce prepended to ciphertext for deterministic decryption
   - Proper base64 encoding

---

## OWASP Top 10 2021 Coverage

| Category | Status | Notes |
|----------|--------|-------|
| A01: Broken Access Control | PARTIAL | CORS misconfiguration, no CSRF protection |
| A02: Cryptographic Failures | GOOD | Strong encryption, key derivation issues noted |
| A03: Injection | CRITICAL | Python code execution vulnerability |
| A04: Insecure Design | N/A | Architecture-level concerns outside scope |
| A05: Security Misconfiguration | PARTIAL | Missing security headers |
| A06: Vulnerable Components | REVIEW | Python version flexibility needed |
| A07: Auth Failures | GOOD | Proper API key handling |
| A08: Data Integrity | N/A | No code signing or integrity checks |
| A09: Logging & Monitoring | PARTIAL | Good logging, potential sensitive data exposure |
| A10: SSRF | N/A | No external URL fetching detected |

---

## Recommendations Priority Matrix

| Priority | Action | Effort | Risk Reduction |
|----------|--------|--------|----------------|
| P0 | Sandbox Python execution | High | Critical |
| P1 | Force encryption key in production | Low | High |
| P1 | Implement strict CORS | Low | High |
| P2 | Add security headers | Low | Medium |
| P2 | Implement distributed rate limiting | Medium | Medium |
| P3 | Input validation enhancement | Medium | Medium |
| P3 | Log sanitization | Medium | Low |

---

## Conclusion

The most critical finding is the arbitrary code execution vulnerability in the Python plot executor. This should be addressed immediately before any production deployment. The encryption and authentication mechanisms are well-implemented, but the CORS configuration and missing security headers need attention.

**Risk Level**: HIGH (due to RCE vulnerability)

**Recommended Actions**:
1. Do not deploy to production without addressing CRITICAL-001
2. Implement all P1 recommendations within 1 week
3. Schedule P2/P3 items for next sprint

---

*Report generated by security audit agent*
*Session: TLV4-drift-deep-dive-20260325*
