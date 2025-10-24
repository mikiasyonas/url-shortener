# Configuration Guide

## Environment Variables

### Server Configuration
- `SERVER_PORT` - Port to run the server on (default: 8080)
- `ENVIRONMENT` - Environment: development, production (default: development)
- `SERVER_READ_TIMEOUT` - Read timeout (default: 10s)
- `SERVER_WRITE_TIMEOUT` - Write timeout (default: 10s)
- `SERVER_IDLE_TIMEOUT` - Idle timeout (default: 60s)
- `CORS_ALLOWED_ORIGINS` - CORS allowed origins (default: *)

### Database Configuration
- `DB_HOST` - Database host (default: localhost)
- `DB_PORT` - Database port (default: 5432)
- `DB_USER` - Database user (default: postgres)
- `DB_PASSWORD` - Database password (default: password)
- `DB_NAME` - Database name (default: url_shortener)
- `DB_SSL_MODE` - SSL mode (default: disable)
- `DB_MAX_OPEN_CONNS` - Max open connections (default: 25)
- `DB_MAX_IDLE_CONNS` - Max idle connections (default: 25)
- `DB_CONN_MAX_LIFETIME` - Connection max lifetime (default: 5m)

### Application Configuration
- `APP_BASE_URL` - Base URL for short links (default: http://localhost:8080)
- `APP_SHORT_CODE_LENGTH` - Length of short codes (default: 6)
- `APP_MAX_URL_LENGTH` - Maximum URL length (default: 2048)
- `APP_RATE_LIMIT_PER_SECOND` - Rate limit per second (default: 100)

## Development Setup
1. Copy `.env.example` to `.env`
2. Update values as needed
3. Run `docker-compose up -d` to start database
4. Run `go run cmd/api/main.go`

## Production Setup
1. Copy `.env.production.example` to `.env`
2. Set all required values for production
3. Deploy with your preferred method