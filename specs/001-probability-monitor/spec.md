# Feature Specification: Event Probability Monitor

**Feature Branch**: `001-probability-monitor`
**Created**: 2026-02-16
**Status**: Draft
**Input**: User description: "Build a web service to monitor major event whose probability (win rate) changes drastically over a certain period, and then push notification about top k events to the use using Telegram bot (it may add other notification channels in the future)."

## Clarifications

### Session 2026-02-16

- Q: How should users configure the system? → A: Single YAML configuration file with all parameters
- Q: How should the system obtain probability data from Polymarket? → A: Automatic polling of Polymarket API at configured intervals; data storage must implement rotation and stay under configurable max size; polling interval is configurable; algorithm detects drastic changes in yes/no ratios smartly without over-engineering
- Q: Should the early version support multiple notification channels? → A: Architecture supports multiple channels but only Telegram implemented initially; keep design simple without over-engineering
- Q: Single-user or multi-user system? → A: Single-user mode for early version (one configuration, one set of preferences)
- Q: How should the service be deployed and run? → A: Single binary executable with Docker container support and systemd service configuration; deployment must be simple, elegant, and robust for lightweight servers

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Real-time Probability Change Alerts (Priority: P1)

As a user, I want to receive notifications when event probabilities change significantly so that I can react quickly to important developments.

**Why this priority**: This is the core value proposition of the service - alerting users to significant probability changes in near real-time. Without this, the service has no purpose.

**Independent Test**: Can be fully tested by configuring an event with a probability change exceeding the threshold and verifying that a notification is received via the configured channel within the expected timeframe.

**Acceptance Scenarios**:

1. **Given** an event with probability 60%, **When** the probability changes to 75% within the monitoring window and exceeds the threshold, **Then** a notification is received identifying this event in the top k list
2. **Given** multiple events with probability changes, **When** changes are detected, **Then** only the top k events by change magnitude are included in the notification
3. **Given** a configured Telegram channel, **When** a significant probability change occurs, **Then** the notification is delivered via Telegram within 2 minutes

---

### User Story 2 - Customizable Monitoring Configuration (Priority: P2)

As the user, I want to customize monitoring parameters (threshold, time window, number of events) so that I can focus on the most relevant changes for my needs.

**Why this priority**: Customization is essential for the service to be useful across different use cases and user preferences. However, the core monitoring can work with defaults initially.

**Independent Test**: Can be fully tested by changing configuration parameters and verifying that subsequent monitoring behavior reflects the new settings.

**Acceptance Scenarios**:

1. **Given** configuration file parameters, **When** threshold is set to 15% and monitoring window to 2 hours, **Then** only events with probability changes of at least 15% within 2 hours trigger notifications
2. **Given** k is set to 5, **When** multiple events have significant changes, **Then** notifications contain only the top 5 events ranked by change magnitude
3. **Given** updated configuration file, **When** the next monitoring cycle runs, **Then** the new parameters are applied immediately

---

### User Story 3 - Multiple Notification Channels (Priority: P3)

As a user, I want to add and manage multiple notification channels so that I can receive alerts through my preferred communication platforms.

**Why this priority**: Multi-channel support adds flexibility but is not critical for MVP. The architecture will support future channels but only Telegram is implemented initially.

**Independent Test**: Can be fully tested by adding a new notification channel configuration, triggering a probability change event, and verifying delivery through the new channel.

**Acceptance Scenarios**:

1. **Given** Telegram configured in the system, **When** an additional notification channel is added to configuration (e.g., email), **Then** subsequent notifications are sent to both channels
2. **Given** multiple notification channels configured, **When** one channel fails to deliver, **Then** the system retries delivery and attempts other configured channels
3. **Given** a channel needs to be disabled, **When** the channel is marked inactive in configuration, **Then** notifications cease for that channel but continue through other active channels

---

### User Story 4 - Event Monitoring Management (Priority: P2)

As the user, I want to specify which events to monitor so that I only receive relevant notifications for events I care about.

**Why this priority**: Users need to control their monitoring scope to avoid notification fatigue from irrelevant events. This is important for user experience.

**Independent Test**: Can be fully tested by adding/removing events from the watchlist and verifying that notifications only include watched events.

**Acceptance Scenarios**:

1. **Given** a list of available events from Polymarket, **When** specific events are added to the watchlist in configuration, **Then** only those events are monitored for probability changes
2. **Given** the watchlist, **When** an event is removed from configuration, **Then** future notifications exclude that event
3. **Given** no events in the watchlist, **When** probability changes occur, **Then** no notifications are sent

---

### Edge Cases

- What happens when probability data is unavailable or delayed from the source?
- How does the system handle events with invalid or out-of-range probability values (e.g., negative, > 100%)?
- What happens when a user's notification channel credentials become invalid?
- How does the system behave when the number of monitored events exceeds system capacity?
- What happens when two or more events have identical change magnitudes competing for the kth position?
- How does the system handle monitoring window boundaries (e.g., event occurs at window start vs. end)?
- What happens when a user updates configuration while a notification is being sent?
- How does the system behave when the notification delivery service experiences an outage?

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: System MUST continuously monitor events for probability changes according to configurable monitoring cycles
- **FR-002**: System MUST calculate probability change magnitude as the absolute difference between current and previous probability values
- **FR-003**: System MUST identify probability changes that exceed the user-configured significance threshold within the specified time window
- **FR-004**: System MUST rank events by probability change magnitude and select the top k events for notification
- **FR-005**: System MUST send formatted notifications containing event details, change magnitude, and timestamps via configured notification channels
- **FR-006**: System MUST allow configuration of monitoring parameters (significance threshold, monitoring window duration, k value) via a YAML configuration file for single-user mode
- **FR-007**: System MUST support Telegram as a notification channel using user-provided bot tokens or chat IDs (initial implementation)
- **FR-008**: System architecture MUST allow for future notification channels without requiring major refactoring, but only Telegram is implemented in the early version to avoid over-engineering
- **FR-009**: System MUST allow management of a watchlist of events to monitor for the single user
- **FR-010**: System MUST persist configuration, watchlists, and notification channel settings in the YAML configuration file for the single user
- **FR-011**: System MUST handle notification delivery failures gracefully with retry logic and fallback to alternative channels
- **FR-012**: System MUST provide a way for users to input or connect event data sources for monitoring
- **FR-013**: System MUST store historical probability snapshots to enable change detection over time windows, implementing data rotation to stay under a configurable maximum storage size
- **FR-014**: System MUST automatically poll the Polymarket API at configurable intervals to fetch probability data for monitored events
- **FR-015**: System MUST process probability data from Polymarket prediction market events, tracking yes/no ratios
- **FR-016**: System MUST use a smart but simple algorithm to detect drastic changes in yes/no probability ratios without over-engineering
- **FR-017**: System MUST allow users to filter and select event categories (e.g., politics, sports, crypto, entertainment) for monitoring
- **FR-018**: System MUST be deployable as a single binary executable for simple deployment
- **FR-019**: System MUST include Docker container support with Dockerfile for containerized deployment
- **FR-020**: System MUST include systemd service configuration file for robust daemon operation on Linux servers

### Key Entities

- **Monitored Event**: Represents an event being tracked, including unique identifier, name, current probability, timestamp, and metadata about the event source
- **Probability Snapshot**: Represents a point-in-time record of an event's probability, including event reference, probability value, timestamp, and data source
- **Probability Change**: Represents a detected change between two probability snapshots, including magnitude, direction (increase/decrease), time window, and significance flag
- **User Configuration**: Represents the single user's monitoring preferences, including significance threshold, monitoring window duration, k value (top events limit), active notification channels, and watchlist of monitored events
- **Notification Channel**: Represents a communication endpoint for alerts, including channel type (Telegram, email, webhook), connection credentials, and active status
- **Notification Message**: Represents an alert sent to a user, including recipient, content (top events with details), delivery status, timestamp, and retry count

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Users receive notification of significant probability changes within 2 minutes of the change being detected
- **SC-002**: System correctly identifies and ranks top k events by change magnitude with 100% accuracy based on configured parameters
- **SC-003**: Users can configure monitoring parameters (threshold, window, k value) and see changes take effect within the next monitoring cycle
- **SC-004**: System processes and monitors at least 1000 events concurrently without performance degradation
- **SC-005**: Notification delivery success rate exceeds 95% for configured channels (accounting for temporary outages)
- **SC-006**: Users can add or remove notification channels within 5 minutes without service interruption
- **SC-007**: System maintains 99.5% uptime for monitoring operations during a 30-day period

## Assumptions

- Users configure the system via a single YAML configuration file with reasonable default values
- The system automatically polls Polymarket API for probability data at configurable intervals
- Data storage implements rotation policies to stay under configurable maximum size limits
- Probability values are expressed as percentages or decimal values between 0 and 1
- The system will use a default significance threshold of 10% change if not configured by the user
- The system will use a default monitoring window of 1 hour if not configured by the user
- The system will use a default k value of 10 top events if not configured by the user
- Monitoring cycles occur at configurable intervals (default: every 5 minutes)
- Users manage their own Telegram bot tokens and provide them to the system
- Event identifiers are unique and consistent across probability data updates
- Polymarket events are organized into categories that users can filter and select
- The service runs as a single binary executable on a lightweight VPS
- Docker container and systemd service configurations are provided for robust deployment
