package ui

import (
	"context"
	"fmt"

	"charm.land/bubbles/v2/spinner"
	tea "charm.land/bubbletea/v2"
	"charm.land/huh/v2"
	"charm.land/lipgloss/v2"
	"github.com/sphr2k/charmcli"
)

// ActivityFunc performs one long-running operation.
type ActivityFunc func(context.Context) error

// Activity runs one operation with an animated Bubbles spinner on an
// interactive terminal and deterministic line-oriented stderr otherwise.
func Activity(ctx context.Context, runtime *charmcli.Runtime, title string, run ActivityFunc) error {
	if runtime == nil {
		return fmt.Errorf("nil charmcli runtime")
	}
	if run == nil {
		return fmt.Errorf("nil activity function")
	}
	if !runtime.Capabilities.Interactive {
		_, _ = fmt.Fprintf(runtime.Streams.Err, "%s...\n", title)
		err := run(ctx)
		if err != nil {
			_, _ = fmt.Fprintf(runtime.Streams.Err, "%s: failed\n", title)
			return err
		}
		_, _ = fmt.Fprintf(runtime.Streams.Err, "%s: done\n", title)
		return nil
	}

	model := newActivityModel(ctx, title, run)
	program := tea.NewProgram(
		model,
		tea.WithInput(runtime.Streams.In),
		tea.WithOutput(runtime.Streams.Err),
		tea.WithContext(ctx),
	)
	finished, err := program.Run()
	if err != nil {
		return err
	}
	result := finished.(activityModel)
	styles := huh.ThemeCharm(false)
	if result.err != nil {
		mark := lipgloss.NewStyle().Foreground(styles.Focused.ErrorMessage.GetForeground()).Render("✗")
		_, _ = fmt.Fprintf(runtime.Streams.Err, "%s %s\n", mark, title)
		return result.err
	}
	mark := lipgloss.NewStyle().Foreground(styles.Focused.SelectedOption.GetForeground()).Render("✓")
	_, _ = fmt.Fprintf(runtime.Streams.Err, "%s %s\n", mark, title)
	return nil
}

// Step is one named operation in a sequential workflow.
type Step struct {
	Title string
	Run   ActivityFunc
}

// RunSteps executes steps in order and stops at the first failure.
func RunSteps(ctx context.Context, runtime *charmcli.Runtime, steps ...Step) error {
	for _, step := range steps {
		if err := Activity(ctx, runtime, step.Title, step.Run); err != nil {
			return err
		}
	}
	return nil
}

type activityDoneMsg struct{ err error }

type activityModel struct {
	ctx     context.Context
	title   string
	run     ActivityFunc
	spinner spinner.Model
	done    bool
	err     error
}

func newActivityModel(ctx context.Context, title string, run ActivityFunc) activityModel {
	model := spinner.New(spinner.WithSpinner(spinner.MiniDot))
	styles := huh.ThemeCharm(false)
	model.Style = lipgloss.NewStyle().Foreground(styles.Focused.SelectSelector.GetForeground())
	return activityModel{ctx: ctx, title: title, run: run, spinner: model}
}

func (m activityModel) Init() tea.Cmd {
	work := func() tea.Msg { return activityDoneMsg{err: m.run(m.ctx)} }
	return tea.Batch(m.spinner.Tick, work)
}

func (m activityModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case activityDoneMsg:
		m.done = true
		m.err = msg.err
		return m, tea.Quit
	default:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	}
}

func (m activityModel) View() tea.View {
	if m.done {
		return tea.NewView("")
	}
	return tea.NewView(fmt.Sprintf("%s %s", m.spinner.View(), m.title))
}
