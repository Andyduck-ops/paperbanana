# PaperBanana Deployment Guide

## Prerequisites

- Docker and Docker Compose
- Go 1.23+ (for local development)
- Node.js 18+ (for frontend development)

## Quick Start

### Using Docker Compose (Recommended)

1. Build and start all services:
   ```bash
   docker-compose up --build
   ```

2. Access the application:
   - Frontend: http://localhost:3000
   - Backend API: http://localhost:8080/api/v1
   - Health check: http://localhost:8080/health

### Manual Setup

1. Start the backend:
   ```bash
   go build -o server ./cmd/server
   ./server --config configs/config.yaml
   ```

2. Build and serve the frontend:
   ```bash
   cd web
   npm install
   npm run build
   # Serve dist/ with any static file server
   ```

## Configuration

### Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `GIN_MODE` | Gin framework mode (debug/release) | `debug` |
| `LOG_LEVEL` | Logging level (debug/info/warn/error) | `info` |

### Configuration File

Edit `configs/config.yaml` for:
- Server port and host
- LLM provider settings
- Database path
- Cache settings

## Health Endpoints

- `GET /health` - Basic liveness check
- `GET /ready` - Readiness check (includes dependency status)

## Architecture

```
┌─────────────┐     ┌─────────────┐     ┌─────────────┐
│   Browser   │────▶│    Nginx    │────▶│   Backend   │
└─────────────┘     └─────────────┘     │   (Go)      │
                    │                   └──────┬──────┘
                    │                          │
                    │                          ▼
                    │                   ┌─────────────┐
                    │                   │   SQLite    │
                    │                   └─────────────┘
                    │
                    │                   ┌─────────────┐
                    └──────────────────▶│    Redis    │
                                        │   (cache)   │
                                        └─────────────┘
```

## Monitoring

### Logs

Logs are structured JSON format. Configure log level via `LOG_LEVEL` environment variable.

### Health Checks

Docker includes health check for the backend:
- Interval: 30s
- Timeout: 3s
- Retries: 3

## Production Considerations

1. **Database**: For production, consider PostgreSQL instead of SQLite
2. **Cache**: Redis is configured but optional; enable in config
3. **SSL/TLS**: Configure nginx or use a reverse proxy (Traefik, Caddy)
4. **Rate Limiting**: Enable in Gin middleware for production
5. **Secrets**: Use environment variables or secret management for API keys

## Troubleshooting

### Backend won't start
- Check Go version (requires 1.23+)
- Verify config file exists
- Check port 8080 is available

### Frontend shows blank page
- Run `npm run build` in web/ directory
- Check nginx configuration
- Verify API proxy settings

### SSE streaming not working
- Check nginx buffering settings (should be disabled for SSE)
- Verify proxy_read_timeout is sufficient
