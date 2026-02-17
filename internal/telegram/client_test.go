package telegram

import (
	"testing"
	"time"
)

func TestFormatDuration(t *testing.T) {
	tests := []struct {
		duration time.Duration
		expected string
	}{
		{1 * time.Hour, "1h"},
		{2 * time.Hour, "2h"},
		{30 * time.Minute, "30m"},
		{1 * time.Minute, "1m"},
	}

	for _, tt := range tests {
		result := formatDuration(tt.duration)
		if result != tt.expected {
			t.Errorf("formatDuration(%v) = %s, expected %s", tt.duration, result, tt.expected)
		}
	}
}

func TestEscapeMarkdownV2(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"Hello World", "Hello World"},
		{"Hello_World", "Hello\\_World"},
		{"Test*bold*", "Test\\*bold\\*"},
		{"Price: $100.50", "Price: $100\\.50"},
		{"[link](url)", "\\[link\\]\\(url\\)"},
		{"~strikethrough~", "\\~strikethrough\\~"},
		{"`code`", "\\`code\\`"},
		{">blockquote", "\\>blockquote"},
		{"#header", "\\#header"},
		{"+plus-minus", "\\+plus\\-minus"},
		{"=equal|pipe", "\\=equal\\|pipe"},
		{"{brace}", "\\{brace\\}"},
		{"end!", "end\\!"},
		{"", ""},
		{"_*[]()~`>#+-=|{}.!", "\\_\\*\\[\\]\\(\\)\\~\\`\\>\\#\\+\\-\\=\\|\\{\\}\\.\\!"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := escapeMarkdownV2(tt.input)
			if result != tt.expected {
				t.Errorf("escapeMarkdownV2(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestNewClient_InvalidChatID(t *testing.T) {
	// NewClient with non-numeric chatID should return an error
	// Note: This test exercises the chat ID parsing error path
	// The bot token validation happens first (network call), so we use a clearly
	// invalid format to test the error handling flow
	_, err := NewClient("", "not-a-number", 3, time.Second)
	if err == nil {
		t.Error("Expected error for invalid chat ID, got nil")
	}
}
