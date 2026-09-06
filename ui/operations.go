package ui

import (
	"fmt"
	"io"

	"github.com/sphr2k/charmcli/render"
)

// Field is a plain-text context field rendered by an operation view.
type Field struct {
	Label string
	Value string
}

// RiskLevel describes the semantic severity of an operation risk.
type RiskLevel uint8

const (
	RiskInfo RiskLevel = iota
	RiskWarning
	RiskBlocking
)

// Risk is a plain-text operation risk.
type Risk struct {
	Level   RiskLevel
	Message string
}

// ChangeSet summarizes a bounded set of changed paths.
type ChangeSet struct {
	Added, Modified, Deleted int
	Paths                    []string
	Omitted                  int
}

// OperationPlan is the shared read-only representation shown before apply.
type OperationPlan struct {
	Title   string
	Context []Field
	Changes ChangeSet
	Risks   []Risk
}

// EventLevel identifies the semantic status of an external engine event.
type EventLevel uint8

const (
	EventInfo EventLevel = iota
	EventSuccess
	EventWarning
	EventError
)

// Event is one plain-text event emitted by a long-running engine.
type Event struct {
	Level   EventLevel
	Message string
}

// EventStream presents engine events without permitting engine-owned styles.
type EventStream struct {
	Writer   io.Writer
	Renderer render.Renderer
}

// Event renders one completed event line.
func (s EventStream) Event(event Event) {
	tone, mark := render.TonePlain, "•"
	switch event.Level {
	case EventSuccess:
		tone, mark = render.ToneSuccess, "✓"
	case EventWarning:
		tone, mark = render.ToneWarning, "!"
	case EventError:
		tone, mark = render.ToneError, "×"
	}
	printStatus(s.Writer, s.Renderer, mark, event.Message, "", tone)
}

// OperationResult is the shared terminal state of an operation.
type OperationResult struct {
	OK      bool
	Message string
	Fields  []Field
}

// PrintOperationResult prints the terminal operation state and artifacts.
func PrintOperationResult(w io.Writer, renderer render.Renderer, result OperationResult) {
	printSection(w, renderer, "Result")
	tone, mark, label := render.ToneSuccess, "✓", "completed"
	if !result.OK {
		tone, mark, label = render.ToneError, "×", "failed"
	}
	printStatus(w, renderer, mark, label, result.Message, tone)
	for _, field := range result.Fields {
		_, _ = fmt.Fprintln(w, "  "+renderer.Muted(field.Label)+"  "+renderer.Accent(field.Value))
	}
}

// PrintOperationPlan renders a scan-friendly operation plan without exposing
// Lip Gloss styles to consumers.
func PrintOperationPlan(w io.Writer, renderer render.Renderer, plan OperationPlan) {
	_, _ = fmt.Fprintln(w, renderer.Title(plan.Title))
	for _, field := range plan.Context {
		_, _ = fmt.Fprintln(w, renderer.Muted(field.Label)+"  "+renderer.Accent(field.Value))
	}
	if plan.Changes.Added != 0 || plan.Changes.Modified != 0 || plan.Changes.Deleted != 0 {
		printSection(w, renderer, "Changes")
		printStatus(w, renderer, "✓", "added", fmt.Sprintf("+%d", plan.Changes.Added), render.ToneSuccess)
		printStatus(w, renderer, "✓", "modified", fmt.Sprintf("~%d", plan.Changes.Modified), render.ToneSuccess)
		deleteTone := render.TonePlain
		if plan.Changes.Deleted != 0 {
			deleteTone = render.ToneWarning
		}
		printStatus(w, renderer, "!", "deleted", fmt.Sprintf("-%d", plan.Changes.Deleted), deleteTone)
		for _, path := range plan.Changes.Paths {
			_, _ = fmt.Fprintln(w, "    "+renderer.Muted(path))
		}
		if plan.Changes.Omitted > 0 {
			_, _ = fmt.Fprintf(w, "    %s\n", renderer.Muted(fmt.Sprintf("… %d more", plan.Changes.Omitted)))
		}
	}
	if len(plan.Risks) > 0 {
		printSection(w, renderer, "Risks")
		for _, risk := range plan.Risks {
			tone, mark := render.TonePlain, "•"
			if risk.Level == RiskWarning {
				tone, mark = render.ToneWarning, "!"
			}
			if risk.Level == RiskBlocking {
				tone, mark = render.ToneError, "×"
			}
			printStatus(w, renderer, mark, risk.Message, "", tone)
		}
	}
}

func printSection(w io.Writer, renderer render.Renderer, title string) {
	_, _ = fmt.Fprintln(w)
	for _, line := range renderer.SectionLines(title, 0) {
		if line != "" {
			_, _ = fmt.Fprintln(w, line)
		}
	}
}

func printStatus(w io.Writer, renderer render.Renderer, mark, label, value string, tone render.Tone) {
	line := renderer.Style(mark, tone) + " " + renderer.Muted(label)
	if value != "" {
		line += "  " + renderer.Style(value, tone)
	}
	_, _ = fmt.Fprintln(w, line)
}
