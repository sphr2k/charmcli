// Package render provides the semantic presentation vocabulary for static
// human-readable command output. Lip Gloss is the rendering substrate; Huh's
// Charm theme supplies the shared color tokens, while charmcli owns layout and
// visual grammar.
package render

import (
	"image/color"
	"io"
	"os"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/colorprofile"
)

// Theme assigns colors to semantic presentation roles. Commands use Renderer
// roles rather than colors, so the visual language remains coherent.
type Theme struct {
	Heading color.Color
	Value   color.Color
	Muted   color.Color
	Success color.Color
	Warning color.Color
	Error   color.Color
}

// DefaultTheme keeps operations readable at terminal density: blue establishes
// hierarchy, cyan carries identifiers and paths, and green/amber/red retain
// their exclusive semantic meaning.
var DefaultTheme = Theme{
	Heading: lipgloss.Color("#5DA8E8"),
	Value:   lipgloss.Color("#179DA7"),
	Muted:   lipgloss.Color("#8C939F"),
	Success: lipgloss.Color("#78B159"),
	Warning: lipgloss.Color("#E7B665"),
	Error:   lipgloss.Color("#EF6B73"),
}

// Renderer applies semantic styles suitable for human output.
type Renderer struct {
	enabled bool
	title   lipgloss.Style
	accent  lipgloss.Style
	success lipgloss.Style
	warning lipgloss.Style
	error   lipgloss.Style
	muted   lipgloss.Style
	code    lipgloss.Style
}

// New derives semantic styles from Huh's Charm theme and converts them to the
// writer's detected color profile. NO_COLOR, CLICOLOR and non-TTY behavior are
// handled by Charm colorprofile detection.
func New(writer io.Writer) Renderer {
	return NewWithTheme(writer, DefaultTheme)
}

// NewWithTheme derives semantic styles from a named role palette. NO_COLOR,
// CLICOLOR and non-TTY behavior are handled by Charm colorprofile detection.
func NewWithTheme(writer io.Writer, theme Theme) Renderer {
	profile := colorprofile.Detect(writer, os.Environ())
	if profile <= colorprofile.ASCII {
		return Renderer{}
	}
	style := func(color color.Color, bold bool) lipgloss.Style {
		result := lipgloss.NewStyle().Foreground(profile.Convert(color)).Inline(true)
		if bold {
			result = result.Bold(true)
		}
		return result
	}

	return Renderer{
		enabled: true,
		title:   style(theme.Heading, true),
		accent:  style(theme.Value, false),
		success: style(theme.Success, false),
		warning: style(theme.Warning, true),
		error:   style(theme.Error, true),
		muted:   style(theme.Muted, false),
		code:    style(theme.Value, false),
	}
}

func (r Renderer) Title(value string) string   { return r.apply(r.title, value) }
func (r Renderer) Heading(value string) string { return r.apply(r.title, value) }
func (r Renderer) Accent(value string) string  { return r.apply(r.accent, value) }
func (r Renderer) Value(value string) string   { return r.apply(r.accent, value) }
func (r Renderer) Success(value string) string { return r.apply(r.success, value) }
func (r Renderer) Warning(value string) string { return r.apply(r.warning, value) }
func (r Renderer) Error(value string) string   { return r.apply(r.error, value) }
func (r Renderer) Muted(value string) string   { return r.apply(r.muted, value) }
func (r Renderer) Label(value string) string   { return r.apply(r.muted, value) }
func (r Renderer) Code(value string) string    { return r.apply(r.code, value) }

// Style applies a semantic tone to a human-readable value.
func (r Renderer) Style(value string, tone Tone) string { return r.applyTone(tone, value) }

func (r Renderer) apply(style lipgloss.Style, value string) string {
	if !r.enabled {
		return value
	}
	return style.Render(value)
}
