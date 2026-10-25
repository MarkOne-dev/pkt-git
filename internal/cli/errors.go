package cli

import (
	"errors"
	"fmt"
	"strings"

	"github.com/MarkOne-dev/pkt-git/internal/codec"
	"github.com/MarkOne-dev/pkt-git/internal/hooks"
	"github.com/MarkOne-dev/pkt-git/internal/i18n"
)

var (
	// ErrUsage marks an invalid CLI invocation (exit code 2).
	ErrUsage = errors.New("usage error")

	// ErrRepo marks repository-level errors (missing repo, uninitialized).
	ErrRepo = errors.New("repository error")

	// ErrCodec marks Packet Tracer container decode/encode failures.
	ErrCodec = errors.New("codec error")

	// ErrSnapshot marks snapshot XML or configuration editing errors.
	ErrSnapshot = errors.New("snapshot error")

	// ErrGitOp marks git command failures.
	ErrGitOp = errors.New("git operation error")
)

// usageError carries a user-facing usage message (exit code 2).
type usageError struct {
	message string
}

func (e *usageError) Error() string {
	if e.message == "" {
		return ErrUsage.Error()
	}
	return e.message
}

func (e *usageError) Unwrap() error {
	return ErrUsage
}

func (e *usageError) Is(target error) bool {
	return target == ErrUsage
}

func usagef(format string, args ...any) error {
	return &usageError{message: fmt.Sprintf(format, args...)}
}

// RepoError indicates a repository environment failure.
type RepoError struct {
	Message string
	Hint    string
	Err     error
}

func (e *RepoError) Error() string {
	if e.Message != "" {
		return e.Message
	}
	if e.Err != nil {
		return e.Err.Error()
	}
	return ErrRepo.Error()
}

func (e *RepoError) Unwrap() error {
	if e.Err != nil {
		return e.Err
	}
	return ErrRepo
}

func (e *RepoError) Is(target error) bool {
	return target == ErrRepo
}

// CodecError indicates a Packet Tracer binary container failure.
type CodecError struct {
	Message string
	Path    string
	Hint    string
	Err     error
}

func (e *CodecError) Error() string {
	if e.Message != "" {
		return e.Message
	}
	if e.Err != nil {
		return e.Err.Error()
	}
	return ErrCodec.Error()
}

func (e *CodecError) Unwrap() error {
	if e.Err != nil {
		return e.Err
	}
	return ErrCodec
}

func (e *CodecError) Is(target error) bool {
	return target == ErrCodec
}

// SnapshotError indicates a snapshot topology or XML editing failure.
type SnapshotError struct {
	Message string
	Path    string
	Device  string
	Hint    string
	Err     error
}

func (e *SnapshotError) Error() string {
	if e.Message != "" {
		return e.Message
	}
	if e.Err != nil {
		return e.Err.Error()
	}
	return ErrSnapshot.Error()
}

func (e *SnapshotError) Unwrap() error {
	if e.Err != nil {
		return e.Err
	}
	return ErrSnapshot
}

func (e *SnapshotError) Is(target error) bool {
	return target == ErrSnapshot
}

// GitOpError indicates an error when running Git commands.
type GitOpError struct {
	Message string
	Command string
	Hint    string
	Err     error
}

func (e *GitOpError) Error() string {
	if e.Message != "" {
		return e.Message
	}
	if e.Err != nil {
		return e.Err.Error()
	}
	return ErrGitOp.Error()
}

func (e *GitOpError) Unwrap() error {
	if e.Err != nil {
		return e.Err
	}
	return ErrGitOp
}

func (e *GitOpError) Is(target error) bool {
	return target == ErrGitOp
}

// localizeCLIError maps internal or structured errors to localized messages and remediation hints.
func localizeCLIError(err error) (msg string, hint string) {
	if err == nil {
		return "", ""
	}

	// 1. Structured RepoError
	var repoErr *RepoError
	if errors.As(err, &repoErr) {
		msg = repoErr.Message
		if msg == "" && repoErr.Err != nil {
			msg = repoErr.Err.Error()
		}
		hint = repoErr.Hint
		return msg, hint
	}

	// 2. Structured CodecError
	var codecErr *CodecError
	if errors.As(err, &codecErr) {
		msg = codecErr.Message
		if msg == "" && codecErr.Err != nil {
			msg = codecErr.Err.Error()
		}
		hint = codecErr.Hint
		return msg, hint
	}

	// 3. Structured SnapshotError
	var snapErr *SnapshotError
	if errors.As(err, &snapErr) {
		msg = snapErr.Message
		if msg == "" && snapErr.Err != nil {
			msg = snapErr.Err.Error()
		}
		hint = snapErr.Hint
		return msg, hint
	}

	// 4. Structured GitOpError
	var gitErr *GitOpError
	if errors.As(err, &gitErr) {
		msg = gitErr.Message
		if msg == "" && gitErr.Err != nil {
			msg = gitErr.Err.Error()
		}
		hint = gitErr.Hint
		return msg, hint
	}

	// 5. Sentinel mapping for common known errors
	if errors.Is(err, hooks.ErrNotRepository) || strings.Contains(err.Error(), "not a git repository") {
		return i18n.T("ErrNotGitRepository", nil), i18n.T("ErrNotGitRepositoryHint", nil)
	}
	if errors.Is(err, codec.ErrTruncated) {
		return i18n.T("ErrContainerTruncated", nil), ""
	}
	if errors.Is(err, codec.ErrLengthMismatch) {
		return i18n.T("ErrLengthMismatch", nil), ""
	}
	if errors.Is(err, codec.ErrInvalidXML) || strings.Contains(err.Error(), "invalid XML") {
		return i18n.T("ErrInvalidXML", nil), ""
	}
	if errors.Is(err, hooks.ErrRebuildFailed) {
		return i18n.T("ErrRebuildFailed", nil), ""
	}

	// Fallback to original error message
	return err.Error(), ""
}
