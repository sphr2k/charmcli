package render

import (
	"fmt"
	"io"
	"strings"

	"charm.land/lipgloss/v2"
)

// Tone identifies the semantic presentation of a table cell.
type Tone uint8

const (
	TonePlain Tone = iota
	ToneAccent
	ToneSuccess
	ToneWarning
	ToneError
	ToneMuted
)

// Cell is a semantic table value. Text remains unstyled until the table is
// rendered, so layout and structured output never contain ANSI sequences.
type Cell struct {
	Text string
	Tone Tone
}

func Text(value string) Cell { return Cell{Text: value} }

func Styled(value string, tone Tone) Cell { return Cell{Text: value, Tone: tone} }

// Table renders a compact, ANSI-safe human table.
type Table struct {
	Headers []string
	Rows    [][]Cell
}

func (r Renderer) Table(w io.Writer, table Table) error {
	widths := make([]int, len(table.Headers))
	for i, header := range table.Headers {
		widths[i] = lipgloss.Width(header)
	}
	for _, row := range table.Rows {
		for i, cell := range row {
			if width := lipgloss.Width(cell.Text); width > widths[i] {
				widths[i] = width
			}
		}
	}
	if err := writeTableRow(w, table.Headers, widths, nil); err != nil {
		return err
	}
	for _, row := range table.Rows {
		if err := writeTableRow(w, cellTexts(row), widths, func(i int, value string) string {
			return r.applyTone(row[i].Tone, value)
		}); err != nil {
			return err
		}
	}
	return nil
}

func cellTexts(row []Cell) []string {
	values := make([]string, len(row))
	for i, cell := range row {
		values[i] = cell.Text
	}
	return values
}

func writeTableRow(w io.Writer, values []string, widths []int, style func(int, string) string) error {
	for i, value := range values {
		if i > 0 {
			if _, err := io.WriteString(w, "  "); err != nil {
				return err
			}
		}
		rendered := value
		if style != nil {
			rendered = style(i, value)
		}
		if _, err := io.WriteString(w, rendered); err != nil {
			return err
		}
		if i < len(values)-1 {
			if _, err := io.WriteString(w, strings.Repeat(" ", widths[i]-lipgloss.Width(value))); err != nil {
				return err
			}
		}
	}
	_, err := fmt.Fprintln(w)
	return err
}

func (r Renderer) applyTone(tone Tone, value string) string {
	switch tone {
	case ToneAccent:
		return r.Accent(value)
	case ToneSuccess:
		return r.Success(value)
	case ToneWarning:
		return r.Warning(value)
	case ToneError:
		return r.Error(value)
	case ToneMuted:
		return r.Muted(value)
	default:
		return value
	}
}
