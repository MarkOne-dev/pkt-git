//go:build !windows

package ptdetect

func queryWindowsRegistry() *Installation {
	return nil
}
