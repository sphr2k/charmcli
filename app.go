package charmcli

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"

	"charm.land/fang/v2"
	"github.com/spf13/cobra"
)

// RootFactory constructs a Cobra command tree for one invocation.
type RootFactory func(*Runtime) *cobra.Command

// ErrorTransform sanitizes or otherwise changes an error for presentation.
// Exit classification always uses the original error.
type ErrorTransform func(error) error

// App owns process-level CLI execution around an ordinary Cobra command tree.
type App struct {
	fangOptions    []fang.Option
	runtimeOptions []RuntimeOption
	errorTransform ErrorTransform
}

type Option func(*App)

// New creates an application with Charm/Fang defaults.
func New(options ...Option) *App {
	app := &App{
		fangOptions: []fang.Option{fang.WithNotifySignal(os.Interrupt)},
	}
	for _, option := range options {
		option(app)
	}
	return app
}

// WithVersion configures Fang's version output.
func WithVersion(version string) Option {
	return func(app *App) {
		app.fangOptions = append(app.fangOptions, fang.WithVersion(version))
	}
}

// WithCommit configures Fang's commit metadata.
func WithCommit(commit string) Option {
	return func(app *App) {
		app.fangOptions = append(app.fangOptions, fang.WithCommit(commit))
	}
}

// WithFangOptions passes native Fang configuration through without mirroring
// Fang's API in charmcli.
func WithFangOptions(options ...fang.Option) Option {
	return func(app *App) {
		app.fangOptions = append(app.fangOptions, options...)
	}
}

// WithRuntimeOptions configures Runtime construction.
func WithRuntimeOptions(options ...RuntimeOption) Option {
	return func(app *App) {
		app.runtimeOptions = append(app.runtimeOptions, options...)
	}
}

// WithErrorTransform installs a presentation-only error transformation hook.
func WithErrorTransform(transform ErrorTransform) Option {
	return func(app *App) {
		app.errorTransform = transform
	}
}

// Run constructs the runtime and root command, executes it through Fang and
// returns the process exit code. args are command arguments without argv[0].
func (app *App) Run(ctx context.Context, args []string, streams Streams, factory RootFactory) int {
	runtime := NewRuntime(streams, app.runtimeOptions...)
	if factory == nil {
		_, _ = fmt.Fprintln(runtime.Streams.Err, "ERROR: nil command factory")
		return 1
	}
	root := factory(runtime)
	if root == nil {
		_, _ = fmt.Fprintln(runtime.Streams.Err, "ERROR: nil root command")
		return 1
	}

	root.SetIn(runtime.Streams.In)
	root.SetOut(runtime.Streams.Out)
	root.SetErr(runtime.Streams.Err)
	root.SetArgs(args)

	// pflag parse errors are unambiguously usage failures. Consumers should use
	// Usage for domain-specific argument validation in Args/RunE callbacks.
	root.SetFlagErrorFunc(func(_ *cobra.Command, err error) error {
		return Usage(err)
	})

	// Cobra returns an ordinary error for an unknown subcommand. Capture command
	// lookup failure before execution so Fang can still render the native error
	// while charmcli applies the shared usage exit code afterwards.
	lookupUsage := false
	if _, _, findErr := root.Find(args); findErr != nil {
		lookupUsage = true
	}

	options := append([]fang.Option{}, app.fangOptions...)
	options = append(options, fang.WithErrorHandler(app.errorHandler()))

	err := fang.Execute(ctx, root, options...)
	if err != nil && lookupUsage {
		return 2
	}
	return ExitCode(err)
}

func (app *App) errorHandler() fang.ErrorHandler {
	return func(w io.Writer, styles fang.Styles, err error) {
		if err == nil || isSilentError(err) || errors.Is(err, context.Canceled) {
			return
		}
		display := err
		if app.errorTransform != nil {
			if transformed := app.errorTransform(err); transformed != nil {
				display = transformed
			}
		}
		fang.DefaultErrorHandler(w, styles, display)
	}
}
