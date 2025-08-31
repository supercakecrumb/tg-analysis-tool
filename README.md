# tg-analysis-tool

A multi-tenant web application that ingests Telegram chats (with explicit consent via MTProto login), stores them, and serves fast, interactive analytics with graphs, timelines, and breakdowns.

## 🚀 Features

- **Secure Telegram Login**: QR code or phone + 2FA authentication via MTProto
- **Real-time Chat Ingestion**: Initial backfill + hourly incremental sync
- **Interactive Analytics**: Activity timelines, heatmaps, message breakdowns, response times
- **Multi-tenant Architecture**: Organizations with role-based access
- **Public Sharing**: Generate shareable links for dashboards
- **Data Privacy**: Encrypted session storage, data export/delete capabilities
- **High Performance**: Sub-second queries on millions of messages

## 🛠️ Tech Stack

### Backend
- **Language**: Go 1.22+
- **Architecture**: Clean Architecture (domain/services/adapters)
- **Web Framework**: Chi
- **Databases**: PostgreSQL (transactional) + ClickHouse (analytics)
- **Telegram Integration**: gotd/td (MTProto)
- **Authentication**: Session-based with AES-GCM encryption

### Frontend
- **Framework**: Next.js 14+ (App Router)
- **Language**: TypeScript
- **Styling**: TailwindCSS + shadcn/ui
- **Charts**: Recharts
- **Theme**: Dark mode with trans-flag accents 🏳️‍⚧️

### Infrastructure
- **Containerization**: Docker
- **CI/CD**: GitHub Actions → GitHub Container Registry
- **Deployment**: Docker Compose (home server)

## 📁 Project Structure

```
tg-analysis-tool/
├── backend/                 # Go backend application
│   ├── cmd/                # Application entrypoints
│   ├── internal/           # Private application code
│   │   ├── domain/        # Business logic & entities
│   │   ├── services/      # Use cases & orchestration
│   │   ├── adapters/      # External interfaces
│   │   └── api/          # HTTP handlers & middleware
│   └── pkg/               # Public packages
├── web/                    # Next.js frontend
│   ├── app/               # App router pages
│   ├── components/        # React components
│   ├── lib/              # Utilities & hooks
│   └── styles/           # Global styles
├── migrations/            # Database migrations
│   ├── postgres/         # PostgreSQL migrations
│   └── clickhouse/       # ClickHouse migrations
├── infra/                 # Infrastructure configs
│   ├── docker/           # Dockerfiles
│   └── compose/          # Docker Compose files
├── docs/                  # Documentation
└── scripts/              # Build & deployment scripts
```

## 🚦 Getting Started

### Prerequisites

- Docker & Docker Compose
- Go 1.22+ (for local development)
- Node.js 20+ (for local development)
- Telegram API credentials (api_id & api_hash)

### Quick Start

1. Clone the repository:
```bash
git clone https://github.com/supercakecrumb/tg-analysis-tool.git
cd tg-analysis-tool
```

2. Copy environment variables:
```bash
cp .env.example .env
# Edit .env with your configuration
```

3. Start the application:
```bash
docker compose up -d
```

4. Access the application:
- Web UI: http://localhost:3000
- API: http://localhost:8080

### Development Setup

See [docs/DEVELOPMENT.md](docs/DEVELOPMENT.md) for detailed development instructions.

## 📊 Analytics Features

### Available Metrics

- **Activity Timeline**: Message volume over time (hourly/daily)
- **Week × Hour Heatmap**: Activity patterns by day and hour
- **Top Chats/Contacts**: Most active conversations
- **Message Types**: Breakdown by text, media, voice, etc.
- **Response Times**: Median response latency per peer
- **Word Frequency**: Top words with language-aware stopword filtering
- **Emoji Usage**: Most frequently used emojis

### Data Storage

- **Text & Stickers**: Full content stored
- **Voice/Video**: Duration only (no media files)
- **Photos/Files**: Metadata only (type, size)
- **Privacy**: Users can exclude specific chats from ingestion

## 🔒 Security

- **End-to-end Encryption**: MTProto sessions encrypted with AES-GCM
- **Row-Level Security**: PostgreSQL RLS for multi-tenant isolation
- **Session Management**: Secure httpOnly cookies with CSRF protection
- **Data Sovereignty**: Self-hosted, no third-party data sharing
- **Compliance**: Uses official Telegram APIs with rate limiting

## 🚀 Deployment

The application is designed to run on a home server using Docker Compose. GitHub Actions automatically builds and pushes container images to GitHub Container Registry.

See [docs/DEPLOYMENT.md](docs/DEPLOYMENT.md) for deployment instructions.

## 📈 Performance

- **Target**: p95 < 300ms for common queries on 500k+ messages
- **Optimization**: ClickHouse materialized views for hot aggregates
- **Caching**: Query result caching for repeated analytics

## 🤝 Contributing

Contributions are welcome! Please read [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines.

## 📄 License

This project is licensed under the MIT License - see [LICENSE](LICENSE) for details.

## 🙏 Acknowledgments

- [gotd/td](https://github.com/gotd/td) for the excellent MTProto implementation
- [shadcn/ui](https://ui.shadcn.com/) for beautiful UI components
- The Telegram team for their comprehensive API
