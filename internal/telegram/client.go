// Package telegram provides a client for sending notifications via Telegram Bot API.
// It formats detected probability changes into human-readable messages and handles
// delivery with retry logic for reliability.
//
// The client supports Markdown formatting and includes error handling for
// common Telegram API issues like rate limiting and network failures.
package telegram

import (
	"fmt"
	"strconv"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/poly-oracle/internal/models"
)

// Client handles Telegram notifications
type Client struct {
	bot    *tgbotapi.BotAPI
	chatID int64
}

// NewClient creates a new Telegram client
func NewClient(botToken, chatID string) (*Client, error) {
	bot, err := tgbotapi.NewBotAPI(botToken)
	if err != nil {
		return nil, fmt.Errorf("failed to create Telegram bot: %w", err)
	}

	chatIDInt, err := strconv.ParseInt(chatID, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid chat ID: %w", err)
	}

	return &Client{
		bot:    bot,
		chatID: chatIDInt,
	}, nil
}

// Send sends a notification with the detected changes
func (c *Client) Send(changes []models.Change) error {
	message := c.formatMessage(changes)

	// Create message
	msg := tgbotapi.NewMessage(c.chatID, message)
	msg.ParseMode = "Markdown"

	// Send with retry
	maxRetries := 3
	var lastErr error

	for i := 0; i < maxRetries; i++ {
		_, err := c.bot.Send(msg)
		if err == nil {
			return nil
		}
		lastErr = err
		time.Sleep(time.Duration(i+1) * time.Second)
	}

	return fmt.Errorf("failed to send message after %d retries: %w", maxRetries, lastErr)
}

// formatMessage formats changes into a Telegram message
func (c *Client) formatMessage(changes []models.Change) string {
	message := "🚨 *Top Probability Changes Detected*\n\n"

	for i, change := range changes {
		// Add emoji for direction
		directionEmoji := "📈"
		if change.Direction == "decrease" {
			directionEmoji = "📉"
		}

		// Format magnitude as percentage
		magnitudePct := change.Magnitude * 100
		oldPct := change.OldProbability * 100
		newPct := change.NewProbability * 100

		message += fmt.Sprintf("%d. %s\n", i+1, change.EventQuestion)
		message += fmt.Sprintf("   %s Change: %.1f%% (%.1f%% → %.1f%%)\n",
			directionEmoji, magnitudePct, oldPct, newPct)
		message += fmt.Sprintf("   ⏱ Window: %s\n", formatDuration(change.TimeWindow))
		message += fmt.Sprintf("   📅 Detected: %s\n\n", change.DetectedAt.Format("2006-01-02 15:04:05"))
	}

	return message
}

// formatDuration formats a duration in a human-readable way
func formatDuration(d time.Duration) string {
	hours := int(d.Hours())
	if hours == 1 {
		return fmt.Sprintf("%dh", hours)
	}
	if hours > 1 {
		return fmt.Sprintf("%dh", hours)
	}

	mins := int(d.Minutes())
	if mins == 1 {
		return fmt.Sprintf("%dm", mins)
	}
	return fmt.Sprintf("%dm", mins)
}
