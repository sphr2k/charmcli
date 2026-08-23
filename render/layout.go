package render

import (
	"strings"

	"charm.land/lipgloss/v2"
)

// ResourceHeader describes the compact identity block at the top of a detail
// view. Status is rendered with a single state marker and Meta is secondary
// evidence separated by middle dots.
type ResourceHeader struct {
	Name   string
	Status Cell
	Meta   []string
}

// ResourceHeaderLines renders resource identity without enclosing it in a box.
// When width permits, status is aligned to the right edge of the first line.
func (r Renderer) ResourceHeaderLines(header ResourceHeader, width int) []string {
	name := r.Title(header.Name)
	line := name

	if header.Status.Text != "" {
		statusText := "● " + header.Status.Text
		status := r.applyTone(header.Status.Tone, statusText)
		minimum := lipgloss.Width(header.Name) + 2 + lipgloss.Width(statusText)
		if width >= minimum {
			line += strings.Repeat(" ", width-lipgloss.Width(header.Name)-lipgloss.Width(statusText)) + status
		} else {
			line += "  " + status
		}
	}

	lines := []string{line}
	if len(header.Meta) > 0 {
		lines = append(lines, r.Muted(strings.Join(header.Meta, " · ")))
	}
	return lines
}

// Rule renders the quiet horizontal separator used under section headings.
func (r Renderer) Rule(width int) string {
	if width <= 0 {
		return ""
	}
	return r.Muted(strings.Repeat("─", width))
}

// SectionLines renders the canonical static section header: an uppercase label
// followed by a quiet rule. It deliberately does not use a timeline glyph.
func (r Renderer) SectionLines(title string, width int) []string {
	title = strings.ToUpper(strings.TrimSpace(title))
	if width < lipgloss.Width(title) {
		width = lipgloss.Width(title)
	}
	return []string{r.Accent(title), r.Rule(width)}
}
