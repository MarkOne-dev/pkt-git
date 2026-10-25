// Package config holds build metadata and the repository-local names shared by
// the CLI and the Git hooks.
package config

const (
	// GitConfigBinary is the repository-local key holding the absolute path of
	// the installed pkt executable.
	GitConfigBinary = "pktgit.bin"

	// GitConfigLanguage is the git config key holding the user's preferred language.
	GitConfigLanguage = "pkt.language"

	// Legacy keys written by the Python implementation. They are read only to
	// migrate or explain old installations.
	GitConfigLegacyPython = "pktgit.py"
	GitConfigLegacyScript = "pktgit.script"

	// HooksDir is the repository-relative directory containing the thin hooks.
	HooksDir = ".githooks"

	// SnapshotDir is the repository-relative directory containing snapshots.
	SnapshotDir = "snapshot"

	// TopologyDir is the repository-relative directory containing binaries.
	TopologyDir = "topologia"
)
