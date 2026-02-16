# Poly Oracle - Event Probability Monitor

A lightweight Go service that monitors Polymarket prediction market events for significant probability changes and sends Telegram notifications about the top k events with the most drastic changes.

## Features

- **Automatic Monitoring**: Polls Polymarket API for probability updates at configurable intervals
- **Smart Detection**: Detects significant probability changes using threshold-based algorithm
- **Telegram Notifications**: Sends formatted alerts to your Telegram account
- **Configurable**: Flexible YAML configuration with reasonable defaults
- **Lightweight**: Designed for deployment on minimal VPS infrastructure
- **Single Binary**: Easy deployment - just copy and run
- **Docker & systemd**: Supports containerized and daemon deployment

## Quick Start

### Prerequisites

- Go 1.24 or later
- Telegram account
- Telegram bot token ([create one with @BotFather](https://t.me/botfather))

### Installation

```bash
# Clone repository
git clone <repository-url>
cd poly-oracle

# Install dependencies
make install

# Create configuration
cp configs/config.yaml.example configs/config.yaml

# Edit configuration with your Telegram credentials
vim configs/config.yaml

# Build and run
make run
```

### Configuration

Edit `configs/config.yaml`:

```yaml
telegram:
  bot_token: "YOUR_BOT_TOKEN"  # From @BotFather
  chat_id: "YOUR_CHAT_ID"      # From @userinfobot
  enabled: true

monitor:
  threshold: 0.10  # 10% change threshold
  window: 1h       # 1 hour time window
  top_k: 10        # Top 10 events to report

polymarket:
  categories:
    - politics
    - sports
    - crypto
```

## Deployment

### Option 1: Binary

```bash
make build
./bin/poly-oracle --config configs/config.yaml
```

### Option 2: Docker

```bash
make docker-build
docker run -d \
  --name poly-oracle \
  -v $(PWD)/configs:/app/configs \
  -v $(PWD)/data:/app/data \
  poly-oracle:latest
```

### Option 3: systemd

```bash
sudo cp deployments/systemd/poly-oracle.service /etc/systemd/system/
sudo systemctl enable poly-oracle
sudo systemctl start poly-oracle
```

## Usage

1. Configure your monitoring parameters in `config.yaml`
2. Start the service
3. Receive Telegram notifications when significant probability changes occur

### Example Notification

```
🚨 Top 10 Probability Changes Detected

1. Will candidate X win the election?
   📊 Change: +15% (60% → 75%)
   ⏱ Window: 1h
   📅 Detected: 2026-02-16 10:30:00

---
Configured threshold: 10%
Monitoring window: 1h
```

## Development

### Run Tests

```bash
make test
```

### Build for Multiple Platforms

```bash
# Linux AMD64
GOOS=linux GOARCH=amd64 make build

# Linux ARM64
GOOS=linux GOARCH=arm64 make build

# macOS
GOOS=darwin GOARCH=amd64 make build
```

### Project Structure

```
cmd/poly-oracle/          # Application entry point
internal/                 # Private application code
  ├── config/            # Configuration management
  ├── models/            # Domain entities
  ├── polymarket/        # Polymarket API client
  ├── monitor/           # Change detection logic
  ├── storage/           # Data persistence
  ├── telegram/          # Telegram client
  └── notify/            # Notification orchestration
configs/                 # Configuration files
deployments/             # Docker & systemd files
```

## Architecture

```
┌─────────────────────────────────────────────────────────┐
│                    Poly Oracle Service                   │
├─────────────────────────────────────────────────────────┤
│  Config Loader → Monitor Service → Telegram Client      │
│                     ↓                                    │
│              Polymarket Client                           │
│                     ↓                                    │
│                  Storage                                 │
└─────────────────────────────────────────────────────────┘

Data Flow:
1. Load configuration from YAML
2. Poll Polymarket API every N minutes
3. Store probability snapshots in memory
4. Detect significant changes using threshold algorithm
5. Send Telegram notifications for top k changes
6. Persist state to disk periodically
```

## Configuration Reference

| Section | Field | Type | Default | Description |
|---------|-------|------|---------|-------------|
| polymarket | poll_interval | duration | 5m | How often to poll for updates |
| polymarket | categories | []string | [] | Categories to monitor |
| monitor | threshold | float | 0.10 | Minimum change magnitude |
| monitor | window | duration | 1h | Time window for detection |
| monitor | top_k | int | 10 | Number of events to notify |
| telegram | bot_token | string | "" | Telegram bot token |
| telegram | chat_id | string | "" | Telegram chat ID |
| storage | max_events | int | 1000 | Maximum events to track |
| storage | max_snapshots_per_event | int | 100 | Snapshots per event |

## Contributing

Contributions welcome! Please ensure all code:
- Follows Go best practices
- Includes unit tests
- Passes `golangci-lint`
- Maintains constitutional principles (simplicity, testability, robustness)

## License

MIT

## Support

- **Documentation**: See `specs/001-probability-monitor/` directory
- **Issues**: Report bugs on project issue tracker

## Acknowledgments

Built with:
- [Viper](https://github.com/spf13/viper) - Configuration management
- [Telegram Bot API](https://github.com/go-telegram-bot-api/telegram-bot-api) - Telegram integration
- [Polymarket API](https://docs.polymarket.com/) - Market data
