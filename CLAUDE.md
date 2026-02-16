# poly_oracle Development Guidelines

Auto-generated from all feature plans. Last updated: 2026-02-16

## Active Technologies

- Go 1.24+ (latest stable) (001-probability-monitor)

## Project Structure

```text
cmd/poly-oracle/        # Application entry point (main.go)
internal/               # Private application code
  ├── config/          # Configuration loading and validation
  ├── models/          # Domain entities (Event, Snapshot, Change)
  ├── polymarket/      # Polymarket API client (Gamma + CLOB APIs)
  ├── monitor/         # Change detection logic
  ├── storage/         # In-memory storage with file persistence
  └── telegram/        # Telegram bot client
configs/               # YAML configuration files
deployments/           # Dockerfile & systemd service
bin/                   # Built binaries
```

## Commands

```bash
# Development
make install           # Install dependencies (go mod download)
make build             # Build binary to bin/poly-oracle
make test              # Run all tests
make test-coverage     # Run tests with coverage
make run               # Build and run with configs/config.yaml
make fmt               # Format code with gofmt
make lint              # Run golangci-lint

# Deployment
make docker-build      # Build Docker image
make docker-run        # Run Docker container

# Clean
make clean             # Remove binaries and data directory
```

## Dependencies

- **Viper** (github.com/spf13/viper) - YAML configuration management
- **Telegram Bot API** (github.com/go-telegram-bot-api/telegram-bot-api/v5) - Telegram notifications

## Architecture

Single binary service with polling architecture:

1. **Config Loader** → Reads YAML from `configs/config.yaml`
2. **Monitor Service** → Orchestrates polling cycles
3. **Polymarket Client** → Fetches events from Gamma API + CLOB API
4. **Storage** → In-memory with file-based persistence (data rotation)
5. **Change Detection** → Threshold-based algorithm (configurable)
6. **Telegram Client** → Sends notifications for top K changes

Data flow: Poll → Store → Detect Changes → Notify → Persist

## Configuration

Required setup before first run:

1. **Copy example config**: `cp configs/config.yaml.example configs/config.yaml`
2. **Get Telegram credentials**:
   - Bot token: Create bot with [@BotFather](https://t.me/botfather)
   - Chat ID: Get from [@userinfobot](https://t.me/userinfobot)
3. **Edit config**: Add bot_token and chat_id to configs/config.yaml

Config file must have:
- `telegram.bot_token` - Required when telegram.enabled = true
- `telegram.chat_id` - Required when telegram.enabled = true
- `polymarket.categories` - At least one category to monitor

Environment variable overrides supported: `POLY_ORACLE_*`

## Testing

Table-driven tests using standard Go testing package:

```bash
make test              # All tests
make test-coverage     # With coverage
go test ./internal/monitor -v   # Specific package
```

Tests located: `internal/**/*_test.go`

## Deployment Options

1. **Binary**: `make build && ./bin/poly-oracle --config configs/config.yaml`
2. **Docker**: `make docker-build && make docker-run`
3. **systemd**: Copy `deployments/systemd/poly-oracle.service` to `/etc/systemd/system/`

## Development Workflow

```bash
# 1. Install dependencies
make install

# 2. Setup configuration
cp configs/config.yaml.example configs/config.yaml
# Edit configs/config.yaml with Telegram credentials

# 3. Run tests
make test

# 4. Build
make build

# 5. Run locally
make run

# 6. Check logs
# Logs output to stderr (terminal/console)
```

## Key Files

- `cmd/poly-oracle/main.go` - Entry point, orchestration
- `configs/config.yaml` - User configuration (gitignored, copy from example)
- `configs/config.yaml.example` - Template with sensible defaults
- `internal/config/config.go` - Config loading & validation
- `internal/monitor/monitor.go` - Change detection algorithm

## Gotchas

- **Config file required**: Service fails without valid config.yaml
- **Telegram credentials**: Must be set before running
- **Storage path**: Default uses OS tmp dir (/tmp/poly-oracle/data.json)
- **Categories filter**: Only monitors events in configured categories
- **Volume filter**: Lower thresholds needed for niche categories (geopolitics/tech/finance)

## Code Style

Go 1.24+ (latest stable): Follow standard conventions

## Recent Changes

- 001-probability-monitor: Added Go 1.24+ (latest stable)

<!-- MANUAL ADDITIONS START -->
<!-- MANUAL ADDITIONS END -->
