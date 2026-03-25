# Phase 4: Security Hardening - Plan

**Created:** 2026-03-24
**Status:** Ready for execution

## Goals

1. Add API key authentication
2. Implement session management
3. Add rate limiting per key
4. Add input validation
5. Add CORS configuration
6. Add audit logging

## Tasks

### Task 4.1: API Key Authentication

**Priority:** P1
**Files:**
- `internal/api/middleware/auth.go` (new) - Authentication middleware
- `internal/config/config.go` - Add API key config

**Implementation:**
1. Create auth middleware with API key validation:
   - Read API key from `X-API-Key` header
   - Validate against configured keys
   - Return 401 Unauthorized if invalid

2. Add API key configuration:
   - `auth.api_keys` - List of valid API keys
   - `auth.enabled` - Toggle authentication

**Acceptance Criteria:**
- API key validation works
- Unauthorized requests return 401
- Can be disabled for development

### Task 4.2: Rate Limiting

**Priority:** P1
**Files:**
- `internal/api/middleware/ratelimit.go` (new) - Rate limiting middleware
- `internal/config/config.go` - Add rate limit config

**Implementation:**
1. Create rate limiter using token bucket algorithm:
   - Per-API-key rate limiting
   - Configurable requests per minute
   - Return 429 Too Many Requests when exceeded

2. Add rate limit configuration:
   - `rate_limit.requests_per_minute` - Default: 60
   - `rate_limit.burst` - Default: 10

**Acceptance Criteria:**
- Rate limiting works per API key
- Rate limit headers included in response
- 429 returned when limit exceeded

### Task 4.3: Input Validation

**Priority:** P1
**Files:**
- `internal/api/middleware/validation.go` (new) - Input validation middleware
- `internal/api/dto/` - Add validation tags

**Implementation:**
1. Add request size limits:
   - Max body size: 10MB
   - Max prompt length: 100KB

2. Validate file uploads:
   - Allowed content types
   - Max file size from config

3. Sanitize prompt inputs:
   - Remove control characters
   - Limit length

**Acceptance Criteria:**
- Large requests rejected with 413
- Invalid content types rejected
- Inputs sanitized

### Task 4.4: CORS Configuration

**Priority:** P1
**Files:**
- `internal/api/middleware/cors.go` (new) - CORS middleware
- `internal/config/config.go` - Add CORS config

**Implementation:**
1. Add CORS middleware:
   - Configurable allowed origins
   - Allow credentials
   - Expose headers

2. Add CORS configuration:
   - `cors.allowed_origins` - List of origins
   - `cors.allowed_methods` - HTTP methods
   - `cors.allowed_headers` - Headers

**Acceptance Criteria:**
- CORS headers set correctly
- Preflight requests handled
- Configurable origins

### Task 4.5: Audit Logging

**Priority:** P1
**Files:**
- `internal/api/middleware/audit.go` (new) - Audit logging middleware

**Implementation:**
1. Log all API requests:
   - Request ID
   - Method and path
   - API key (hashed)
   - Response status
   - Duration

2. Include request ID in response headers

**Acceptance Criteria:**
- All requests logged
- Request ID traceable
- Sensitive data not logged

## Verification

- Authentication works with valid/invalid keys
- Rate limiting returns 429 when exceeded
- Input validation rejects invalid requests
- CORS headers present
- Audit logs captured

## Files Changed

| File | Change |
|------|--------|
| `internal/api/middleware/auth.go` | New - API key authentication |
| `internal/api/middleware/ratelimit.go` | New - Rate limiting |
| `internal/api/middleware/validation.go` | New - Input validation |
| `internal/api/middleware/cors.go` | New - CORS handling |
| `internal/api/middleware/audit.go` | New - Audit logging |
| `internal/api/router.go` | Modify - Add middleware |
| `internal/config/config.go` | Modify - Add security config |
