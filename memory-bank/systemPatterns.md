# System Patterns *Optional*

This file documents recurring patterns...

*
## Architectural Patterns

[2025-08-31 14:54:19] - ## Telegram Integration Patterns

### Authentication Flow
- QR code authentication using gotd/td library
- Phone number verification with SMS codes
- 2FA password support
- Session encryption using AES-GCM
- Session persistence in PostgreSQL

### Data Synchronization
- Batch processing for message imports
- Temporary tables for efficient bulk inserts
- Table partitioning by date for scalability
- Progress tracking for long-running operations

### Clean Architecture Integration
- Domain models define business entities
- Service layer handles Telegram client operations
- Repository layer manages data persistence
- Handlers orchestrate service calls

### Security Considerations
- Sessions encrypted before storage
- User-specific encryption contexts
- Secure session restoration
