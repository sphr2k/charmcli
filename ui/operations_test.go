package ui

import (
	"bytes"
	"strings"
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

func TestGroupedDetailsAlignsValuesAcrossWavesAndIndentsDetails(t *testing.T) {
	var out bytes.Buffer
	renderer := render.New(&out)
	PrintGroupedDetails(&out, renderer, GroupedDetails{
		Title: "Applications by wave",
		Groups: []DetailGroup{
			{Title: "Wave 10", Rows: []DetailRow{{Level: EventSuccess, Label: "platform/10-cilium"}, {Level: EventSuccess, Label: "platform/10-cert-manager", Value: "cert-manager", Details: []string{"ns · prereqs · workload · ops"}}}},
			{Title: "Wave 40", Rows: []DetailRow{{Level: EventSuccess, Label: "platform/40-authorino", Value: "authorino-system", Details: []string{"prereqs · workload · routing"}}}},
		},
	})

	lines := strings.Split(out.String(), "\n")
	var certLine, authorinoLine string
	for _, line := range lines {
		if strings.Contains(line, "cert-manager") && strings.Contains(line, "platform/") {
			certLine = line
		}
		if strings.Contains(line, "authorino-system") {
			authorinoLine = line
		}
	}
	if strings.LastIndex(certLine, "cert-manager") != strings.LastIndex(authorinoLine, "authorino-system") {
		t.Fatalf("values are not globally aligned:\n%s", out.String())
	}
	if !strings.Contains(out.String(), "    ns · prereqs · workload · ops") {
		t.Fatalf("details are not indented:\n%s", out.String())
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

func TestEventStreamAndGroupedTableKeepOperationalDetail(t *testing.T) {
	var out bytes.Buffer
	renderer := render.New(&out)
	stream := EventStream{Writer: &out, Renderer: renderer}
	stream.Event(Event{Level: EventSuccess, Label: "Pulling Helm Chart", Value: "cilium 1.20.1"})
	PrintGroupedTable(&out, renderer, GroupedTable{
		Title: "Applications by wave",
		Groups: []TableGroup{{
			Title: "Wave 10",
			Rows: []TableRow{{
				Level: EventSuccess,
				Cells: []render.Cell{
					render.Styled("platform/10-cilium", render.ToneAccent),
					render.Styled("kube-system", render.ToneSuccess),
					render.Styled("ns · workload · routing", render.ToneMuted),
				},
			}},
		}},
	})

	got := out.String()
	for _, want := range []string{"Pulling Helm Chart", "cilium 1.20.1", "Applications by wave", "Wave 10", "platform/10-cilium", "kube-system", "ns · workload · routing"} {
		if !bytes.Contains([]byte(got), []byte(want)) {
			t.Fatalf("output missing %q:\n%s", want, got)
		}
	}
}
