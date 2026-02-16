# API Contracts: Event Probability Monitor

**Date**: 2026-02-16
**Purpose**: Define integration contracts and internal interfaces

## Overview

This document defines the API contracts for the Event Probability Monitor service. As a polling service with no external user-facing API, the contracts focus on:
1. Polymarket API integration
2. Telegram Bot API integration
3. Internal component interfaces
4. Configuration file schema

---

## 1. Polymarket API Contract

### Base URL
- **Gamma API**: `https://gamma-api.polymarket.com`
- **CLOB API**: `https://clob.polymarket.com`

### Authentication
- No authentication required for public market data
- Rate limiting applies (conservative: 1 request/second)

### Endpoints

#### GET /events

**Purpose**: Retrieve list of prediction market events

**Request**:
```http
GET /events HTTP/1.1
Host: gamma-api.polymarket.com
Accept: application/json
```

**Query Parameters**:
- `active` (boolean, optional): Filter by active status (default: true)
- `category` (string, optional): Filter by category
- `limit` (integer, optional): Maximum results to return (default: 100)

**Response** (200 OK):
```json
{
  "events": [
    {
      "id": "event-123",
      "question": "Will candidate X win the election?",
      "description": "Detailed description...",
      "category": "politics",
      "active": true,
      "markets": [
        {
          "id": "market-456",
          "outcome": "Yes",
          "outcome_prices": ["0.75"]
        },
        {
          "id": "market-789",
          "outcome": "No",
          "outcome_prices": ["0.25"]
        }
      ],
      "tags": ["politics", "election"]
    }
  ],
  "pagination": {
    "total": 150,
    "limit": 100,
    "offset": 0
  }
}
```

**Error Responses**:
- 429 Too Many Requests: Rate limit exceeded
- 500 Internal Server Error: Polymarket API error

**Contract**:
- Service MUST handle rate limit errors with exponential backoff
- Service MUST validate response structure before processing
- Service MUST log API errors for debugging
- Service SHOULD retry failed requests up to 3 times

---

#### GET /markets/{market_id}

**Purpose**: Get detailed market information including current price

**Request**:
```http
GET /markets/market-456 HTTP/1.1
Host: clob.polymarket.com
Accept: application/json
```

**Response** (200 OK):
```json
{
  "id": "market-456",
  "event_id": "event-123",
  "outcome": "Yes",
  "outcome_prices": ["0.75"],
  "active": true,
  "volume": "1500000.00",
  "liquidity": "500000.00"
}
```

**Contract**:
- Price field represents probability (0.0-1.0)
- Service MUST handle missing or invalid price data gracefully
- Service MUST NOT assume field presence (validate all fields)

---

## 2. Telegram Bot API Contract

### Base URL
`https://api.telegram.org/bot<token>`

### Authentication
- Bot token provided via configuration
- Token format: `123456789:ABCdefGHIjklMNOpqrsTUVwxyz`

### Endpoints

#### POST /sendMessage

**Purpose**: Send notification message to Telegram chat

**Request**:
```http
POST /sendMessage HTTP/1.1
Host: api.telegram.org
Content-Type: application/json

{
  "chat_id": "123456789",
  "text": "🚨 Significant probability changes detected:\n\n1. Event: Will X happen?\n   Change: +15% (60% → 75%)\n   Window: 1h",
  "parse_mode": "Markdown"
}
```

**Request Fields**:
- `chat_id` (string, required): Target chat ID
- `text` (string, required): Message text (max 4096 chars)
- `parse_mode` (string, optional): "Markdown" or "HTML" for formatting
- `disable_notification` (boolean, optional): Send silently (default: false)

**Response** (200 OK):
```json
{
  "ok": true,
  "result": {
    "message_id": 123,
    "from": {
      "id": 987654321,
      "is_bot": true,
      "first_name": "Poly Oracle Bot"
    },
    "chat": {
      "id": 123456789,
      "type": "private"
    },
    "date": 1708084800,
    "text": "Message text..."
  }
}
```

**Error Responses**:
- 400 Bad Request: Invalid parameters (invalid chat_id, token, etc.)
- 401 Unauthorized: Invalid bot token
- 429 Too Many Requests: Rate limit exceeded

**Contract**:
- Service MUST handle and log all error responses
- Service MUST validate bot token on startup
- Service MUST format messages within 4096 character limit
- Service SHOULD use Markdown formatting for better readability
- Service MUST implement retry logic with backoff on rate limits

**Message Format Template**:
```
🚨 Top {k} Probability Changes Detected

{rank}. {event_question}
   📊 Change: {direction} {magnitude}% ({old}% → {new}%)
   ⏱ Window: {window}
   📅 Detected: {timestamp}

---
Configured threshold: {threshold}%
Monitoring window: {window}
```

---

## 3. Internal Component Interfaces

### Config Interface

```go
type ConfigLoader interface {
    Load(path string) (*Config, error)
    Validate() error
    GetPolymarketConfig() PolymarketConfig
    GetMonitorConfig() MonitorConfig
    GetTelegramConfig() TelegramConfig
    GetStorageConfig() StorageConfig
}
```

**Contract**:
- Load MUST return error if file doesn't exist or is malformed
- Validate MUST check all required fields and constraints
- Defaults MUST be applied for optional fields

---

### Storage Interface

```go
type Storage interface {
    // Event operations
    AddEvent(event *Event) error
    GetEvent(id string) (*Event, error)
    GetAllEvents() ([]*Event, error)
    UpdateEvent(event *Event) error

    // Snapshot operations
    AddSnapshot(snapshot *Snapshot) error
    GetSnapshots(eventID string) ([]Snapshot, error)
    GetSnapshotsInWindow(eventID string, window time.Duration) ([]Snapshot, error)

    // Change operations
    AddChange(change *Change) error
    GetTopChanges(k int) ([]Change, error)
    ClearChanges() error

    // Persistence
    Save() error
    Load() error

    // Rotation
    RotateSnapshots() error
    RotateEvents() error
}
```

**Contract**:
- All operations MUST be thread-safe
- GetSnapshotsInWindow MUST filter by timestamp within window
- GetTopChanges MUST return changes sorted by magnitude descending
- Save MUST be atomic (write to temp file, then rename)
- Load MUST handle corrupted file gracefully (log error, return empty state)
- RotateSnapshots MUST respect max_snapshots_per_event config
- RotateEvents MUST respect max_events config

---

### Polymarket Client Interface

```go
type PolymarketClient interface {
    FetchEvents(categories []string) ([]Event, error)
    FetchMarketData(eventID string) ([]Market, error)
    SetPollInterval(interval time.Duration)
    Start() error
    Stop() error
}
```

**Contract**:
- FetchEvents MUST return only events matching configured categories
- FetchMarketData MUST include current probability data
- SetPollInterval MUST apply immediately to next poll cycle
- Start MUST begin background polling goroutine
- Stop MUST gracefully stop polling (complete current poll)
- All fetch methods MUST retry on transient errors (max 3 retries)

---

### Monitor Interface

```go
type Monitor interface {
    DetectChanges(events []Event, threshold float64, window time.Duration) ([]Change, error)
    RankChanges(changes []Change, k int) []Change
    Start() error
    Stop() error
}
```

**Contract**:
- DetectChanges MUST use algorithm from research.md
- DetectChanges MUST return only changes exceeding threshold
- RankChanges MUST sort by magnitude descending
- RankChanges MUST return exactly k changes (or fewer if insufficient)
- Start MUST begin monitoring cycle loop
- Stop MUST complete current detection cycle before stopping

---

### Notifier Interface

```go
type Notifier interface {
    Send(changes []Change) error
    SetEnabled(enabled bool)
    IsEnabled() bool
}
```

**Contract**:
- Send MUST format message according to template
- Send MUST handle Telegram API errors gracefully
- Send MUST log delivery success/failure
- SetEnabled MUST apply immediately (no queued messages)

---

## 4. Configuration File Schema

### YAML Schema

```yaml
# Polymarket API configuration
polymarket:
  api_base_url: "https://gamma-api.polymarket.com"  # Required, valid URL
  poll_interval: 5m                                  # Required, >= 1m
  categories:                                        # Required, >= 1 category
    - politics
    - sports
    - crypto
    - entertainment
  timeout: 30s                                       # Optional, default 30s

# Monitoring behavior configuration
monitor:
  threshold: 0.10        # Required, 0.0-1.0, significance threshold
  window: 1h             # Required, >= 1m, time window for detection
  top_k: 10              # Required, >= 1, number of events to notify
  enabled: true          # Optional, default true

# Telegram notification configuration
telegram:
  bot_token: "YOUR_BOT_TOKEN"  # Required if enabled, non-empty
  chat_id: "YOUR_CHAT_ID"      # Required if enabled, non-empty
  enabled: true                # Required

# Storage and persistence configuration
storage:
  max_events: 1000               # Required, >= 1
  max_snapshots_per_event: 100   # Required, >= 10
  max_file_size_mb: 100          # Required, >= 1
  persistence_interval: 5m       # Required, >= 1m
  file_path: "./data/poly-oracle.json"  # Required, valid path
  data_dir: "./data"             # Optional, default "./data"

# Logging configuration
logging:
  level: "info"    # Required, one of: debug, info, warn, error
  format: "json"   # Required, one of: json, text
```

### Validation Rules

| Field | Type | Required | Validation |
|-------|------|----------|------------|
| polymarket.api_base_url | string | Yes | Valid URL |
| polymarket.poll_interval | duration | Yes | >= 1m |
| polymarket.categories | []string | Yes | >= 1 item |
| polymarket.timeout | duration | No | Default: 30s |
| monitor.threshold | float | Yes | 0.0-1.0 |
| monitor.window | duration | Yes | >= 1m |
| monitor.top_k | int | Yes | >= 1 |
| monitor.enabled | bool | No | Default: true |
| telegram.bot_token | string | Conditional | Non-empty if enabled |
| telegram.chat_id | string | Conditional | Non-empty if enabled |
| telegram.enabled | bool | Yes | - |
| storage.max_events | int | Yes | >= 1 |
| storage.max_snapshots_per_event | int | Yes | >= 10 |
| storage.max_file_size_mb | int | Yes | >= 1 |
| storage.persistence_interval | duration | Yes | >= 1m |
| storage.file_path | string | Yes | Valid file path |
| storage.data_dir | string | No | Default: "./data" |
| logging.level | string | Yes | One of: debug, info, warn, error |
| logging.format | string | Yes | One of: json, text |

### Environment Variable Overrides

All configuration values can be overridden with environment variables:
- Format: `POLY_ORACLE_<SECTION>_<FIELD>` (uppercase, underscores)
- Example: `POLY_ORACLE_TELEGRAM_BOT_TOKEN=your_token`
- Example: `POLY_ORACLE_MONITOR_THRESHOLD=0.15`

---

## 5. Error Handling Contract

### Error Categories

1. **Configuration Errors**: Invalid config, missing required fields
   - Action: Log error, exit with code 1
   - Message: Clear description of what's wrong

2. **API Errors**: Polymarket/Telegram API failures
   - Action: Log error, retry with backoff
   - After max retries: Continue monitoring cycle, skip failed operation

3. **Storage Errors**: Persistence failures, disk full, corruption
   - Action: Log error, continue with in-memory data
   - Next cycle: Attempt persistence again

4. **Internal Errors**: Unexpected panics, assertion failures
   - Action: Log with stack trace, recover if possible
   - Continue service operation if safe

### Logging Contract

All errors MUST be logged with:
- Timestamp (ISO 8601)
- Error level (debug, info, warn, error)
- Component name (config, storage, polymarket, telegram, monitor)
- Error message
- Context data (event_id, api_endpoint, etc.)
- Stack trace (for errors and panics only)

Example:
```json
{
  "timestamp": "2026-02-16T10:30:00Z",
  "level": "error",
  "component": "telegram",
  "message": "Failed to send notification",
  "error": "API error 429: Too Many Requests",
  "context": {
    "chat_id": "123456789",
    "message_length": 512,
    "retry_count": 3
  },
  "stack_trace": "..."
}
```

---

## 6. Testing Contract

### Unit Test Requirements

Each component MUST have unit tests covering:
- Happy path scenarios
- Error handling scenarios
- Edge cases (empty data, invalid input, boundary conditions)

### Integration Test Requirements

Integration tests MUST cover:
- Polymarket API client with mock HTTP server
- Telegram client with mock API
- Storage persistence and rotation
- End-to-end monitoring cycle

### Contract Tests

For each interface, tests MUST verify:
- All methods are implemented
- Error conditions return errors (not panic)
- Thread-safety (concurrent access)
- Contract compliance (return types, behaviors)

---

## Next Steps

1. Implement interfaces with contract-compliant behavior
2. Write contract tests for each interface
3. Create mock implementations for testing
4. Document any deviations from contracts with justification