# Implementation Plan: Event Probability Monitor

**Branch**: `001-probability-monitor` | **Date**: 2026-02-16 | **Spec**: [spec.md](spec.md)
**Input**: Feature specification from `/specs/001-probability-monitor/spec.md`

## Summary

Build a lightweight Go service that monitors Polymarket prediction market events for significant probability changes and sends Telegram notifications about the top k events with the most drastic changes. The service automatically polls the Polymarket API, tracks yes/no probability ratios, and uses a simple but effective change detection algorithm with configurable thresholds and time windows. Configuration is managed via a single YAML file, and deployment is via single binary, Docker container, or systemd service.

## Technical Context

**Language/Version**: Go 1.24+ (latest stable)

**Primary Dependencies**:
- **Configuration**: `github.com/spf13/viper` - industry standard for YAML configuration management
- **Telegram Bot**: NEEDS CLARIFICATION - researching best Go Telegram library
- **HTTP Client**: NEEDS CLARIFICATION - evaluating standard library vs enhanced client
- **JSON Parsing**: NEEDS CLARIFICATION - evaluating standard library vs performance libraries

**Storage**: NEEDS CLARIFICATION - in-memory with file-based persistence and rotation, or database?

**Testing**: Go standard `testing` package with table-driven tests

**Target Platform**: Linux server (lightweight VPS), single binary executable

**Project Type**: Single backend service

**Performance Goals**:
- Monitor at least 1000 events concurrently
- Notification delivery within 2 minutes of detection
- 99.5% uptime over 30 days

**Constraints**:
- Lightweight VPS deployment (minimal resource footprint)
- Single-user mode for early version
- Simple, maintainable codebase (no over-engineering)
- Single binary with optional Docker/systemd

**Scale/Scope**:
- Single-user configuration
- 1000+ events monitored
- Polymarket API integration
- Telegram notifications (extensible to other channels)

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

### Principle I: Simplicity and Maintainability
✅ **PASS** - Single-user mode, YAML configuration, simple deployment model, focused core functionality

### Principle II: Go Language
✅ **PASS** - Using Go 1.24+, following idiomatic Go patterns, leveraging standard library where possible

### Principle III: Latest and Robust Dependencies
⚠️ **RESEARCH REQUIRED** - Need to evaluate Telegram bot libraries, HTTP clients, and JSON parsers for well-maintained options

### Principle IV: Comprehensive Unit Testing
✅ **PASS** - Will use Go testing package with table-driven tests, reasonable coverage for critical paths

### Principle V: Code Quality and Taste
✅ **PASS** - Clear naming, explicit error handling, focused functions, avoiding over-engineering

**Gate Status**: Conditional pass - proceeding to Phase 0 research to resolve dependency selections

## Project Structure

### Documentation (this feature)

```text
specs/001-probability-monitor/
├── plan.md              # This file
├── research.md          # Phase 0 output
├── data-model.md        # Phase 1 output
├── quickstart.md        # Phase 1 output
├── contracts/           # Phase 1 output
└── tasks.md             # Phase 2 output (from /speckit.tasks)
```

### Source Code (repository root)

```text
cmd/
└── poly-oracle/
    └── main.go          # Application entry point

internal/
├── config/              # Configuration loading and validation
├── models/              # Domain entities (Event, Snapshot, Change)
├── polymarket/          # Polymarket API client
├── monitor/             # Event monitoring and change detection logic
├── storage/             # Data persistence and rotation
├── telegram/            # Telegram notification client
└── notify/              # Notification orchestration (extensible)

pkg/
└── api/                 # Public API types (if needed)

configs/
└── config.yaml          # Default configuration file

deployments/
├── Dockerfile           # Docker container definition
└── systemd/
    └── poly-oracle.service  # systemd service unit file

tests/
├── integration/         # Integration tests
└── testdata/            # Test fixtures and mock data

Makefile                 # Build automation
go.mod                   # Go module definition
go.sum                   # Dependency checksums
README.md                # Project documentation
```

**Structure Decision**: Single backend service with clear separation of concerns. Using `internal/` for private application code, `cmd/` for entry points, and `pkg/` for any public APIs. This follows Go project layout conventions while keeping the structure simple and maintainable.

## Complexity Tracking

> No violations detected - design adheres to all constitutional principles

| Aspect | Approach | Justification |
|--------|----------|---------------|
| Single-user mode | Simplified architecture | Avoids authentication/authorization complexity for early version |
| File-based config | YAML configuration | Simple deployment, no database required for config |
| In-memory storage | With rotation to disk | Lightweight, suitable for single-user, easy to manage |

---

## Phase 0: Research ✅ COMPLETE

**Status**: Completed

**Artifacts**:
- `research.md` - Technical decisions and dependency selections

**Key Decisions**:
1. **Telegram Bot Library**: `github.com/go-telegram-bot-api/telegram-bot-api` (v5.5.0+)
   - Rationale: Most mature, widely adopted, simple API

2. **HTTP Client**: Standard library `net/http` with custom retry wrapper
   - Rationale: Constitution mandates standard library preference, adequate for polling

3. **JSON Parsing**: Standard library `encoding/json`
   - Rationale: Sufficient performance for scale, no extra dependency

4. **Storage**: In-memory with file-based persistence and rotation
   - Rationale: Lightweight, suitable for single-user, no database overhead

5. **Polymarket API**: Gamma API for events, CLOB API for prices
   - Rationale: Comprehensive data access, well-documented

6. **Change Detection**: Simple threshold-based with time-window filtering
   - Rationale: Effective, explainable, not over-engineered

7. **Configuration**: `github.com/spf13/viper` for YAML
   - Rationale: Industry standard, minimal API

**Total Dependencies**: 2 external (Viper, Telegram Bot API) - meets "minimal dependencies" principle

---

## Phase 1: Design & Contracts ✅ COMPLETE

**Status**: Completed

**Artifacts**:
- `data-model.md` - Entity definitions and relationships
- `contracts/api-contracts.md` - Interface contracts and API integration specs
- `quickstart.md` - Developer onboarding guide

**Data Model Summary**:
- **Event**: Monitored prediction market event from Polymarket
- **Snapshot**: Point-in-time probability reading
- **Change**: Detected significant probability change
- **Config**: User configuration for monitoring behavior

**API Contracts**:
- Polymarket API integration (Gamma + CLOB endpoints)
- Telegram Bot API integration (sendMessage endpoint)
- Internal component interfaces (Config, Storage, Monitor, Notifier)
- Configuration file schema with validation rules

**Quickstart Coverage**:
- 5-minute setup guide
- 3 deployment options (binary, Docker, systemd)
- Configuration reference and tuning guide
- Common issues and troubleshooting
- Development and testing instructions

---

## Constitution Check (Post-Design)

### Principle I: Simplicity and Maintainability
✅ **PASS**
- Clear separation of concerns (monitor, storage, notifier, client)
- Single-user mode eliminates auth complexity
- File-based config, no database required
- Minimal abstractions, straightforward data flow

### Principle II: Go Language
✅ **PASS**
- Go 1.24+ using latest stable version
- Idiomatic Go patterns (interfaces, error handling, goroutines)
- Standard library used preferentially (net/http, encoding/json)
- Project structure follows Go conventions (cmd/, internal/, pkg/)

### Principle III: Latest and Robust Dependencies
✅ **PASS**
- Only 2 external dependencies (Viper, Telegram Bot API)
- Both are well-maintained, widely adopted, stable
- Viper v1.19.0+ (industry standard for config)
- Telegram Bot API v5.5.0+ (most mature Go Telegram library)
- All dependencies use semantic versioning

### Principle IV: Comprehensive Unit Testing
✅ **PLANNED**
- Table-driven tests for all components
- Integration tests with mock HTTP servers
- Contract tests for interfaces
- Test coverage for critical business logic (change detection, storage rotation)
- Benchmark tests for performance monitoring

### Principle V: Code Quality and Taste
✅ **PASS**
- Clear naming conventions (Event, Snapshot, Change, Config)
- Explicit error handling (no ignored errors)
- Small, focused functions (per interface contracts)
- Self-documenting code with clear structure
- No over-engineering (simple threshold algorithm vs ML)

**Gate Status**: ✅ ALL PRINCIPLES SATISFIED - Ready for implementation

---

## Implementation Readiness

### Completed Artifacts
- ✅ `plan.md` - This document
- ✅ `research.md` - Technical research and decisions
- ✅ `data-model.md` - Entity definitions
- ✅ `contracts/api-contracts.md` - Interface contracts
- ✅ `quickstart.md` - Developer guide

### Next Phase
**Phase 2**: Task Generation (`/speckit.tasks`)

The implementation plan is complete. All technical unknowns have been resolved, dependencies selected, and contracts defined. The design adheres to all constitutional principles and is ready for task generation and implementation.
