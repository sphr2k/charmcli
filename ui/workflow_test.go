package ui

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"sync"
	"testing"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/sphr2k/charmcli"
	"github.com/sphr2k/charmcli/render"
	"github.com/sphr2k/charmcli/testkit"
)

func testWorkflow() Workflow {
	return Workflow{
		Title: "Apply node  homelab-node-3",
		Steps: []WorkflowStep{
			{ID: "detect", Label: "Detect lost worker"},
			{ID: "bootstrap", Label: "Bootstrap worker"},
			{ID: "kubernetes", Label: "Wait for Kubernetes"},
		},
	}
}

func TestWorkflowNonInteractiveWritesDeterministicProgressToStderr(t *testing.T) {
	h := testkit.New(charmcli.Capabilities{})
	err := testWorkflow().Run(context.Background(), h.Runtime, func(_ context.Context, progress Progress) error {
		progress.Start("detect")
		progress.Succeed("detect", Field{Label: "provider", Value: "43ba1bd7"})
		progress.Start("bootstrap")
		progress.Detail("bootstrap", Field{Label: "address", Value: "134.213.168.4"})
		progress.Succeed("bootstrap")
		progress.Skip("kubernetes", Field{Value: "already Ready"})
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	got := h.Err.String()
	for _, want := range []string{
		"Apply node  homelab-node-3",
		"running    Detect lost worker",
		"succeeded  Detect lost worker",
		"provider   43ba1bd7",
		"address    134.213.168.4",
		"skipped    Wait for Kubernetes",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("stderr missing %q:\n%s", want, got)
		}
	}
	if strings.Contains(got, "\x1b[") {
		t.Fatalf("non-interactive stderr contains ANSI: %q", got)
	}
	if h.Out.Len() != 0 {
		t.Fatalf("stdout polluted: %q", h.Out.String())
	}
}

func TestWorkflowTerminalViewKeepsOneSpinnerAndFinalizesIt(t *testing.T) {
	var output bytes.Buffer
	model := newWorkflowModel(context.Background(), testWorkflow(), 80, nil, render.New(&output))
	model.state.apply(workflowEvent{kind: workflowStart, stepID: "detect"})
	model.state.apply(workflowEvent{kind: workflowStart, stepID: "bootstrap"})
	view := model.View().Content
	if strings.Count(view, string(model.spinner.View())) != 1 {
		t.Fatalf("expected exactly one spinner in view:\n%s", view)
	}
	if !strings.Contains(view, "• Detect lost worker") {
		t.Fatalf("earlier running step did not become stable:\n%s", view)
	}

	model.state.apply(workflowEvent{kind: workflowSucceed, stepID: "detect"})
	model.state.apply(workflowEvent{kind: workflowSucceed, stepID: "bootstrap"})
	model.done = true
	view = model.View().Content
	if strings.Contains(view, string(model.spinner.View())) {
		t.Fatalf("spinner survived completion:\n%s", view)
	}
	if strings.Count(view, "✓") != 2 {
		t.Fatalf("completed steps are not durable:\n%s", view)
	}
}

func TestWorkflowWrapsDetailsToTerminalWidth(t *testing.T) {
	state := newWorkflowState(testWorkflow())
	state.apply(workflowEvent{kind: workflowStart, stepID: "bootstrap"})
	state.apply(workflowEvent{kind: workflowDetail, stepID: "bootstrap", details: []Field{{Label: "error", Value: "the replacement server did not become reachable before the configured timeout elapsed"}}})
	view := renderWorkflow(state, 38, render.Renderer{}, func(string) string { return "◌" })
	for _, line := range strings.Split(strings.TrimSuffix(view, "\n"), "\n") {
		if width := lipgloss.Width(line); width > 38 {
			t.Fatalf("line width %d exceeds terminal width: %q\n%s", width, line, view)
		}
	}
	if !strings.Contains(view, "error") || !strings.Contains(view, "configured") {
		t.Fatalf("detail missing after wrapping:\n%s", view)
	}
}

func TestWorkflowProgressAcceptsParallelEvents(t *testing.T) {
	ctx := context.Background()
	events := make(chan tea.Msg, 4)
	progress := channelProgress{ctx: ctx, events: events}
	var wait sync.WaitGroup
	for _, id := range []string{"detect", "bootstrap"} {
		wait.Add(1)
		go func(stepID string) {
			defer wait.Done()
			progress.Start(stepID)
			progress.Succeed(stepID)
		}(id)
	}
	wait.Wait()
	close(events)

	state := newWorkflowState(testWorkflow())
	for message := range events {
		state.apply(message.(workflowEvent))
	}
	if state.states["detect"] != StepSucceeded || state.states["bootstrap"] != StepSucceeded {
		t.Fatalf("parallel events lost: %#v", state.states)
	}
}

func TestWorkflowFinishesFailedAndCancelledSteps(t *testing.T) {
	state := newWorkflowState(testWorkflow())
	state.apply(workflowEvent{kind: workflowStart, stepID: "bootstrap"})
	boom := errors.New("boom")
	if got := state.finish(boom); !errors.Is(got, boom) {
		t.Fatalf("finish error = %v, want %v", got, boom)
	}
	if state.states["bootstrap"] != StepFailed {
		t.Fatalf("failed state = %v", state.states["bootstrap"])
	}

	state = newWorkflowState(testWorkflow())
	state.apply(workflowEvent{kind: workflowStart, stepID: "bootstrap"})
	state.finish(context.Canceled)
	if state.states["bootstrap"] != StepCancelled {
		t.Fatalf("cancelled state = %v", state.states["bootstrap"])
	}
	if marker := workflowMarker(state.states["bootstrap"], ""); marker != "×" {
		t.Fatalf("cancelled marker = %q", marker)
	}
}
