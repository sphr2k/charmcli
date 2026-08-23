package ui

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"charm.land/huh/v2"
	"github.com/sphr2k/charmcli"
)

// ErrNonInteractive is returned when a required prompt cannot safely run.
var ErrNonInteractive = errors.New("interactive input required")

type SecretOptions struct {
	Title       string
	Description string
	Placeholder string
	Validate    func(string) error
}

// Secret reads a no-echo value through Huh on the runtime's input/stderr.
func Secret(ctx context.Context, runtime *charmcli.Runtime, options SecretOptions) (string, error) {
	if err := requireInteractive(runtime); err != nil {
		return "", err
	}
	var value string
	field := huh.NewInput().
		Title(options.Title).
		Description(options.Description).
		Placeholder(options.Placeholder).
		EchoMode(huh.EchoModeNone).
		Value(&value)
	if options.Validate != nil {
		field.Validate(options.Validate)
	}
	if err := runField(ctx, runtime, field); err != nil {
		return "", err
	}
	return value, nil
}

type ConfirmOptions struct {
	Title       string
	Description string
	Default     bool
	Bypass      bool
	Affirmative string
	Negative    string
}

// Confirm asks for a boolean confirmation. Bypass is intended for an explicit
// command-level flag such as --yes, not as an implicit non-interactive default.
func Confirm(ctx context.Context, runtime *charmcli.Runtime, options ConfirmOptions) (bool, error) {
	if options.Bypass {
		return true, nil
	}
	if err := requireInteractive(runtime); err != nil {
		return false, err
	}
	value := options.Default
	field := huh.NewConfirm().
		Title(options.Title).
		Description(options.Description).
		Value(&value)
	if options.Affirmative != "" {
		field.Affirmative(options.Affirmative)
	}
	if options.Negative != "" {
		field.Negative(options.Negative)
	}
	if err := runField(ctx, runtime, field); err != nil {
		return false, err
	}
	return value, nil
}

type ConfirmExactOptions struct {
	Title       string
	Description string
	Expected    string
	Placeholder string
	Bypass      bool
}

// ConfirmExact requires an exact typed token/name before returning success.
func ConfirmExact(ctx context.Context, runtime *charmcli.Runtime, options ConfirmExactOptions) error {
	if options.Bypass {
		return nil
	}
	if err := requireInteractive(runtime); err != nil {
		return err
	}
	if options.Expected == "" {
		return errors.New("exact confirmation requires a non-empty expected value")
	}

	var value string
	field := huh.NewInput().
		Title(options.Title).
		Description(options.Description).
		Placeholder(options.Placeholder).
		Value(&value).
		Validate(validateExact(options.Expected))
	return runField(ctx, runtime, field)
}

// Option is a small presentation/value pair used by Select and MultiSelect.
type Option[T comparable] struct {
	Label string
	Value T
}

type SelectOptions[T comparable] struct {
	Title       string
	Description string
	Options     []Option[T]
	Initial     T
	Filtering   bool
	Height      int
}

// Select chooses one value through Huh. Advanced/dynamic forms should use Huh
// directly rather than extending this helper into another form framework.
func Select[T comparable](ctx context.Context, runtime *charmcli.Runtime, options SelectOptions[T]) (T, error) {
	var zero T
	if err := requireInteractive(runtime); err != nil {
		return zero, err
	}
	value := options.Initial
	field := huh.NewSelect[T]().
		Title(options.Title).
		Description(options.Description).
		Options(toHuhOptions(options.Options)...).
		Value(&value).
		Filtering(options.Filtering)
	if options.Height > 0 {
		field.Height(options.Height)
	}
	if err := runField(ctx, runtime, field); err != nil {
		return zero, err
	}
	return value, nil
}

type MultiSelectOptions[T comparable] struct {
	Title       string
	Description string
	Options     []Option[T]
	Initial     []T
	Height      int
}

// MultiSelect chooses multiple values through Huh.
func MultiSelect[T comparable](ctx context.Context, runtime *charmcli.Runtime, options MultiSelectOptions[T]) ([]T, error) {
	if err := requireInteractive(runtime); err != nil {
		return nil, err
	}
	value := append([]T(nil), options.Initial...)
	field := huh.NewMultiSelect[T]().
		Title(options.Title).
		Description(options.Description).
		Options(toHuhOptions(options.Options)...).
		Value(&value)
	if options.Height > 0 {
		field.Height(options.Height)
	}
	if err := runField(ctx, runtime, field); err != nil {
		return nil, err
	}
	return value, nil
}

func validateExact(expected string) func(string) error {
	return func(value string) error {
		if strings.TrimSpace(value) != expected {
			return fmt.Errorf("type %s exactly to confirm", expected)
		}
		return nil
	}
}

func toHuhOptions[T comparable](options []Option[T]) []huh.Option[T] {
	result := make([]huh.Option[T], 0, len(options))
	for _, option := range options {
		result = append(result, huh.NewOption(option.Label, option.Value))
	}
	return result
}

func runField(ctx context.Context, runtime *charmcli.Runtime, field huh.Field) error {
	form := huh.NewForm(huh.NewGroup(field)).
		WithInput(runtime.Streams.In).
		WithOutput(runtime.Streams.Err).
		WithAccessible(runtime.Capabilities.Accessible)
	return form.RunWithContext(ctx)
}

func requireInteractive(runtime *charmcli.Runtime) error {
	if runtime == nil || !runtime.Capabilities.Interactive {
		return ErrNonInteractive
	}
	return nil
}
