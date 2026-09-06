package ui

import (
	"fmt"
	"io"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/sphr2k/charmcli/render"
)

// Field is a plain-text context field rendered by an operation view.
type Field struct {
	Label string
	Value string
}

// Header identifies an operation and its plain-text context. Consumers pass
// domain values, never pre-styled terminal markup.
type Header struct {
	Title   string
	Context []Field
}

// Section renders one named operation phase.
type Section struct {
	Title string
}

// StatusLine is a semantic, scan-friendly outcome line.
type StatusLine struct {
	Level EventLevel
	Label string
	Value string
}

// Group is a titled collection of status lines, such as applications in one
// Argo CD wave.
type Group struct {
	Title string
	Items []StatusLine
}

// TableRow is one semantic, marker-prefixed row in a grouped operations
// table. Cell styles remain role-based and are never supplied as markup.
type TableRow struct {
	Level EventLevel
	Cells []render.Cell
}

// TableGroup collects related table rows, such as applications sharing an
// Argo CD sync wave.
type TableGroup struct {
	Title string
	Rows  []TableRow
}

// GroupedTable is the detailed companion to GroupedList when each item needs
// multiple scan-friendly attributes.
type GroupedTable struct {
	Title  string
	Groups []TableGroup
}

// DetailRow is one compact operation item with optional globally aligned
// context and indented secondary detail lines.
type DetailRow struct {
	Level   EventLevel
	Label   string
	Value   string
	Details []string
}

// DetailGroup collects compact detail rows under one operational phase.
type DetailGroup struct {
	Title string
	Rows  []DetailRow
}

// GroupedDetails renders grouped operations without turning every group into
// a separate table. Values align across the complete view; details are only
// shown when callers explicitly provide them.
type GroupedDetails struct {
	Title  string
	Groups []DetailGroup
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
	Label   string
	Value   string
}

// EventStream presents engine events without permitting engine-owned styles.
type EventStream struct {
	Writer   io.Writer
	Renderer render.Renderer
}

// PrintHeader renders the shared operation identity block.
func PrintHeader(w io.Writer, renderer render.Renderer, header Header) {
	_, _ = fmt.Fprintln(w, renderer.Heading(header.Title))
	for _, field := range header.Context {
		_, _ = fmt.Fprintln(w, renderer.Label(field.Label)+"  "+renderer.Value(field.Value))
	}
}

// PrintSection renders a visible workflow phase with the shared quiet rule.
func PrintSection(w io.Writer, renderer render.Renderer, section Section) {
	printSection(w, renderer, section.Title)
}

// PrintStatusLine renders a plain-text status using only semantic tones.
func PrintStatusLine(w io.Writer, renderer render.Renderer, status StatusLine) {
	tone, mark := eventTone(status.Level)
	printStatus(w, renderer, mark, status.Label, status.Value, tone)
}

// PrintGroupedList renders grouped status entries without local layout or
// color decisions in domain commands.
func PrintGroupedList(w io.Writer, renderer render.Renderer, title string, groups []Group) {
	printSection(w, renderer, title)
	for _, group := range groups {
		_, _ = fmt.Fprintln(w, "  "+renderer.Warning(group.Title))
		for _, item := range group.Items {
			PrintStatusLine(w, renderer, item)
		}
	}
}

// PrintGroupedTable renders aligned, semantic component rows beneath groups.
func PrintGroupedTable(w io.Writer, renderer render.Renderer, table GroupedTable) {
	printSection(w, renderer, table.Title)
	for _, group := range table.Groups {
		_, _ = fmt.Fprintln(w, "  "+renderer.Warning(group.Title))
		rows := make([][]render.Cell, 0, len(group.Rows))
		for _, row := range group.Rows {
			tone, mark := eventTone(row.Level)
			cells := make([]render.Cell, 0, len(row.Cells)+1)
			cells = append(cells, render.Styled(mark, tone))
			cells = append(cells, row.Cells...)
			rows = append(rows, cells)
		}
		_ = renderer.Table(w, render.Table{Rows: rows, Indent: 2})
	}
}

// PrintGroupedDetails renders an airy grouped list with an optional aligned
// value column and vertically indented secondary details.
func PrintGroupedDetails(w io.Writer, renderer render.Renderer, details GroupedDetails) {
	printSection(w, renderer, details.Title)
	labelWidth := 0
	for _, group := range details.Groups {
		for _, row := range group.Rows {
			labelWidth = max(labelWidth, lipgloss.Width(row.Label))
		}
	}
	for _, group := range details.Groups {
		_, _ = fmt.Fprintln(w, "  "+renderer.Warning(group.Title))
		for _, row := range group.Rows {
			tone, mark := eventTone(row.Level)
			line := "  " + renderer.Style(mark, tone) + " " + renderer.Accent(row.Label)
			if row.Value != "" {
				line += strings.Repeat(" ", labelWidth-lipgloss.Width(row.Label)+2) + renderer.Success(row.Value)
			}
			_, _ = fmt.Fprintln(w, line)
			for _, detail := range row.Details {
				_, _ = fmt.Fprintln(w, "    "+renderer.Muted(detail))
			}
		}
	}
}

// Event renders one completed event line.
func (s EventStream) Event(event Event) {
	tone, mark := eventTone(event.Level)
	label := event.Label
	if label == "" {
		label = event.Message
	}
	line := s.Renderer.Style(mark, tone) + " " + label
	if event.Value != "" {
		line += "  " + s.Renderer.Style(event.Value, tone)
	}
	_, _ = fmt.Fprintln(s.Writer, line)
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
		_, _ = fmt.Fprintln(w, "  "+renderer.Label(field.Label)+"  "+renderer.Value(field.Value))
	}
}

// PrintOperationPlan renders a scan-friendly operation plan without exposing
// Lip Gloss styles to consumers.
func PrintOperationPlan(w io.Writer, renderer render.Renderer, plan OperationPlan) {
	PrintHeader(w, renderer, Header{Title: plan.Title, Context: plan.Context})
	PrintChangeSet(w, renderer, plan.Changes)
	PrintRisks(w, renderer, plan.Risks)
}

// PrintChangeSet renders a bounded changed-path summary as its own section.
func PrintChangeSet(w io.Writer, renderer render.Renderer, changes ChangeSet) {
	if changes.Added == 0 && changes.Modified == 0 && changes.Deleted == 0 {
		return
	}
	printSection(w, renderer, "Changes")
	printStatus(w, renderer, "✓", "added", fmt.Sprintf("+%d", changes.Added), render.ToneSuccess)
	printStatus(w, renderer, "✓", "modified", fmt.Sprintf("~%d", changes.Modified), render.ToneSuccess)
	deleteTone := render.TonePlain
	if changes.Deleted != 0 {
		deleteTone = render.ToneWarning
	}
	printStatus(w, renderer, "!", "deleted", fmt.Sprintf("-%d", changes.Deleted), deleteTone)
	for _, path := range changes.Paths {
		_, _ = fmt.Fprintln(w, "    "+renderer.Muted(path))
	}
	if changes.Omitted > 0 {
		_, _ = fmt.Fprintf(w, "    %s\n", renderer.Muted(fmt.Sprintf("… %d more", changes.Omitted)))
	}
}

// PrintRisks renders semantic operation risks as their own section.
func PrintRisks(w io.Writer, renderer render.Renderer, risks []Risk) {
	if len(risks) == 0 {
		return
	}
	printSection(w, renderer, "Risks")
	for _, risk := range risks {
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

func eventTone(level EventLevel) (render.Tone, string) {
	switch level {
	case EventSuccess:
		return render.ToneSuccess, "✓"
	case EventWarning:
		return render.ToneWarning, "!"
	case EventError:
		return render.ToneError, "×"
	default:
		return render.TonePlain, "•"
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
	line := renderer.Style(mark, tone) + " " + renderer.Label(label)
	if value != "" {
		line += "  " + renderer.Style(value, tone)
	}
	_, _ = fmt.Fprintln(w, line)
}
