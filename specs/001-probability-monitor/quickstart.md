# Quickstart Guide: Event Probability Monitor

**Date**: 2026-02-16
**Purpose**: Get the service running in 10 minutes

## Prerequisites

- Go 1.24 or later
- Telegram account
- Telegram bot token ([create one with @BotFather](https://t.me/botfather))
- Lightweight VPS or local machine

## Quick Setup (5 minutes)

### 1. Clone and Build

```bash
# Clone repository
git clone <repository-url>
cd poly-oracle

# Install dependencies
go mod download

# Build binary
make build
# Or: go build -o bin/poly-oracle ./cmd/poly-oracle
```

### 2. Create Configuration

```bash
# Copy example config
cp configs/config.yaml.example configs/config.yaml

# Edit configuration
vim configs/config.yaml
```

**Minimal Configuration** (`configs/config.yaml`):

```yaml
polymarket:
  api_base_url: "https://gamma-api.polymarket.com"
  poll_interval: 5m
  categories:
    - politics
    - sports
    - crypto

monitor:
  threshold: 0.10    # 10% change threshold
  window: 1h         # 1 hour time window
  top_k: 10          # Top 10 events
  enabled: true

telegram:
  bot_token: "YOUR_BOT_TOKEN_HERE"  # Get from @BotFather
  chat_id: "YOUR_CHAT_ID_HERE"      # Get from @userinfobot
  enabled: true

storage:
  max_events: 1000
  max_snapshots_per_event: 100
  max_file_size_mb: 100
  persistence_interval: 5m
  file_path: "./data/poly-oracle.json"
  data_dir: "./data"

logging:
  level: "info"
  format: "json"
```

### 3. Get Telegram Credentials

**Bot Token**:
1. Open Telegram and search for `@BotFather`
2. Send `/newbot` command
3. Follow prompts to name your bot
4. Copy the bot token (format: `123456789:ABCdef...`)

**Chat ID**:
1. Search for `@userinfobot` in Telegram
2. Start conversation and it will show your chat ID
3. Copy the numeric chat ID

### 4. Run Service

```bash
# Create data directory
mkdir -p data

# Run with config file
./bin/poly-oracle --config configs/config.yaml

# Or run directly
go run ./cmd/poly-oracle --config configs/config.yaml
```

**Expected Output**:
```
2026-02-16T10:30:00Z info Configuration loaded from configs/config.yaml
2026-02-16T10:30:00Z info Storage initialized with 0 events
2026-02-16T10:30:00Z info Telegram bot connected: @YourBotName
2026-02-16T10:30:00Z info Starting monitoring service (poll: 5m, threshold: 0.10, window: 1h, top_k: 10)
2026-02-16T10:30:00Z info Service started successfully
2026-02-16T10:35:00Z info Polled 150 events from 3 categories
2026-02-16T10:35:00Z info Detected 5 significant changes
2026-02-16T10:35:00Z info Sent Telegram notification with 5 changes
```

## Deployment Options

### Option 1: Direct Binary (Simplest)

```bash
# Build for Linux
GOOS=linux GOARCH=amd64 make build

# Copy to VPS
scp bin/poly-oracle user@vps:/opt/poly-oracle/
scp configs/config.yaml user@vps:/opt/poly-oracle/

# SSH to VPS and run
ssh user@vps
cd /opt/poly-oracle
./poly-oracle --config config.yaml
```

### Option 2: Docker

```bash
# Build Docker image
make docker-build
# Or: docker build -t poly-oracle:latest .

# Run container
docker run -d \
  --name poly-oracle \
  --restart unless-stopped \
  -v $(pwd)/configs:/app/configs \
  -v $(pwd)/data:/app/data \
  poly-oracle:latest

# View logs
docker logs -f poly-oracle
```

### Option 3: Systemd Service

```bash
# Copy systemd unit file
sudo cp deployments/systemd/poly-oracle.service /etc/systemd/system/

# Edit paths in service file
sudo vim /etc/systemd/system/poly-oracle.service

# Enable and start service
sudo systemctl daemon-reload
sudo systemctl enable poly-oracle
sudo systemctl start poly-oracle

# Check status
sudo systemctl status poly-oracle

# View logs
sudo journalctl -u poly-oracle -f
```

## Configuration Reference

### All Configuration Options

| Section | Field | Type | Default | Description |
|---------|-------|------|---------|-------------|
| polymarket | api_base_url | string | https://gamma-api.polymarket.com | Polymarket API base URL |
| polymarket | poll_interval | duration | 5m | How often to poll for updates |
| polymarket | categories | []string | [] | Categories to monitor |
| polymarket | timeout | duration | 30s | API request timeout |
| monitor | threshold | float | 0.10 | Minimum change magnitude |
| monitor | window | duration | 1h | Time window for detection |
| monitor | top_k | int | 10 | Number of top events to notify |
| monitor | enabled | bool | true | Enable monitoring |
| telegram | bot_token | string | "" | Telegram bot token |
| telegram | chat_id | string | "" | Telegram chat ID |
| telegram | enabled | bool | false | Enable Telegram notifications |
| storage | max_events | int | 1000 | Maximum events to track |
| storage | max_snapshots_per_event | int | 100 | Snapshots per event |
| storage | max_file_size_mb | int | 100 | Maximum persistence file size |
| storage | persistence_interval | duration | 5m | How often to save to disk |
| storage | file_path | string | ./data/poly-oracle.json | Persistence file path |
| storage | data_dir | string | ./data | Data directory |
| logging | level | string | info | Log level (debug, info, warn, error) |
| logging | format | string | json | Log format (json, text) |

### Environment Variables

Override any config value with environment variables:

```bash
export POLY_ORACLE_TELEGRAM_BOT_TOKEN="your_token"
export POLY_ORACLE_TELEGRAM_CHAT_ID="your_chat_id"
export POLY_ORACLE_MONITOR_THRESHOLD="0.15"
export POLY_ORACLE_MONITOR_WINDOW="2h"

./poly-oracle --config configs/config.yaml
```

### Tuning for Your Use Case

**More Sensitive Detection**:
```yaml
monitor:
  threshold: 0.05    # 5% change threshold
  window: 30m        # 30 minute window
  top_k: 20          # More events
```

**Less Noise**:
```yaml
monitor:
  threshold: 0.20    # 20% change threshold
  window: 2h         # 2 hour window
  top_k: 5           # Fewer events
```

**Specific Categories Only**:
```yaml
polymarket:
  categories:
    - politics       # Only politics events
```

**Faster Monitoring**:
```yaml
polymarket:
  poll_interval: 1m  # Poll every minute

monitor:
  window: 10m        # 10 minute window
```

## Testing Your Setup

### Send Test Notification

```bash
# Run with debug logging
./poly-oracle --config configs/config.yaml --log-level debug

# Wait for first poll cycle (check logs)
# You should receive a test notification on Telegram
```

### Verify Data Collection

```bash
# Check persistence file
ls -lh data/poly-oracle.json

# View file contents
cat data/poly-oracle.json | jq .

# Check logs for errors
grep -i error logs/poly-oracle.log
```

### Simulate Changes (Development)

```go
// In test mode, you can inject mock events
// See: internal/monitor/monitor_test.go
```

## Common Issues

### Issue: No Notifications Received

**Possible Causes**:
1. Bot token or chat ID incorrect
2. Bot not started in Telegram (send `/start` to your bot)
3. Threshold too high, no changes detected
4. No events in selected categories

**Solution**:
```bash
# Check Telegram connection
curl "https://api.telegram.org/bot<YOUR_TOKEN>/getMe"

# Verify chat ID
curl "https://api.telegram.org/bot<YOUR_TOKEN>/getUpdates"

# Lower threshold temporarily
monitor:
  threshold: 0.01  # 1% threshold
```

### Issue: Polymarket API Errors

**Possible Causes**:
1. Rate limiting (too many requests)
2. Network connectivity issues
3. API endpoint changed

**Solution**:
```yaml
# Increase poll interval
polymarket:
  poll_interval: 10m  # Slower polling

# Add timeout
polymarket:
  timeout: 60s  # Longer timeout
```

### Issue: High Memory Usage

**Possible Causes**:
1. Too many events tracked
2. Too many snapshots stored

**Solution**:
```yaml
# Reduce storage limits
storage:
  max_events: 500
  max_snapshots_per_event: 50
```

## Monitoring and Logs

### Log Locations

- **Console**: Standard output (stdout)
- **Systemd**: `journalctl -u poly-oracle`
- **Docker**: `docker logs poly-oracle`

### Log Levels

- **debug**: All operations (verbose)
- **info**: Normal operations, notifications sent
- **warn**: Recoverable errors, retries
- **error**: Critical errors, failed operations

### Health Check

```bash
# Check if process running
ps aux | grep poly-oracle

# Check systemd status
systemctl status poly-oracle

# Check recent logs
tail -100 logs/poly-oracle.log | grep error
```

## Development

### Run Tests

```bash
# All tests
make test
# Or: go test ./...

# With coverage
make test-coverage
# Or: go test -cover ./...

# Specific package
go test ./internal/monitor -v
```

### Build for Multiple Platforms

```bash
# Linux AMD64
GOOS=linux GOARCH=amd64 go build -o bin/poly-oracle-linux-amd64 ./cmd/poly-oracle

# Linux ARM64
GOOS=linux GOARCH=arm64 go build -o bin/poly-oracle-linux-arm64 ./cmd/poly-oracle

# macOS
GOOS=darwin GOARCH=amd64 go build -o bin/poly-oracle-darwin-amd64 ./cmd/poly-oracle
```

### Local Development

```bash
# Run with auto-reload (using entr or similar)
ls -d internal/**/*.go | entr -r go run ./cmd/poly-oracle --config configs/config.yaml

# Run with debug logging
go run ./cmd/poly-oracle --config configs/config.yaml --log-level debug
```

## Architecture Overview

```
┌─────────────────────────────────────────────────────────┐
│                    Poly Oracle Service                   │
├─────────────────────────────────────────────────────────┤
│                                                          │
│  ┌──────────────┐      ┌──────────────┐                │
│  │   Config     │──────│   Monitor    │                │
│  │   Loader     │      │   Service    │                │
│  └──────────────┘      └──────┬───────┘                │
│                               │                          │
│  ┌──────────────┐             │         ┌─────────────┐│
│  │  Polymarket  │─────────────┤         │  Telegram   ││
│  │   Client     │             │         │   Client    ││
│  └──────────────┘             │         └─────────────┘│
│                               │                          │
│                        ┌──────▼───────┐                  │
│                        │   Storage    │                  │
│                        │  (In-Memory) │                  │
│                        └──────────────┘                  │
│                                                          │
└─────────────────────────────────────────────────────────┘

Data Flow:
1. Config loaded from YAML file
2. Polymarket client polls events every N minutes
3. Storage saves snapshots and calculates changes
4. Monitor detects significant changes
5. Telegram client sends notifications
6. Storage persists to disk periodically
```

## Next Steps

1. **Customize**: Adjust threshold, window, and categories to your needs
2. **Monitor**: Check logs regularly for errors or issues
3. **Tune**: Adjust parameters based on notification frequency
4. **Scale**: Increase max_events if monitoring many events
5. **Extend**: Add custom notification channels (see architecture)

## Getting Help

- **Documentation**: See `specs/001-probability-monitor/` directory
- **Issues**: Report bugs on project issue tracker
- **Logs**: Always include logs when reporting issues

## Useful Commands

```bash
# Check service health
./poly-oracle --config configs/config.yaml --health-check

# Validate configuration
./poly-oracle --config configs/config.yaml --validate-config

# Show version
./poly-oracle --version

# Generate example config
./poly-oracle --generate-config > configs/config.yaml.example
```

## Congratulations! 🎉

Your Event Probability Monitor is now running. You'll receive Telegram notifications when significant probability changes occur in your selected Polymarket categories.

**What's Next**:
- Monitor the logs for the first few hours
- Adjust threshold if you receive too many/few notifications
- Explore different categories and time windows
- Consider setting up monitoring/alerting for the service itself