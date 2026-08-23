package charmcli

import (
	"io"
	"os"

	"github.com/charmbracelet/x/term"
)

// Streams are the process streams used by a CLI invocation.
// They are explicit so commands and tests do not need process-global IO.
type Streams struct {
	In  io.Reader
	Out io.Writer
	Err io.Writer
}

// DefaultStreams returns the process standard streams.
func DefaultStreams() Streams {
	return Streams{In: os.Stdin, Out: os.Stdout, Err: os.Stderr}
}

// Capabilities describe terminal behavior relevant to CLI presentation.
type Capabilities struct {
	Interactive bool
	Accessible  bool
}

// Runtime is the shared execution environment passed to command constructors.
type Runtime struct {
	Streams      Streams
	Capabilities Capabilities
}

type RuntimeOption func(*runtimeSettings)

type runtimeSettings struct {
	capabilities *Capabilities
}

// WithCapabilities overrides terminal capability detection. It is primarily
// useful for tests and remote/non-standard execution environments.
func WithCapabilities(capabilities Capabilities) RuntimeOption {
	return func(s *runtimeSettings) {
		s.capabilities = &capabilities
	}
}

// NewRuntime constructs a runtime from explicit streams.
func NewRuntime(streams Streams, options ...RuntimeOption) *Runtime {
	streams = normalizeStreams(streams)
	settings := runtimeSettings{}
	for _, option := range options {
		option(&settings)
	}

	capabilities := detectCapabilities(streams)
	if settings.capabilities != nil {
		capabilities = *settings.capabilities
	}

	return &Runtime{Streams: streams, Capabilities: capabilities}
}

func normalizeStreams(streams Streams) Streams {
	defaults := DefaultStreams()
	if streams.In == nil {
		streams.In = defaults.In
	}
	if streams.Out == nil {
		streams.Out = defaults.Out
	}
	if streams.Err == nil {
		streams.Err = defaults.Err
	}
	return streams
}

func detectCapabilities(streams Streams) Capabilities {
	interactive := isTerminal(streams.In) && isTerminal(streams.Err)
	accessible := os.Getenv("TERM") == "dumb" || os.Getenv("CHARMCLI_ACCESSIBLE") != ""
	return Capabilities{Interactive: interactive, Accessible: accessible}
}

func isTerminal(value any) bool {
	fd, ok := value.(interface{ Fd() uintptr })
	return ok && term.IsTerminal(fd.Fd())
}
