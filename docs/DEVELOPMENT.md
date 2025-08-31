# Development Guide

## Prerequisites

- Go 1.22+
- Node.js 20+ and pnpm
- Docker and Docker Compose
- PostgreSQL client tools
- ClickHouse client (optional)
- Make

## Environment Setup

### 1. Clone the Repository

```bash
git clone https://github.com/supercakecrumb/tg-analysis-tool.git
cd tg-analysis-tool
```

### 2. Install Dependencies

Backend:
```bash
cd backend
go mod download
```

Frontend:
```bash
cd web
pnpm install
```

### 3. Environment Variables

Copy the example environment file:
```bash
cp .env.example .env
```

Required environment variables:
```env
# Database
DATABASE_URL=postgres://tg_user:password@localhost:5432/tg_analysis?sslmode=disable
CLICKHOUSE_URL=clickhouse://localhost:9000/tg_analytics

# Telegram API (get from https://my.telegram.org)
TELEGRAM_API_ID=your_api_id
TELEGRAM_API_HASH=your_api_hash

# Security
SESSION_ENC_KEY=base64_encoded_32_byte_key
COOKIE_SECRET=random_32_char_string

# Application
BASE_URL=http://localhost:3000
API_URL=http://localhost:8080
TIMEZONE=Europe/Moscow

# Optional
LOG_LEVEL=debug
ENABLE_METRICS=false
ENABLE_PPROF=false
```

### 4. Generate Encryption Keys

```bash
# Generate SESSION_ENC_KEY
openssl rand -base64 32

# Generate COOKIE_SECRET
openssl rand -hex 16
```

## Local Development

### Using Docker Compose (Recommended)

Start all services:
```bash
docker compose -f infra/compose/docker-compose.dev.yml up -d
```

This starts:
- PostgreSQL on port 5432
- ClickHouse on port 9000/8123
- Redis on port 6379
- Backend API on port 8080
- Frontend dev server on port 3000

### Manual Setup

1. Start databases:
```bash
docker compose -f infra/compose/docker-compose.dev.yml up -d postgres clickhouse redis
```

2. Run migrations:
```bash
# PostgreSQL
cd backend
go run cmd/migrate/main.go up

# ClickHouse
docker exec -i tg-clickhouse clickhouse-client < migrations/clickhouse/001_initial.sql
```

3. Start backend:
```bash
cd backend
go run cmd/api/main.go
```

4. Start frontend:
```bash
cd web
pnpm dev
```

## Project Structure

### Backend Structure

```
backend/
├── cmd/
│   ├── api/              # API server entrypoint
│   ├── migrate/          # Migration tool
│   └── worker/           # Background worker
├── internal/
│   ├── domain/           # Business entities
│   │   ├── user.go
│   │   ├── message.go
│   │   └── analytics.go
│   ├── services/         # Business logic
│   │   ├── auth/
│   │   ├── telegram/
│   │   ├── ingest/
│   │   └── analytics/
│   ├── adapters/         # External interfaces
│   │   ├── postgres/
│   │   ├── clickhouse/
│   │   ├── mtproto/
│   │   └── cache/
│   └── api/              # HTTP layer
│       ├── handlers/
│       ├── middleware/
│       └── routes.go
└── pkg/                  # Shared packages
    ├── crypto/
    ├── logger/
    └── validator/
```

### Frontend Structure

```
web/
├── app/                  # Next.js App Router
│   ├── (auth)/          # Auth routes
│   ├── (dashboard)/     # Protected routes
│   ├── api/             # API routes
│   └── layout.tsx       # Root layout
├── components/
│   ├── ui/              # shadcn/ui components
│   ├── charts/          # Chart components
│   └── dashboard/       # Dashboard components
├── lib/
│   ├── api/             # API client
│   ├── hooks/           # Custom hooks
│   └── utils/           # Utilities
└── styles/
    └── globals.css      # Global styles
```

## Code Style

### Go

We use `golangci-lint` with the following configuration:

```yaml
linters:
  enable:
    - gofmt
    - goimports
    - govet
    - errcheck
    - staticcheck
    - gosimple
    - ineffassign
    - unused
    - misspell
```

Run linting:
```bash
cd backend
golangci-lint run
```

### TypeScript/React

We use ESLint and Prettier:

```bash
cd web
pnpm lint
pnpm format
```

## Testing

### Backend Tests

Unit tests:
```bash
cd backend
go test ./...
```

Integration tests:
```bash
cd backend
go test -tags=integration ./...
```

Coverage:
```bash
cd backend
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### Frontend Tests

```bash
cd web
pnpm test
pnpm test:e2e
```

## Database Management

### PostgreSQL Migrations

Create a new migration:
```bash
cd backend
go run cmd/migrate/main.go create add_user_preferences
```

Apply migrations:
```bash
go run cmd/migrate/main.go up
```

Rollback:
```bash
go run cmd/migrate/main.go down 1
```

### ClickHouse Migrations

Migrations are in `migrations/clickhouse/` and applied manually:

```bash
docker exec -i tg-clickhouse clickhouse-client < migrations/clickhouse/002_new_migration.sql
```

## API Development

### Adding a New Endpoint

1. Define the handler in `backend/internal/api/handlers/`:
```go
func (h *Handler) GetAnalytics(w http.ResponseWriter, r *http.Request) {
    // Implementation
}
```

2. Add route in `backend/internal/api/routes.go`:
```go
r.Get("/api/analytics/{metric}", h.GetAnalytics)
```

3. Add service method in appropriate service:
```go
func (s *AnalyticsService) GetMetric(ctx context.Context, metric string) (*Result, error) {
    // Implementation
}
```

### Adding a New Analytics Metric

1. Define metric in `backend/internal/domain/analytics.go`:
```go
type Metric interface {
    ID() string
    Execute(ctx context.Context, params Params) (Result, error)
}
```

2. Implement metric in `backend/internal/services/analytics/metrics/`:
```go
type ActivityMetric struct{}

func (m *ActivityMetric) ID() string { return "activity" }

func (m *ActivityMetric) Execute(ctx context.Context, params Params) (Result, error) {
    // Query implementation
}
```

3. Register metric in `backend/internal/services/analytics/registry.go`:
```go
func init() {
    RegisterMetric(&ActivityMetric{})
}
```

## Frontend Development

### Adding a New Dashboard Component

1. Create component in `web/components/dashboard/`:
```tsx
export function ActivityChart({ data }: { data: ActivityData[] }) {
  return (
    <Card>
      <CardHeader>
        <CardTitle>Activity Timeline</CardTitle>
      </CardHeader>
      <CardContent>
        <ResponsiveContainer width="100%" height={300}>
          <LineChart data={data}>
            {/* Chart configuration */}
          </LineChart>
        </ResponsiveContainer>
      </CardContent>
    </Card>
  );
}
```

2. Use in dashboard page:
```tsx
import { ActivityChart } from '@/components/dashboard/activity-chart';

export default function DashboardPage() {
  const { data } = useAnalytics('activity');
  
  return <ActivityChart data={data} />;
}
```

### Theme Customization

Theme variables are in `web/styles/globals.css`:

```css
:root {
  --color-primary: #55CDFC;      /* Trans blue */
  --color-secondary: #F7A8B8;    /* Trans pink */
  --color-neutral-900: #0B0F19;  /* Dark background */
}
```

## Debugging

### Backend Debugging

Enable debug logging:
```bash
LOG_LEVEL=debug go run cmd/api/main.go
```

Enable pprof:
```bash
ENABLE_PPROF=true go run cmd/api/main.go
# Visit http://localhost:8080/debug/pprof
```

### Frontend Debugging

Use React Developer Tools and Redux DevTools (if using state management).

Debug API calls:
```typescript
// In lib/api/client.ts
if (process.env.NODE_ENV === 'development') {
  console.log('API Request:', method, url, data);
}
```

## Performance Profiling

### Backend Profiling

CPU profiling:
```bash
go test -cpuprofile=cpu.prof -bench=.
go tool pprof cpu.prof
```

Memory profiling:
```bash
go test -memprofile=mem.prof -bench=.
go tool pprof mem.prof
```

### ClickHouse Query Analysis

Enable query logging:
```sql
SET log_queries = 1;
```

View slow queries:
```sql
SELECT 
    query,
    query_duration_ms,
    read_rows,
    read_bytes
FROM system.query_log
WHERE type = 'QueryFinish'
ORDER BY query_duration_ms DESC
LIMIT 10;
```

## Troubleshooting

### Common Issues

1. **Telegram API errors**:
   - Check API credentials
   - Verify rate limits
   - Check session encryption

2. **Database connection issues**:
   - Verify connection strings
   - Check firewall rules
   - Ensure migrations are applied

3. **Build failures**:
   - Clear Go module cache: `go clean -modcache`
   - Clear Node modules: `rm -rf node_modules && pnpm install`

### Debug Commands

```bash
# Check database connectivity
psql $DATABASE_URL -c "SELECT 1"

# Check ClickHouse
curl http://localhost:8123/ping

# Test Telegram connection
go run cmd/tools/test-telegram/main.go
```

## Contributing

1. Fork the repository
2. Create a feature branch: `git checkout -b feature/amazing-feature`
3. Commit changes: `git commit -m 'Add amazing feature'`
4. Push to branch: `git push origin feature/amazing-feature`
5. Open a Pull Request

### Commit Message Format

```
type(scope): subject

body

footer
```

Types:
- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation
- `style`: Code style
- `refactor`: Refactoring
- `test`: Tests
- `chore`: Maintenance

Example:
```
feat(analytics): add emoji frequency metric

Implements emoji usage statistics across messages
with support for filtering by date range and chat.

Closes #123