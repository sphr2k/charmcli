package render

import (
	"strings"

	"charm.land/lipgloss/v2"
)

const (
	// DefaultDetailGap separates adjacent key/value columns.
	DefaultDetailGap = 8
	// WideDetailMinWidth is the default breakpoint for considering two-column detail views.
	WideDetailMinWidth = 80
)

// Detail is a semantic key/value item for human-readable detail views.
type Detail struct {
	Key   string
	Value string
	Tone  Tone
	From  string
	To    string
}

// DetailColumns returns the default maximum detail density for a terminal width.
// Content-aware rendering may still collapse a wide terminal to one column.
func DetailColumns(width int) int {
	if width >= WideDetailMinWidth {
		return 2
	}
	return 1
}

// DetailLines formats aligned key/value lines without leaking styling into
// the values used for width calculation.
func (r Renderer) DetailLines(details []Detail) []string {
	keyWidth := detailKeyWidth(details)
	lines := make([]string, 0, len(details))
	for _, detail := range details {
		lines = append(lines, r.detailLine(detail, keyWidth))
	}
	return lines
}

// DetailGridForWidth chooses the densest canonical layout that actually fits
// the available width. Wide terminals prefer two columns but fall back to one
// when the rendered content would overflow.
func (r Renderer) DetailGridForWidth(details []Detail, width int) []string {
	columns := DetailColumns(width)
	if columns == 1 {
		return r.DetailLines(details)
	}

	lines := r.DetailGrid(details, columns, DefaultDetailGap)
	for _, line := range lines {
		if lipgloss.Width(line) > width {
			return r.DetailLines(details)
		}
	}
	return lines
}

// DetailGrid formats details row-major into compact aligned columns. Prefer
// DetailGridForWidth for terminal-aware output; use this method when a caller
// intentionally controls the number of columns.
func (r Renderer) DetailGrid(details []Detail, columns, gap int) []string {
	if len(details) == 0 {
		return nil
	}
	if columns <= 1 {
		return r.DetailLines(details)
	}
	if columns > len(details) {
		columns = len(details)
	}
	if gap < 2 {
		gap = 2
	}

	keyWidths := make([]int, columns)
	for i, detail := range details {
		column := i % columns
		if width := lipgloss.Width(detail.Key); width > keyWidths[column] {
			keyWidths[column] = width
		}
	}

	cellWidths := make([]int, columns)
	for i, detail := range details {
		column := i % columns
		width := keyWidths[column] + 2 + lipgloss.Width(detailValueText(detail))
		if width > cellWidths[column] {
			cellWidths[column] = width
		}
	}

	rows := (len(details) + columns - 1) / columns
	lines := make([]string, 0, rows)
	for row := 0; row < rows; row++ {
		var line strings.Builder
		for column := 0; column < columns; column++ {
			index := row*columns + column
			if index >= len(details) {
				break
			}
			if column > 0 {
				line.WriteString(strings.Repeat(" ", gap))
			}

			rendered := r.detailLine(details[index], keyWidths[column])
			line.WriteString(rendered)
			if column < columns-1 && index+1 < len(details) {
				line.WriteString(strings.Repeat(" ", cellWidths[column]-lipgloss.Width(rendered)))
			}
		}
		lines = append(lines, line.String())
	}
	return lines
}

func detailKeyWidth(details []Detail) int {
	width := 0
	for _, detail := range details {
		if current := lipgloss.Width(detail.Key); current > width {
			width = current
		}
	}
	return width
}

func detailValueText(detail Detail) string {
	if detail.From != "" && detail.To != "" {
		return detail.From + " → " + detail.To
	}
	return detail.Value
}

func (r Renderer) detailLine(detail Detail, keyWidth int) string {
	key := detail.Key + strings.Repeat(" ", keyWidth-lipgloss.Width(detail.Key))
	value := detail.Value
	if detail.From != "" && detail.To != "" {
		value = r.Muted(detail.From) + r.Muted(" → ") + r.Warning(detail.To)
	} else {
		value = r.applyTone(detail.Tone, value)
	}
	return r.Muted(key) + "  " + value
}
