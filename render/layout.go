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
// A mandatory left accent (▌) uses the shared accent tone. When width permits,
// status is aligned to the right edge of the first line. The returned slice
// ends with a blank line for breathing room before the first section.
func (r Renderer) ResourceHeaderLines(header ResourceHeader, width int) []string {
	const accentGlyph = "▌"
	accent := r.Accent(accentGlyph)
	name := r.Title(header.Name)

	// Unstyled widths for layout math (ANSI must not affect spacing).
	nameW := lipgloss.Width(header.Name)
	accentW := lipgloss.Width(accentGlyph) + 1 // glyph + following space

	line := accent + " " + name

	if header.Status.Text != "" {
		statusText := "● " + header.Status.Text
		status := r.applyTone(header.Status.Tone, statusText)
		statusW := lipgloss.Width(statusText)
		minimum := accentW + nameW + 2 + statusW
		if width >= minimum {
			pad := width - accentW - nameW - statusW
			if pad < 2 {
				pad = 2
			}
			line += strings.Repeat(" ", pad) + status
		} else {
			line += "  " + status
		}
	}

	lines := []string{line}
	if len(header.Meta) > 0 {
		// Indent meta to align under the name (past the accent + space).
		lines = append(lines, strings.Repeat(" ", accentW)+r.Muted(strings.Join(header.Meta, " · ")))
	}
	// Breathing room before the first section.
	lines = append(lines, "")
	return lines
}

// Rule renders the quiet horizontal separator used under section headings.
func (r Renderer) Rule(width int) string {
	if width <= 0 {
		return ""
	}
	return r.Muted(strings.Repeat("─", width))
}

// SectionLines renders the canonical static section header: a Title Case label
// followed by a quiet rule. It deliberately does not use a timeline glyph and
// does not force ALL CAPS.
func (r Renderer) SectionLines(title string, width int) []string {
	title = strings.TrimSpace(title)
	if width < lipgloss.Width(title) {
		width = lipgloss.Width(title)
	}
	return []string{r.Accent(title), r.Rule(width)}
}
