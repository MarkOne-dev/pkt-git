// Package security centralizes credential redaction for human-readable output.
package security

import "regexp"

const redacted = "<redacted>"

var credentialPatterns = []*regexp.Regexp{
	regexp.MustCompile(`(?i)\b(?:password|secret|community|pre-shared-key|wpa-psk|key-string)\b\s+(.*)$`),
	regexp.MustCompile(`(?i)\bcrypto\s+key\b\s+(.*)$`),
	regexp.MustCompile(`(?i)\bkey\b\s+(.*)$`),
}

// RedactLine masks the value after a credential keyword. Lines without a
// credential keyword are returned unchanged.
func RedactLine(line string) string {
	for _, pattern := range credentialPatterns {
		value := pattern.FindStringSubmatchIndex(line)
		if value == nil {
			continue
		}
		// Group 1 starts at the value; everything before it is preserved.
		return line[:value[2]] + redacted
	}
	return line
}

// RedactLines applies RedactLine to every line.
func RedactLines(lines []string) []string {
	out := make([]string, len(lines))
	for i, line := range lines {
		out[i] = RedactLine(line)
	}
	return out
}
