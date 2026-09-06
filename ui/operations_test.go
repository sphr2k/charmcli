package ui

import (
	"bytes"
	"testing"

	"github.com/sphr2k/charmcli/render"
)

func TestOperationPlanPlainTextKeepsHierarchyAndChanges(t *testing.T) {
	var out bytes.Buffer
	renderer := render.New(&out)
	PrintOperationPlan(&out, renderer, OperationPlan{
		Title:   "GitOps plan",
		Context: []Field{{Label: "target", Value: "clastix"}},
		Changes: ChangeSet{Added: 2, Modified: 3, Deleted: 1, Paths: []string{"rendered/apps.yaml"}},
		Risks:   []Risk{{Level: RiskWarning, Message: "one rendered file will be deleted"}},
	})

	got := out.String()
	for _, want := range []string{"GitOps plan", "target", "Changes", "+2", "~3", "-1", "Risks", "one rendered file"} {
		if !bytes.Contains([]byte(got), []byte(want)) {
			t.Fatalf("output missing %q:\n%s", want, got)
		}
	}
}

func TestEventStreamAndResultUseSemanticMarkers(t *testing.T) {
	var out bytes.Buffer
	renderer := render.New(&out)
	stream := EventStream{Writer: &out, Renderer: renderer}
	stream.Event(Event{Level: EventSuccess, Message: "Rendering Helm Charts"})
	stream.Event(Event{Level: EventWarning, Message: "one chart is outdated"})
	PrintOperationResult(&out, renderer, OperationResult{OK: true, Message: "pushed main"})

	got := out.String()
	for _, want := range []string{"✓", "Rendering Helm Charts", "!", "one chart is outdated", "Result", "pushed main"} {
		if !bytes.Contains([]byte(got), []byte(want)) {
			t.Fatalf("output missing %q:\n%s", want, got)
		}
	}
}

func TestToolkitPrimitivesKeepPhaseAndWaveHierarchy(t *testing.T) {
	var out bytes.Buffer
	renderer := render.New(&out)
	PrintHeader(&out, renderer, Header{Title: "GitOps render", Context: []Field{{Label: "target", Value: "clastix"}}})
	PrintSection(&out, renderer, Section{Title: "Postprocess"})
	PrintStatusLine(&out, renderer, StatusLine{Level: EventSuccess, Label: "processed applications", Value: "61"})
	PrintGroupedList(&out, renderer, "Applications by wave", []Group{{Title: "Wave 10", Items: []StatusLine{{Level: EventSuccess, Label: "platform/cilium", Value: "cilium"}}}})

	got := out.String()
	for _, want := range []string{"GitOps render", "target", "Postprocess", "✓", "processed applications", "Applications by wave", "Wave 10", "platform/cilium"} {
		if !bytes.Contains([]byte(got), []byte(want)) {
			t.Fatalf("output missing %q:\n%s", want, got)
		}
	}
}
