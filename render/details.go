package render

import (
	"strings"

	"charm.land/lipgloss/v2"
)

// Detail is a semantic key/value item for human-readable detail views.
type Detail struct {
	Key   string
	Value string
	Tone  Tone
	From  string
	To    string
}

// DetailLines formats aligned key/value lines without leaking styling into
// the values used for width calculation.
func (r Renderer) DetailLines(details []Detail) []string {
	width := 0
	for _, detail := range details {
		if current := lipgloss.Width(detail.Key); current > width {
			width = current
		}
	}
	lines := make([]string, 0, len(details))
	for _, detail := range details {
		key := detail.Key + strings.Repeat(" ", width-lipgloss.Width(detail.Key))
		value := detail.Value
		if detail.From != "" && detail.To != "" {
			value = r.Muted(detail.From) + r.Muted(" → ") + r.Warning(detail.To)
		} else {
			value = r.applyTone(detail.Tone, value)
		}
		lines = append(lines, r.Muted(key)+"  "+value)
	}
	return lines
}
