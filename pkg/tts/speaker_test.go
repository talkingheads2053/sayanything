package tts

import (
	"testing"
)

func TestRemoveEmoji(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "No emoji in string",
			input:    "Hello, world!",
			expected: "Hello, world!",
		},
		{
			name:     "String with emojis",
			input:    "Hello, world! 😊🌍",
			expected: "Hello, world! ",
		},
		{
			name:     "Only emojis",
			input:    "😊🌍🚀",
			expected: "",
		},
		{
			name:     "Mixed text and emojis",
			input:    "Go is awesome! 🚀🔥",
			expected: "Go is awesome! ",
		},
		{
			name:     "Japanese text",
			input:    "こんにちは、世界。カタカナと漢字です！",
			expected: "こんにちは、世界。カタカナと漢字です！",
		},
		{
			name:     "Japanese text with emojis",
			input:    "GPUで話しています😊🇯🇵",
			expected: "GPUで話しています",
		},
		{
			name:     "Emoji with variation selector",
			input:    "晴れ☀️です",
			expected: "晴れです",
		},
		{
			name:     "Emoji ZWJ sequence",
			input:    "family 👨‍👩‍👧",
			expected: "family ",
		},
		{
			name:     "Chinese and Korean text",
			input:    "你好 안녕하세요",
			expected: "你好 안녕하세요",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := RemoveEmoji(tt.input)
			if result != tt.expected {
				t.Errorf("RemoveEmoji(%q) = %q; want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestRemoveExtraStrings(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		remove   []string
		expected string
	}{
		{
			name:     "No strings to remove",
			input:    "Hello, world!",
			remove:   []string{},
			expected: "Hello, world!",
		},
		{
			name:     "Remove one string",
			input:    "Hello, world!",
			remove:   []string{"world"},
			expected: "Hello, !",
		},
		{
			name:     "Remove multiple strings",
			input:    "Hello, world! Go is awesome!",
			remove:   []string{"world", "awesome"},
			expected: "Hello, ! Go is !",
		},
		{
			name:     "Remove overlapping strings",
			input:    "ababab",
			remove:   []string{"ab"},
			expected: "",
		},
		{
			name:     "Remove strings not present",
			input:    "Hello, world!",
			remove:   []string{"test", "example"},
			expected: "Hello, world!",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := RemoveExtraStrings(tt.input, tt.remove)
			if result != tt.expected {
				t.Errorf("RemoveExtraStrings(%q, %v) = %q; want %q", tt.input, tt.remove, result, tt.expected)
			}
		})
	}
}
