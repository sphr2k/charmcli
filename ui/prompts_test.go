package ui

import (
	"context"
	"errors"
	"testing"

	"github.com/sphr2k/charmcli"
	"github.com/sphr2k/charmcli/testkit"
)

func TestSecretRefusesNonInteractiveRuntime(t *testing.T) {
	h := testkit.New(charmcli.Capabilities{})
	_, err := Secret(context.Background(), h.Runtime, SecretOptions{Title: "secret"})
	if !errors.Is(err, ErrNonInteractive) {
		t.Fatalf("err=%v, want ErrNonInteractive", err)
	}
}

func TestConfirmBypassWorksWithoutTTY(t *testing.T) {
	h := testkit.New(charmcli.Capabilities{})
	confirmed, err := Confirm(context.Background(), h.Runtime, ConfirmOptions{Bypass: true})
	if err != nil || !confirmed {
		t.Fatalf("confirmed=%v err=%v", confirmed, err)
	}
}

func TestConfirmExactValidation(t *testing.T) {
	validate := validateExact("worker-03")
	if err := validate("worker-02"); err == nil {
		t.Fatal("wrong value unexpectedly accepted")
	}
	if err := validate(" worker-03 \n"); err != nil {
		t.Fatalf("exact value rejected: %v", err)
	}
}

func TestConfirmExactBypassWorksWithoutTTY(t *testing.T) {
	h := testkit.New(charmcli.Capabilities{})
	if err := ConfirmExact(context.Background(), h.Runtime, ConfirmExactOptions{Expected: "worker-03", Bypass: true}); err != nil {
		t.Fatal(err)
	}
}
