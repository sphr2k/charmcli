package charmcli

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"

	"charm.land/huh/v2"
	"github.com/spf13/cobra"
)

func testStreams() (Streams, *bytes.Buffer, *bytes.Buffer) {
	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	return Streams{In: &bytes.Buffer{}, Out: out, Err: errOut}, out, errOut
}

func TestAppSuccessUsesInjectedStreams(t *testing.T) {
	streams, out, _ := testStreams()
	app := New(WithRuntimeOptions(WithCapabilities(Capabilities{})))
	code := app.Run(context.Background(), nil, streams, func(_ *Runtime) *cobra.Command {
		return &cobra.Command{Use: "test", RunE: func(cmd *cobra.Command, _ []string) error {
			_, err := cmd.OutOrStdout().Write([]byte("ok\n"))
			return err
		}}
	})
	if code != 0 || out.String() != "ok\n" {
		t.Fatalf("code=%d out=%q", code, out.String())
	}
}

func TestAppTypedUsageExit(t *testing.T) {
	streams, _, _ := testStreams()
	app := New(WithRuntimeOptions(WithCapabilities(Capabilities{})))
	code := app.Run(context.Background(), nil, streams, func(_ *Runtime) *cobra.Command {
		return &cobra.Command{Use: "test", RunE: func(*cobra.Command, []string) error {
			return Usage(errors.New("bad invocation"))
		}}
	})
	if code != 2 {
		t.Fatalf("code=%d, want 2", code)
	}
}

func TestAppExecutionFailureExit(t *testing.T) {
	streams, _, _ := testStreams()
	app := New(WithRuntimeOptions(WithCapabilities(Capabilities{})))
	code := app.Run(context.Background(), nil, streams, func(_ *Runtime) *cobra.Command {
		return &cobra.Command{Use: "test", RunE: func(*cobra.Command, []string) error {
			return Failure(errors.New("API unavailable"))
		}}
	})
	if code != 2 {
		t.Fatalf("code=%d, want 2", code)
	}
}

func TestAppPlainExecutionErrorExit(t *testing.T) {
	streams, _, _ := testStreams()
	app := New(WithRuntimeOptions(WithCapabilities(Capabilities{})))
	code := app.Run(context.Background(), nil, streams, func(_ *Runtime) *cobra.Command {
		return &cobra.Command{Use: "test", RunE: func(*cobra.Command, []string) error {
			return errors.New("API unavailable")
		}}
	})
	if code != 2 {
		t.Fatalf("code=%d, want 2", code)
	}
}

func TestAppUnknownCommandIsUsage(t *testing.T) {
	streams, _, errOut := testStreams()
	app := New(WithRuntimeOptions(WithCapabilities(Capabilities{})))
	code := app.Run(context.Background(), []string{"missing"}, streams, func(_ *Runtime) *cobra.Command {
		root := &cobra.Command{Use: "test"}
		root.AddCommand(&cobra.Command{Use: "get"})
		return root
	})
	if code != 2 {
		t.Fatalf("code=%d, want 2; stderr=%q", code, errOut.String())
	}
	if !strings.Contains(strings.ToLower(errOut.String()), "unknown command") {
		t.Fatalf("stderr=%q", errOut.String())
	}
}

func TestAppSilentExitDoesNotRenderError(t *testing.T) {
	streams, _, errOut := testStreams()
	app := New(WithRuntimeOptions(WithCapabilities(Capabilities{})))
	code := app.Run(context.Background(), nil, streams, func(_ *Runtime) *cobra.Command {
		return &cobra.Command{Use: "doctor", RunE: func(*cobra.Command, []string) error {
			return Exit(1)
		}}
	})
	if code != 1 {
		t.Fatalf("code=%d, want 1", code)
	}
	if strings.TrimSpace(errOut.String()) != "" {
		t.Fatalf("stderr=%q, want empty", errOut.String())
	}
}

func TestAppUserAbortDoesNotRenderError(t *testing.T) {
	streams, _, errOut := testStreams()
	app := New(WithRuntimeOptions(WithCapabilities(Capabilities{})))
	code := app.Run(context.Background(), nil, streams, func(_ *Runtime) *cobra.Command {
		return &cobra.Command{Use: "test", RunE: func(*cobra.Command, []string) error {
			return huh.ErrUserAborted
		}}
	})
	if code != 130 {
		t.Fatalf("code=%d, want 130", code)
	}
	if strings.TrimSpace(errOut.String()) != "" {
		t.Fatalf("stderr=%q, want empty", errOut.String())
	}
}

func TestAppErrorTransformChangesPresentationOnly(t *testing.T) {
	streams, _, errOut := testStreams()
	app := New(
		WithRuntimeOptions(WithCapabilities(Capabilities{})),
		WithErrorTransform(func(error) error { return errors.New("redacted") }),
	)
	code := app.Run(context.Background(), nil, streams, func(_ *Runtime) *cobra.Command {
		return &cobra.Command{Use: "test", RunE: func(*cobra.Command, []string) error {
			return Usage(errors.New("secret token"))
		}}
	})
	if code != 2 {
		t.Fatalf("code=%d, want 2", code)
	}
	presentation := strings.ToLower(errOut.String())
	if !strings.Contains(presentation, "redacted") || strings.Contains(presentation, "secret token") {
		t.Fatalf("stderr=%q", errOut.String())
	}
}
