package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/poly-oracle/internal/config"
	"github.com/poly-oracle/internal/logger"
	"github.com/poly-oracle/internal/models"
	"github.com/poly-oracle/internal/monitor"
	"github.com/poly-oracle/internal/polymarket"
	"github.com/poly-oracle/internal/storage"
	"github.com/poly-oracle/internal/telegram"
)

var (
	configPath = flag.String("config", "configs/config.yaml", "Path to configuration file")
)

func main() {
	flag.Parse()

	// Load configuration
	cfg, err := config.Load(*configPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		log.Fatalf("Invalid configuration: %v", err)
	}

	// Setup logging with level support
	logger.Init(cfg.Logging.Level, cfg.Logging.Format)
	logger.Info("Configuration loaded from %s", *configPath)

	// Initialize storage
	store := storage.New(
		cfg.Storage.MaxEvents,
		cfg.Storage.MaxSnapshotsPerEvent,
		cfg.Storage.FilePath,
	)

	// Load persisted data
	if err := store.Load(); err != nil {
		logger.Warn("Failed to load persisted data: %v", err)
	} else {
		logger.Debug("Successfully loaded persisted data from storage")
	}

	// Initialize Polymarket client
	polyClient := polymarket.NewClient(
		cfg.Polymarket.GammaAPIURL,
		cfg.Polymarket.CLOBAPIURL,
		cfg.Polymarket.Timeout,
	)

	// Initialize monitor
	mon := monitor.New(store)

	// Initialize Telegram client
	var telegramClient *telegram.Client
	if cfg.Telegram.Enabled {
		telegramClient, err = telegram.NewClient(cfg.Telegram.BotToken, cfg.Telegram.ChatID)
		if err != nil {
			logger.Fatal("Failed to initialize Telegram client: %v", err)
		}
		logger.Info("Telegram client initialized successfully")
	} else {
		logger.Debug("Telegram notifications disabled")
	}

	// Setup graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		logger.Info("Shutdown signal received, cleaning up...")
		cancel()
	}()

	// Start monitoring loop
	logger.Info("Starting monitoring service (poll: %v, threshold: %.2f, window: %v, top_k: %d)",
		cfg.Polymarket.PollInterval,
		cfg.Monitor.Threshold,
		cfg.Monitor.Window,
		cfg.Monitor.TopK,
	)
	logger.Debug("Monitoring configuration: categories=%v, volume_24hr_min=%.0f, volume_filter_or=%v",
		cfg.Polymarket.Categories,
		cfg.Polymarket.Volume24hrMin,
		cfg.Polymarket.VolumeFilterOR,
	)

	ticker := time.NewTicker(cfg.Polymarket.PollInterval)
	defer ticker.Stop()

	persistenceTicker := time.NewTicker(cfg.Storage.PersistenceInterval)
	defer persistenceTicker.Stop()

	// Run initial poll immediately
	logger.Debug("Running initial monitoring cycle")
	if err := runMonitoringCycle(ctx, polyClient, mon, store, telegramClient, cfg); err != nil {
		logger.Error("Monitoring cycle failed: %v", err)
	}

	for {
		select {
		case <-ctx.Done():
			// Save state before shutdown
			if err := store.Save(); err != nil {
				logger.Error("Failed to save state: %v", err)
			} else {
				logger.Info("State saved successfully before shutdown")
			}
			logger.Info("Service stopped")
			return

		case <-ticker.C:
			logger.Debug("Starting scheduled monitoring cycle")
			if err := runMonitoringCycle(ctx, polyClient, mon, store, telegramClient, cfg); err != nil {
				logger.Error("Monitoring cycle failed: %v", err)
			}

		case <-persistenceTicker.C:
			if err := store.Save(); err != nil {
				logger.Error("Failed to persist data: %v", err)
			} else {
				logger.Debug("Data persisted to disk successfully")
			}

			// Rotate old data
			if err := store.RotateSnapshots(); err != nil {
				logger.Warn("Failed to rotate snapshots: %v", err)
			}
			if err := store.RotateEvents(); err != nil {
				logger.Warn("Failed to rotate events: %v", err)
			}
		}
	}
}

func runMonitoringCycle(
	ctx context.Context,
	polyClient *polymarket.Client,
	mon *monitor.Monitor,
	store *storage.Storage,
	telegramClient *telegram.Client,
	cfg *config.Config,
) error {
	startTime := time.Now()
	logger.Info("Starting monitoring cycle")

	// Fetch events from Polymarket
	logger.Debug("Fetching events from Polymarket API (categories: %v, limit: %d)", cfg.Polymarket.Categories, cfg.Polymarket.Limit)
	events, err := polyClient.FetchEvents(
		ctx,
		cfg.Polymarket.Categories,
		cfg.Polymarket.Volume24hrMin,
		cfg.Polymarket.Volume1wkMin,
		cfg.Polymarket.Volume1moMin,
		cfg.Polymarket.VolumeFilterOR,
		cfg.Polymarket.Limit,
	)
	if err != nil {
		return fmt.Errorf("failed to fetch events: %w", err)
	}
	logger.Info("Fetched %d events from %d categories", len(events), len(cfg.Polymarket.Categories))

	// Update storage with new events and create snapshots
	logger.Debug("Processing fetched events and creating snapshots")
	newEvents := 0
	updatedEvents := 0
	for i := range events {
		event := &events[i]

		// Add or update event
		existingEvent, err := store.GetEvent(event.ID)
		if err != nil {
			// Event doesn't exist, create it
			if err := store.AddEvent(event); err != nil {
				logger.Warn("Failed to add event %s: %v", event.ID, err)
				continue
			}
			newEvents++
		} else {
			// Update existing event
			event.CreatedAt = existingEvent.CreatedAt
			if err := store.UpdateEvent(event); err != nil {
				logger.Warn("Failed to update event %s: %v", event.ID, err)
				continue
			}
			updatedEvents++
		}

		// Create snapshot for current probability
		snapshot := &models.Snapshot{
			ID:             generateID(),
			EventID:        event.ID,
			YesProbability: event.YesProbability,
			NoProbability:  event.NoProbability,
			Timestamp:      time.Now(),
			Source:         "polymarket-gamma-api",
		}

		if err := store.AddSnapshot(snapshot); err != nil {
			logger.Warn("Failed to add snapshot for event %s: %v", event.ID, err)
		}
	}
	logger.Debug("Event processing complete: %d new, %d updated", newEvents, updatedEvents)

	// Detect significant changes
	allEvents, err := store.GetAllEvents()
	if err != nil {
		return fmt.Errorf("failed to get events: %w", err)
	}
	logger.Debug("Detecting changes across %d total events with threshold %.2f", len(allEvents), cfg.Monitor.Threshold)

	changes, err := mon.DetectChanges(convertEvents(allEvents), cfg.Monitor.Threshold, cfg.Monitor.Window)
	if err != nil {
		return fmt.Errorf("failed to detect changes: %w", err)
	}

	// Clear old changes and store new ones
	if err := store.ClearChanges(); err != nil {
		logger.Warn("Failed to clear old changes: %v", err)
	}
	for i := range changes {
		if err := store.AddChange(&changes[i]); err != nil {
			logger.Warn("Failed to add change: %v", err)
		}
	}

	logger.Info("Detected %d significant changes", len(changes))

	// Get top K changes and send notifications
	if len(changes) > 0 && cfg.Telegram.Enabled && telegramClient != nil {
		topChanges := mon.RankChanges(changes, cfg.Monitor.TopK)
		logger.Debug("Ranked changes, sending top %d to Telegram", len(topChanges))

		if err := telegramClient.Send(topChanges); err != nil {
			logger.Error("Failed to send Telegram notification: %v", err)
		} else {
			logger.Info("Sent Telegram notification with top %d changes", len(topChanges))
		}
	} else if len(changes) > 0 {
		logger.Debug("Changes detected but Telegram notifications disabled or client not initialized")
	}

	duration := time.Since(startTime)
	logger.Info("Monitoring cycle completed in %v", duration)

	return nil
}

func generateID() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}

func convertEvents(events []*models.Event) []models.Event {
	result := make([]models.Event, len(events))
	for i, event := range events {
		result[i] = *event
	}
	return result
}
