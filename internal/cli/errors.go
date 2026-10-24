package cli

import "errors"

// errUsage marks an invalid invocation so Execute can return exit code 2.
var errUsage = errors.New("usage error")
