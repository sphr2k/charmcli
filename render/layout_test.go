package render

import (
	"reflect"
	"strings"
	"testing"

	"charm.land/huh/v2"
)

func TestSemanticStyleDoesNotLeakHuhPromptString(t *testing.T) {
	style := huh.ThemeCharm(false).Focused.SelectSelector
	if got := style.UnsetString().Render("Server"); strings.Contains(got, "> ") {
		t.Fatalf("static accent style leaked prompt string: %q", got)
	}
}

func TestDetailColumns(t *testing.T) {
	if got := DetailColumns(79); got != 1 {
		t.Fatalf("DetailColumns(79) = %d, want 1", got)
	}
	if got := DetailColumns(80); got != 2 {
		t.Fatalf("DetailColumns(80) = %d, want 2", got)
	}
}

func TestDetailGridPlain(t *testing.T) {
	r := Renderer{}
	details := []Detail{
		{Key: "power", Value: "poweredOn"},
		{Key: "ssh", Value: "reachable", Tone: ToneSuccess},
		{Key: "ipv4", Value: "10.0.7.20"},
	}

	got := r.DetailGrid(details, 2, 4)
	want := []string{
		"power  poweredOn    ssh  reachable",
		"ipv4   10.0.7.20",
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("DetailGrid() = %#v, want %#v", got, want)
	}
}

func TestDetailGridForWidthUsesTwoColumnsWhenTheyFit(t *testing.T) {
	r := Renderer{}
	details := []Detail{
		{Key: "power", Value: "poweredOn"},
		{Key: "ssh", Value: "reachable"},
		{Key: "ipv4", Value: "10.0.7.20"},
		{Key: "state", Value: "Ready"},
	}

	got := r.DetailGridForWidth(details, 80)
	if len(got) != 2 {
		t.Fatalf("DetailGridForWidth() returned %d rows, want 2", len(got))
	}
}

func TestDetailGridForWidthFallsBackWhenContentIsTooWide(t *testing.T) {
	r := Renderer{}
	details := []Detail{
		{Key: "os", Value: "Ubuntu 24.04.4 LTS with a deliberately long descriptive suffix"},
		{Key: "runtime", Value: "containerd://2.3.2"},
	}

	got := r.DetailGridForWidth(details, 80)
	if len(got) != len(details) {
		t.Fatalf("DetailGridForWidth() returned %d rows, want one-column %d", len(got), len(details))
	}
}

func TestDetailGridForWidthKeepsSecondColumnAtStableTrack(t *testing.T) {
	r := Renderer{}
	got := r.DetailGridForWidth([]Detail{
		{Key: "power", Value: "poweredOn"},
		{Key: "ssh", Value: "reachable"},
		{Key: "ipv4", Value: "10.0.7.20"},
		{Key: "disk", Value: "68% · 25G"},
	}, 80)
	if len(got) != 2 {
		t.Fatalf("DetailGridForWidth() returned %d rows, want 2", len(got))
	}
	if index := strings.Index(got[0], "ssh"); index != 40 {
		t.Fatalf("second column starts at %d, want 40: %q", index, got[0])
	}
	if index := strings.Index(got[1], "disk"); index != 40 {
		t.Fatalf("second column starts at %d, want 40: %q", index, got[1])
	}
}

func TestResourceHeaderKeepsStatusTogetherWhenNarrow(t *testing.T) {
	r := Renderer{}
	got := r.ResourceHeaderLines(ResourceHeader{
		Name:   "homelab-node-1",
		Status: Styled("READY", ToneSuccess),
	}, 20)
	if len(got) < 2 || got[1] != "  ● READY" {
		t.Fatalf("narrow header split status: %#v", got)
	}
}

func TestResourceHeaderLinesPlain(t *testing.T) {
	r := Renderer{}
	got := r.ResourceHeaderLines(ResourceHeader{
		Name:   "homelab-node-1",
		Status: Styled("Ready", ToneSuccess),
		Meta:   []string{"esxi", "pet", "192.168.200.1"},
	}, 0)
	want := []string{
		"▌ homelab-node-1  ● Ready",
		"  esxi · pet · 192.168.200.1",
		"",
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("ResourceHeaderLines() = %#v, want %#v", got, want)
	}
}

func TestSectionLinesPlain(t *testing.T) {
	r := Renderer{}
	got := r.SectionLines("Host", 8)
	want := []string{"Host", "────────"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("SectionLines() = %#v, want %#v", got, want)
	}
}
