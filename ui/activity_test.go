package ui

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/sphr2k/charmcli"
	"github.com/sphr2k/charmcli/testkit"
)

func TestActivityNonInteractiveFallback(t *testing.T) {
	h := testkit.New(charmcli.Capabilities{})
	if err := Activity(context.Background(), h.Runtime, "Check API", func(context.Context) error { return nil }); err != nil {
		t.Fatal(err)
	}
	if got := h.Err.String(); got != "Check API...\nCheck API: done\n" {
		t.Fatalf("stderr=%q", got)
	}
	if h.Out.Len() != 0 {
		t.Fatalf("stdout polluted: %q", h.Out.String())
	}
}

func TestActivityFailure(t *testing.T) {
	h := testkit.New(charmcli.Capabilities{})
	boom := errors.New("boom")
	err := Activity(context.Background(), h.Runtime, "Check API", func(context.Context) error { return boom })
	if !errors.Is(err, boom) || !strings.Contains(h.Err.String(), "failed") {
		t.Fatalf("err=%v stderr=%q", err, h.Err.String())
	}
}
