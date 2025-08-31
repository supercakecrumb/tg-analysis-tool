# Active Context

This file tracks the project's current status...

*
[2025-08-31 13:35:45] - ## Project Setup Analysis

- Project is in initial state with only basic files (README.md, .gitignore)
- .gitignore is configured for Go projects
- Need to create full project structure from scratch
- Monorepo structure planned: backend/, web/, infra/, migrations/

## Architecture Overview

### Backend (Go)
- Clean Architecture pattern
- Chi web framework
- sqlc/pgx for Postgres
- ch-go for ClickHouse
- gotd/td for Telegram MTProto
- Modules: auth/telegram, ingest, pipeline, analytics, sharing, api

### Frontend (Next.js)
- App Router with TypeScript
- TailwindCSS + shadcn/ui
- Recharts for data visualization
- Dark theme with trans-flag accents

### Infrastructure
- Docker containers
- GitHub Actions CI/CD
- Postgres + ClickHouse databases
- Home server deployment
