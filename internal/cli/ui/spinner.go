package ui

import (
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/charmbracelet/lipgloss"
)

var defaultSpinnerFrames = []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}

// Spinner provides animated terminal progress indicators with automatic fallback
// for non-interactive or uncolored environments.
type Spinner struct {
	w           io.Writer
	msg         string
	frames      []string
	interval    time.Duration
	done        chan struct{}
	wg          sync.WaitGroup
	mu          sync.Mutex
	active      bool
	interactive bool
}

// NewSpinner creates a new terminal progress spinner for the given writer and initial message.
func NewSpinner(w io.Writer, msg string) *Spinner {
	return &Spinner{
		w:           w,
		msg:         msg,
		frames:      defaultSpinnerFrames,
		interval:    80 * time.Millisecond,
		done:        make(chan struct{}),
		interactive: ColorEnabled(w),
	}
}

// SetInteractive explicitly sets whether animated TTY escapes are allowed (useful in testing).
func (s *Spinner) SetInteractive(v bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.interactive = v
}

// SetInterval sets the frame delay (useful in testing).
func (s *Spinner) SetInterval(d time.Duration) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.interval = d
}

// Start begins the spinner animation.
func (s *Spinner) Start() {
	s.mu.Lock()
	if s.active {
		s.mu.Unlock()
		return
	}
	s.active = true
	interactive := s.interactive
	s.mu.Unlock()

	if !interactive {
		fmt.Fprintf(s.w, "[..] %s...\n", s.msg)
		return
	}

	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		ticker := time.NewTicker(s.interval)
		defer ticker.Stop()

		r := rendererFor(s.w)
		frameStyle := r.NewStyle().Foreground(lipgloss.Color("39")).Bold(true)
		idx := 0

		for {
			select {
			case <-s.done:
				return
			case <-ticker.C:
				s.mu.Lock()
				msg := s.msg
				s.mu.Unlock()

				frame := s.frames[idx%len(s.frames)]
				idx++
				// \r + clear line (\x1b[2K) + styled frame + message
				fmt.Fprintf(s.w, "\r\x1b[2K%s %s", frameStyle.Render(frame), msg)
			}
		}
	}()
}

// UpdateMessage updates the text displayed next to the spinner.
func (s *Spinner) UpdateMessage(msg string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.msg = msg
}

// Stop cleanly terminates the spinner without printing a completion status.
func (s *Spinner) Stop() {
	s.mu.Lock()
	if !s.active {
		s.mu.Unlock()
		return
	}
	s.active = false
	interactive := s.interactive
	close(s.done)
	s.mu.Unlock()

	if interactive {
		s.wg.Wait()
		fmt.Fprint(s.w, "\r\x1b[2K")
	}
}

// Success stops the spinner and prints a success indicator with the given message.
func (s *Spinner) Success(format string, args ...any) {
	s.Stop()
	msg := fmt.Sprintf(format, args...)
	if s.interactive {
		r := rendererFor(s.w)
		sym := r.NewStyle().Foreground(lipgloss.Color("10")).Bold(true).Render("✔")
		fmt.Fprintf(s.w, "%s %s\n", sym, msg)
	} else {
		fmt.Fprintf(s.w, "[ok] %s\n", msg)
	}
}

// Fail stops the spinner and prints a failure indicator with the given message.
func (s *Spinner) Fail(format string, args ...any) {
	s.Stop()
	msg := fmt.Sprintf(format, args...)
	if s.interactive {
		r := rendererFor(s.w)
		sym := r.NewStyle().Foreground(lipgloss.Color("9")).Bold(true).Render("✖")
		fmt.Fprintf(s.w, "%s %s\n", sym, msg)
	} else {
		fmt.Fprintf(s.w, "pkt: %s\n", msg)
	}
}
