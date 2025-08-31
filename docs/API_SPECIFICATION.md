# tg-analysis-tool API Specification

## Overview

The tg-analysis-tool API provides RESTful endpoints for Telegram authentication, data synchronization, analytics queries, and multi-tenant management. All endpoints return JSON responses and use standard HTTP status codes.

## Base URL

```
http://localhost:8080/api
```

## Authentication

The API uses session-based authentication with secure httpOnly cookies. CSRF tokens are required for all mutating operations.

### Headers

```
X-CSRF-Token: <csrf-token>
Content-Type: application/json
```

## Endpoints

### Authentication

#### Initialize QR Code Login

```http
POST /api/login/telegram/qr
```

Initializes a new MTProto session and returns a QR code for Telegram login.

**Response:**
```json
{
  "qr_code": "data:image/png;base64,...",
  "session_id": "550e8400-e29b-41d4-a716-446655440000",
  "expires_at": "2024-01-01T12:30:00Z"
}
```

#### Phone Number Login

```http
POST /api/login/telegram/phone
```

Initiates login with phone number.

**Request:**
```json
{
  "phone": "+1234567890"
}
```

**Response:**
```json
{
  "session_id": "550e8400-e29b-41d4-a716-446655440000",
  "code_length": 5,
  "code_type": "sms"
}
```

#### Verify Code

```http
POST /api/login/telegram/verify
```

Verifies the authentication code sent to the phone.

**Request:**
```json
{
  "session_id": "550e8400-e29b-41d4-a716-446655440000",
  "code": "12345"
}
```

**Response:**
```json
{
  "status": "success",
  "requires_2fa": false,
  "user": {
    "id": "user-uuid",
    "phone": "+1234567890",
    "first_name": "John",
    "last_name": "Doe"
  }
}
```

#### Submit 2FA Password

```http
POST /api/login/telegram/password
```

Submits 2FA password if required.

**Request:**
```json
{
  "session_id": "550e8400-e29b-41d4-a716-446655440000",
  "password": "my2fapassword"
}
```

### Data Synchronization

#### Start Sync

```http
POST /api/sync/start
```

Initiates data synchronization for the authenticated user.

**Request:**
```json
{
  "mode": "full", // "full" or "incremental"
  "exclude_chat_ids": [123456, 789012]
}
```

**Response:**
```json
{
  "sync_id": "sync-uuid",
  "status": "started",
  "estimated_messages": 50000
}
```

#### Get Sync Status (SSE)

```http
GET /api/sync/status
```

Server-Sent Events endpoint for real-time sync progress.

**Event Stream:**
```
event: progress
data: {"sync_id":"sync-uuid","processed":1000,"total":50000,"current_chat":"Chat Name"}

event: complete
data: {"sync_id":"sync-uuid","processed":50000,"duration_seconds":120}

event: error
data: {"sync_id":"sync-uuid","error":"Rate limit exceeded","retry_after":60}
```

### Analytics

#### Activity Timeline

```http
GET /api/analytics/activity
```

Returns message activity over time.

**Query Parameters:**
- `start_date` (required): ISO 8601 date
- `end_date` (required): ISO 8601 date
- `granularity`: `hour`, `day`, `week`, `month` (default: `day`)
- `chat_id`: Filter by specific chat
- `chat_type`: `dm`, `group`, `channel`

**Response:**
```json
{
  "series": [
    {
      "timestamp": "2024-01-01T00:00:00Z",
      "count": 150,
      "sent": 75,
      "received": 75
    }
  ],
  "summary": {
    "total_messages": 4500,
    "daily_average": 150
  }
}
```

#### Activity Heatmap

```http
GET /api/analytics/heatmap
```

Returns week × hour activity heatmap.

**Query Parameters:**
- `start_date`: ISO 8601 date
- `end_date`: ISO 8601 date
- `timezone`: IANA timezone (default: `Europe/Moscow`)

**Response:**
```json
{
  "heatmap": [
    {
      "weekday": 1, // Monday
      "hour": 9,
      "count": 250
    }
  ],
  "peak_times": [
    {
      "weekday": 1,
      "hour": 9,
      "description": "Monday 9:00 AM"
    }
  ]
}
```

#### Top Chats

```http
GET /api/analytics/top-chats
```

Returns most active chats by message count.

**Query Parameters:**
- `start_date`: ISO 8601 date
- `end_date`: ISO 8601 date
- `limit`: Number of results (default: 20, max: 100)
- `metric`: `count`, `duration`, `media` (default: `count`)

**Response:**
```json
{
  "chats": [
    {
      "chat_id": "chat-uuid",
      "title": "Family Group",
      "type": "group",
      "message_count": 1500,
      "percentage": 33.3,
      "last_message_at": "2024-01-01T12:00:00Z"
    }
  ]
}
```

#### Message Types

```http
GET /api/analytics/message-types
```

Returns breakdown of message types.

**Query Parameters:**
- `start_date`: ISO 8601 date
- `end_date`: ISO 8601 date
- `chat_id`: Filter by specific chat

**Response:**
```json
{
  "types": [
    {
      "type": "text",
      "count": 3000,
      "percentage": 60.0
    },
    {
      "type": "photo",
      "count": 1000,
      "percentage": 20.0
    },
    {
      "type": "voice",
      "count": 500,
      "percentage": 10.0,
      "total_duration_seconds": 15000
    }
  ],
  "media_stats": {
    "total_voice_duration": "4h 10m",
    "total_video_duration": "2h 30m",
    "average_voice_length": 30
  }
}
```

#### Response Times

```http
GET /api/analytics/response-times
```

Returns response time statistics per peer.

**Query Parameters:**
- `start_date`: ISO 8601 date
- `end_date`: ISO 8601 date
- `peer_id`: Filter by specific peer

**Response:**
```json
{
  "peers": [
    {
      "peer_id": "peer-uuid",
      "name": "John Doe",
      "median_response_seconds": 300,
      "p95_response_seconds": 1800,
      "response_rate": 0.85
    }
  ],
  "overall": {
    "median_response_seconds": 420,
    "average_response_seconds": 600
  }
}
```

#### Top Words

```http
GET /api/analytics/top-words
```

Returns most frequently used words.

**Query Parameters:**
- `start_date`: ISO 8601 date
- `end_date`: ISO 8601 date
- `language`: `en`, `ru` (for stopword filtering)
- `limit`: Number of results (default: 50)
- `min_length`: Minimum word length (default: 3)

**Response:**
```json
{
  "words": [
    {
      "word": "meeting",
      "count": 150,
      "percentage": 2.5
    }
  ],
  "total_words": 6000,
  "unique_words": 1200
}
```

#### Emoji Frequency

```http
GET /api/analytics/emojis
```

Returns emoji usage statistics.

**Response:**
```json
{
  "emojis": [
    {
      "emoji": "😂",
      "count": 500,
      "percentage": 25.0
    }
  ],
  "total_emojis": 2000,
  "messages_with_emojis": 800
}
```

### Settings

#### Update Chat Exclusions

```http
POST /api/settings/exclusions
```

Updates the list of excluded chats.

**Request:**
```json
{
  "excluded_chat_ids": [123456, 789012],
  "action": "replace" // "replace" or "append"
}
```

#### Export Data

```http
GET /api/export
```

Exports user data in requested format.

**Query Parameters:**
- `format`: `json`, `csv` (default: `json`)
- `include`: Comma-separated list of data types

**Response:**
Returns file download with appropriate content-type.

#### Delete Account

```http
DELETE /api/account
```

Permanently deletes user account and all associated data.

**Request:**
```json
{
  "confirmation": "DELETE MY ACCOUNT"
}
```

### Sharing

#### Create Share

```http
POST /api/shares
```

Creates a shareable link for a dashboard.

**Request:**
```json
{
  "resource_type": "dashboard",
  "resource_id": "dashboard-uuid",
  "visibility": "public", // "private", "org", "public"
  "expires_at": "2024-12-31T23:59:59Z"
}
```

**Response:**
```json
{
  "share_id": "share-uuid",
  "token": "signed-token",
  "url": "https://app.example.com/p/signed-token",
  "expires_at": "2024-12-31T23:59:59Z"
}
```

#### List Shares

```http
GET /api/shares
```

Lists all shares created by the user.

**Response:**
```json
{
  "shares": [
    {
      "id": "share-uuid",
      "resource_type": "dashboard",
      "resource_id": "dashboard-uuid",
      "visibility": "public",
      "created_at": "2024-01-01T00:00:00Z",
      "expires_at": "2024-12-31T23:59:59Z",
      "view_count": 42
    }
  ]
}
```

#### Delete Share

```http
DELETE /api/shares/:id
```

Deletes a share link.

### Public Access

#### View Shared Dashboard

```http
GET /p/:token
```

Returns the shared dashboard data (no authentication required).

**Response:**
HTML page with embedded dashboard or JSON data depending on Accept header.

## Error Responses

All endpoints use standard HTTP status codes and return errors in the following format:

```json
{
  "error": {
    "code": "RATE_LIMIT_EXCEEDED",
    "message": "Too many requests. Please try again later.",
    "details": {
      "retry_after": 60
    }
  }
}
```

### Common Error Codes

- `UNAUTHORIZED`: Authentication required
- `FORBIDDEN`: Insufficient permissions
- `NOT_FOUND`: Resource not found
- `VALIDATION_ERROR`: Invalid request parameters
- `RATE_LIMIT_EXCEEDED`: Too many requests
- `TELEGRAM_ERROR`: Telegram API error
- `INTERNAL_ERROR`: Server error

## Rate Limiting

- Authentication endpoints: 5 requests per minute
- Analytics endpoints: 100 requests per minute
- Sync endpoints: 10 requests per hour
- Export endpoints: 5 requests per hour

Rate limit headers:
```
X-RateLimit-Limit: 100
X-RateLimit-Remaining: 95
X-RateLimit-Reset: 1704067200
```

## Pagination

List endpoints support pagination:

**Query Parameters:**
- `page`: Page number (default: 1)
- `per_page`: Items per page (default: 20, max: 100)

**Response Headers:**
```
X-Total-Count: 500
X-Page: 1
X-Per-Page: 20
Link: <http://api/endpoint?page=2>; rel="next"
```

## WebSocket Events (Future)

For real-time updates:

```
ws://localhost:8080/api/ws
```

Events:
- `sync.progress`: Sync progress updates
- `message.new`: New message received
- `analytics.update`: Analytics data updated