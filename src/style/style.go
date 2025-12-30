package style

import (
	"fmt"
	"log"
	"os"
)

// Style controls how log messages are printed to the screen and optionally to a log file.
type Style struct {
	out    *os.File
	logger *log.Logger
}

// New creates a new Style that prints to stdout and uses the provided log.Logger
// for optional log-file output.
func New(logger *log.Logger) *Style {
	return &Style{
		out:    os.Stdout,
		logger: logger,
	}
}

// ---- Options ----

type options struct {
	bold  bool
	log   bool
	label bool
}

// Option configures how a Style method behaves.
type Option func(*options)

// Bold makes the message bold on the screen.
func Bold() Option {
	return func(o *options) { o.bold = true }
}

// Log makes the message also be written to the log file (via the logger).
func Log() Option {
	return func(o *options) { o.log = true }
}

// Label causes the appropriate label (e.g. [INFO]) to be prepended for
// Info/Warn/Err/Ok methods.
func Label() Option {
	return func(o *options) { o.label = true }
}

// ---- ANSI helpers ----

const (
	ansiReset = "\x1b[0m"
	ansiBold  = "\x1b[1m"

	// 8-color ANSI
	ansiFgCyan   = "\x1b[36m"
	ansiFgYellow = "\x1b[33m"
	ansiFgRed    = "\x1b[31m"
	ansiFgGreen  = "\x1b[32m"

	// 24‑bit RGB
	ansiSubGray = "\x1b[38;2;150;150;150m"
	ansiSign    = "\x1b[38;2;242;103;18m"
)

// core printing helper; NEVER appends newline.
func (s *Style) print(msg, color, defaultLabel string, opts ...Option) {
	if s == nil {
		return
	}

	cfg := options{}
	for _, opt := range opts {
		opt(&cfg)
	}

	text := msg
	if defaultLabel != "" && cfg.label {
		text = defaultLabel + " " + text
	}

	prefix := ""
	suffix := ""

	if color != "" {
		prefix += color
		suffix = ansiReset
	}
	if cfg.bold {
		prefix = ansiBold + prefix
		if suffix == "" {
			suffix = ansiReset
		}
	}

	// Print to screen, no automatic newline.
	fmt.Fprint(s.out, prefix+text+suffix)

	// Optionally write to log file via logger (plain text, no ANSI codes).
	if cfg.log && s.logger != nil {
		s.logger.Print(text)
	}
}

// Plain prints a simple message, optionally bold, optionally logged.
// No color, no label.
func (s *Style) Plain(msg string, opts ...Option) {
	s.print(msg, "", "", opts...)
}

// Sub prints a "sub" message in RGB(150,150,150), optionally bold, optionally logged.
func (s *Style) Sub(msg string, opts ...Option) {
	s.print(msg, ansiSubGray, "", opts...)
}

// Info prints an info message in FgCyan, optionally bold, optionally with "[INFO]",
// and optionally logged.
func (s *Style) Info(msg string, opts ...Option) {
	s.print(msg, ansiFgCyan, "[INFO]", opts...)
}

// Warn prints a warning message in FgYellow, optionally bold, optionally with "[WARN]",
// and optionally logged.
func (s *Style) Warn(msg string, opts ...Option) {
	s.print(msg, ansiFgYellow, "[WARN]", opts...)
}

// Err prints an error message in FgRed, optionally bold, optionally with "[ERR!]",
// and optionally logged.
func (s *Style) Err(msg string, opts ...Option) {
	s.print(msg, ansiFgRed, "[ERR!]", opts...)
}

// Ok prints a success message in FgGreen, optionally bold, optionally with "[OK]",
// and optionally logged.
func (s *Style) Ok(msg string, opts ...Option) {
	s.print(msg, ansiFgGreen, "[OK]", opts...)
}

// Sign prints a signature message in RGB(242,103,18), optionally bold, optionally logged.
// No label.
func (s *Style) Sign(msg string, opts ...Option) {
	s.print(msg, ansiSign, "", opts...)
}
