package util

import (
	"fmt"
	"strings"
)

// Fractional Indexing implementation
// Based on https://www.figma.com/blog/realtime-editing-of-ordered-sequences/
// Generates lexicographically sortable position strings for efficient ordering

const (
	// Base62 characters for position encoding (0-9, A-Z, a-z)
	base62Chars = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
	midChar     = "V" // Middle character in base62 (roughly)
)

// GenerateInitialPosition generates the first position string
func GenerateInitialPosition() string {
	return "a0"
}

// GeneratePositionBetween generates a new position between two existing positions
// If before is empty, generates position before after
// If after is empty, generates position after before
// If both are empty, generates initial position
func GeneratePositionBetween(before, after string) string {
	if before == "" && after == "" {
		return GenerateInitialPosition()
	}

	if before == "" {
		// Insert before 'after'
		return generateBefore(after)
	}

	if after == "" {
		// Insert after 'before'
		return generateAfter(before)
	}

	// Insert between 'before' and 'after'
	return generateBetween(before, after)
}

// generateBefore creates a position before the given position
func generateBefore(pos string) string {
	if pos == "" || pos == "a0" {
		return "a"
	}

	// Decrement the last character
	runes := []rune(pos)
	lastIdx := len(runes) - 1
	lastChar := runes[lastIdx]

	if lastChar > '0' {
		// Can decrement
		runes[lastIdx] = lastChar - 1
		return string(runes)
	}

	// Last char is '0', need to add a suffix
	return pos[:len(pos)-1]
}

// generateAfter creates a position after the given position
func generateAfter(pos string) string {
	if pos == "" {
		return "a0"
	}

	// Try incrementing last character
	runes := []rune(pos)
	lastIdx := len(runes) - 1
	lastChar := runes[lastIdx]

	// Check if we can increment
	charIdx := strings.IndexRune(base62Chars, lastChar)
	if charIdx >= 0 && charIdx < len(base62Chars)-1 {
		runes[lastIdx] = rune(base62Chars[charIdx+1])
		return string(runes)
	}

	// Last char is 'z', append midpoint
	return pos + midChar
}

// generateBetween creates a position between two positions
func generateBetween(before, after string) string {
	// Ensure before < after
	if before >= after {
		panic(fmt.Sprintf("Invalid order: before (%s) >= after (%s)", before, after))
	}

	// Find common prefix
	commonLen := 0
	minLen := len(before)
	if len(after) < minLen {
		minLen = len(after)
	}

	for i := 0; i < minLen; i++ {
		if before[i] != after[i] {
			break
		}
		commonLen++
	}

	commonPrefix := before[:commonLen]

	// Get the differing parts
	beforeSuffix := ""
	afterSuffix := ""

	if commonLen < len(before) {
		beforeSuffix = before[commonLen:]
	}
	if commonLen < len(after) {
		afterSuffix = after[commonLen:]
	}

	// If after is longer and before is a prefix of after
	if beforeSuffix == "" && afterSuffix != "" {
		// before = "a", after = "a1"
		// result = "a" + midChar = "aV"
		return commonPrefix + midChar
	}

	// If before is longer
	if beforeSuffix != "" && afterSuffix == "" {
		// before = "a1", after = "a"
		// This shouldn't happen if before < after
		return generateAfter(before)
	}

	// Both have suffixes
	if beforeSuffix != "" && afterSuffix != "" {
		beforeChar := rune(beforeSuffix[0])
		afterChar := rune(afterSuffix[0])

		// If characters are adjacent
		if afterChar-beforeChar == 1 {
			// Need to go deeper
			// before = "a0", after = "a1"
			// result = "a0" + midChar = "a0V"
			return before + midChar
		}

		// Characters have gap
		// before = "a0", after = "a5"
		// Find midpoint
		midpoint := (beforeChar + afterChar) / 2
		return commonPrefix + string(midpoint)
	}

	// Should not reach here
	return commonPrefix + midChar
}

// ValidatePosition checks if a position string is valid
func ValidatePosition(pos string) bool {
	if pos == "" {
		return false
	}

	for _, r := range pos {
		if !strings.ContainsRune(base62Chars, r) {
			return false
		}
	}

	return true
}

// ComparePositions compares two position strings lexicographically
// Returns: -1 if a < b, 0 if a == b, 1 if a > b
func ComparePositions(a, b string) int {
	if a < b {
		return -1
	}
	if a > b {
		return 1
	}
	return 0
}
