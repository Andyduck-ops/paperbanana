# Security Audit Decisions Log

## 2026-03-25: Security Vulnerability Prioritization

### Decision: Python Execution Sandbox Architecture
**Context**: The plot executor uses `exec()` in Python to run user-provided code, creating a critical RCE vulnerability.

**Options Considered**:
1. RestrictedPython + AST validation
2. Docker container isolation
3. seccomp/AppArmor sandboxing
4. Remove dynamic execution entirely

**Decision**: Recommend Docker container isolation as primary remediation, with AST validation as defense-in-depth.

**Rationale**:
- Container isolation provides strongest security boundary
- Allows resource limiting (CPU, memory, time)
- Maintains flexibility for legitimate use cases
- AST validation adds additional layer of protection

---

### Decision: Encryption Key Management Strategy
**Context**: Development keys are persisted to filesystem and may be committed to VCS.

**Decision**: Require explicit encryption key configuration in production.

**Rationale**:
- Fail-fast principle for security misconfigurations
- Development mode should be clearly distinguished
- Existing `.gitignore` entry provides secondary protection

---

### Decision: CORS Configuration Policy
**Context**: Default CORS allows all origins.

**Decision**: Environment-aware CORS configuration.

**Rationale**:
- Development: Permissive for local testing
- Production: Strict allowlist from configuration
- No wildcard with credentials (already correct)
