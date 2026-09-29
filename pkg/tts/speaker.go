package tts

import (
	"regexp"
	"strings"
)

// Speaker defines the interface for text-to-speech engines. It includes methods for connecting
// to any necessary resources, closing those resources, and converting text to speech data.
type Speaker interface {
	Connect(string) error
	Close()
	Speech(text string) ([]byte, error)
}

// RemoveEmoji removes the typical emoji characters from the input string.
func RemoveEmoji(str string) string {
	// Emoji blocks only, so CJK text is kept. See https://www.unicode.org/charts/ for the block ranges.
	emojiPattern := "[\U0001F100-\U0001F2FF\U0001F300-\U0001F5FF\U0001F600-\U0001F64F\U0001F680-\U0001F6FF\U0001F700-\U0001FAFF\U00002600-\U000026FF\U00002702-\U000027B0\U000024C2\U0000FE0F\U0000200D]+"
	re := regexp.MustCompile(emojiPattern)
	// Replace matched emoji with an empty string to remove it
	return re.ReplaceAllString(str, "")
}

// RemoveExtraStrings removes any of the removal strings from the input string.
func RemoveExtraStrings(str string, remove []string) string {
	result := str
	for _, s := range remove {
		result = strings.ReplaceAll(result, s, "")
	}

	return result
}
