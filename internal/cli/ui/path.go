package ui

import (
	"strings"
)

// TruncatePath truncates a file or directory path if its length exceeds maxLen,
// prepending "..." while preserving trailing directory segments.
func TruncatePath(p string, maxLen int) string {
	p = strings.ReplaceAll(p, "\\", "/")
	if len(p) <= maxLen || maxLen <= 3 {
		return p
	}

	targetLen := maxLen - 3
	tail := p[len(p)-targetLen:]

	if slashIdx := strings.Index(tail, "/"); slashIdx != -1 && slashIdx < len(tail)-1 {
		return "..." + tail[slashIdx:]
	}

	return "..." + tail
}
