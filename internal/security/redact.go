// Package security centralizes credential redaction for human-readable output.
package security

import "regexp"

const redacted = "<redacted>"

var credentialPatterns = []*regexp.Regexp{
	regexp.MustCompile(`(?i)\b(?:password|secret|community|pre-shared-key|wpa-psk|key-string)\b\s+(.*)$`),
	regexp.MustCompile(`(?i)\bcrypto\s+key\b\s+(.*)$`),
	regexp.MustCompile(`(?i)\bkey\b\s+(.*)$`),
	regexp.MustCompile(`(?i)\bsnmp-server\s+host\s+\S+\s+(?:(?:informs|traps)\s+)?(?:version\s+(?:1|2c|3)\s+)?(\S.*)$`),
	regexp.MustCompile(`(?i)\bsnmp-server\s+user\s+\S+\s+\S+(?:\s+\S+)*?\s+auth\s+(?:md5|sha-1|sha-224|sha-256|sha-384|sha-512|sha)\s+(\S+).*$`),
	regexp.MustCompile(`(?i)\bsnmp-server\s+user\s+\S+\s+\S+(?:\s+\S+)*?\s+priv\s+(?:des|des56|3des|aes)(?:\s+\d+)?\s+(\S+).*$`),
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
