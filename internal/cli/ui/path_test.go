package ui

import (
	"strings"
	"testing"
)

func TestTruncatePath(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		maxLen int
		want   string
	}{
		{
			name:   "short path unchanged",
			input:  "/home/user/project",
			maxLen: 30,
			want:   "/home/user/project",
		},
		{
			name:   "exact length unchanged",
			input:  "/home/user/project",
			maxLen: 18,
			want:   "/home/user/project",
		},
		{
			name:   "long path truncated with boundary",
			input:  "/home/user/deep/nested/subfolder/my-cisco-topology-lab",
			maxLen: 35,
			want:   ".../subfolder/my-cisco-topology-lab",
		},
		{
			name:   "windows path normalized and truncated",
			input:  `C:\Users\runneradmin\AppData\Local\Temp\long-path\repo`,
			maxLen: 30,
			want:   ".../Local/Temp/long-path/repo",
		},
		{
			name:   "very small maxLen",
			input:  "/a/b/c",
			maxLen: 2,
			want:   "/a/b/c",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := TruncatePath(tt.input, tt.maxLen)
			if got != tt.want {
				t.Errorf("TruncatePath(%q, %d) = %q, want %q", tt.input, tt.maxLen, got, tt.want)
			}
			if len(tt.input) > tt.maxLen && tt.maxLen > 3 && !strings.HasPrefix(got, "...") {
				t.Errorf("truncated path missing ... prefix: %q", got)
			}
		})
	}
}
