// Package render provides a small semantic presentation vocabulary derived
// from Charm's native Huh theme. It intentionally contains no domain statuses.
package render

import (
	"io"
	"os"

	"charm.land/huh/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/colorprofile"
)

// Renderer applies semantic Charm-native styles suitable for human output.
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

// New derives styles for the writer's detected color profile. NO_COLOR,
// CLICOLOR and non-TTY behavior are handled by Charm colorprofile detection.
func New(writer io.Writer) Renderer {
	profile := colorprofile.Detect(writer, os.Environ())
	if profile <= colorprofile.ASCII {
		return Renderer{}
	}

	theme := huh.ThemeCharm(false)
	convert := func(style lipgloss.Style) lipgloss.Style {
		if foreground := style.GetForeground(); foreground != nil {
			style = style.Foreground(profile.Convert(foreground))
		}
		if background := style.GetBackground(); background != nil {
			style = style.Background(profile.Convert(background))
		}
		return style
	}

	return Renderer{
		enabled: true,
		title:   convert(theme.Focused.Title),
		accent:  convert(theme.Focused.SelectSelector),
		success: convert(theme.Focused.SelectedOption),
		warning: lipgloss.NewStyle().Foreground(profile.Convert(lipgloss.Yellow)).Bold(true),
		error:   convert(theme.Focused.ErrorMessage).Bold(true),
		muted:   convert(theme.Focused.Description),
		code:    convert(theme.Focused.TextInput.Prompt),
	}
}

func (r Renderer) Title(value string) string   { return r.apply(r.title, value) }
func (r Renderer) Accent(value string) string  { return r.apply(r.accent, value) }
func (r Renderer) Success(value string) string { return r.apply(r.success, value) }
func (r Renderer) Warning(value string) string { return r.apply(r.warning, value) }
func (r Renderer) Error(value string) string   { return r.apply(r.error, value) }
func (r Renderer) Muted(value string) string   { return r.apply(r.muted, value) }
func (r Renderer) Code(value string) string    { return r.apply(r.code, value) }

// Style applies a semantic tone to a human-readable value.
func (r Renderer) Style(value string, tone Tone) string { return r.applyTone(tone, value) }

func (r Renderer) apply(style lipgloss.Style, value string) string {
	if !r.enabled {
		return value
	}
	return style.Render(value)
}
