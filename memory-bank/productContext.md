# Product Context

This file provides a high-level overviewbased on project brief:

# tg-analysis-tool

A multi-tenant web app that ingests a user's own Telegram chats (with explicit consent via MTProto login), stores them, and serves fast, interactive analytics.

## Key Technical Decisions
- **Backend**: Go 1.22+, Clean Architecture
- **Frontend**: Next.js (React) + TypeScript, TailwindCSS, shadcn/ui, Recharts
- **Auth**: Telegram MTProto login (gotd/td)
- **Storage**: Postgres (multi-tenant) + ClickHouse (analytics)
- **Deploy**: Docker on home server via GitHub Actions

## Core Features
1. MTProto login (QR code or phone + 2FA)
2. Chat ingestion with hourly sync
3. Analytics dashboards (timelines, heatmaps, message breakdowns)
4. Multi-tenant with public sharing
5. Data export/delete capabilities

...

*