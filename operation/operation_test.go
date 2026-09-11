package operation

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/sphr2k/charmcli"
	"github.com/sphr2k/charmcli/render"
)

func TestPlanEmpty(t *testing.T) {
	if !(Plan{}).Empty() {
		t.Fatal("zero plan should be empty")
	}
	if (Plan{Changes: []Change{{Resource: "node/a", Action: "update"}}}).Empty() {
		t.Fatal("plan with change should not be empty")
	}
}

func TestPreparedCloseRunsOnce(t *testing.T) {
	calls := 0
	want := errors.New("close failed")
	prepared := NewPrepared(Plan{}, "value", func() error {
		calls++
		return want
	})
	if err := prepared.Close(); !errors.Is(err, want) {
		t.Fatalf("first close error = %v, want %v", err, want)
	}
	if err := prepared.Close(); !errors.Is(err, want) {
		t.Fatalf("second close error = %v, want %v", err, want)
	}
	if calls != 1 {
		t.Fatalf("cleanup calls = %d, want 1", calls)
	}
}

func TestPrintPlanRendersChangesAndRisks(t *testing.T) {
	var out bytes.Buffer
	PrintPlan(&out, render.New(&out), Plan{
		Summary: "2 resources drifted",
		Changes: []Change{{Resource: "node/worker-01", Action: "replace", Details: []string{"image changed"}}},
		Risks:   []Risk{{Level: RiskWarning, Message: "node will restart"}},
	}, RenderOptions{Title: "Node plan"})
	for _, want := range []string{"Node plan", "2 resources drifted", "node/worker-01", "replace", "image changed", "node will restart"} {
		if !strings.Contains(out.String(), want) {
			t.Fatalf("rendered plan missing %q:\n%s", want, out.String())
		}
	}
}

func TestConfirmApplySkipsPromptForEmptyPlan(t *testing.T) {
	ok, err := ConfirmApply(context.Background(), nil, Plan{}, ConfirmOptions{})
	if err != nil || !ok {
		t.Fatalf("empty confirm = (%t, %v), want (true, nil)", ok, err)
	}
}

func TestConfirmApplyHonorsBypass(t *testing.T) {
	runtime := &charmcli.Runtime{}
	ok, err := ConfirmApply(context.Background(), runtime, Plan{Changes: []Change{{Resource: "gitops", Action: "publish"}}}, ConfirmOptions{Bypass: true})
	if err != nil || !ok {
		t.Fatalf("bypassed confirm = (%t, %v), want (true, nil)", ok, err)
	}
}
