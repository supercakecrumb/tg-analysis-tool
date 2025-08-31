# Decision Log

This file records architectural and implementation decisions...

*
[2025-08-31 13:36:12] - ## Architecture Decisions

### ADR-001: Monorepo Structure
**Decision**: Use monorepo with backend/, web/, infra/, migrations/ directories
**Rationale**: Simplifies CI/CD, ensures consistency, enables atomic commits across stack

### ADR-002: Database Split (Postgres + ClickHouse)
**Decision**: Use Postgres for transactional data, ClickHouse for analytics
**Rationale**: 
- Postgres: ACID compliance for auth, multi-tenancy, configuration
- ClickHouse: Columnar storage for sub-second analytics on millions of messages

### ADR-003: Clean Architecture for Backend
**Decision**: Implement Clean Architecture with domain/services/adapters layers
**Rationale**: Testability, maintainability, clear separation of concerns

### ADR-004: MTProto via gotd/td
**Decision**: Use gotd/td library for Telegram integration
**Rationale**: Pure Go implementation, no CGO dependencies, actively maintained

### ADR-005: Session Encryption
**Decision**: AES-GCM encryption with HKDF key derivation
**Rationale**: Industry standard, supports key rotation, protects user credentials
