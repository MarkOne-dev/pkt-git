// Package config holds build metadata and platform paths.
package config

// Build metadata. These are overridden at release time with -ldflags.
var (
	Version = "dev"
	Commit  = "none"
	Date    = "unknown"
)
