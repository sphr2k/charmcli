package ui

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"
	"sync"

	"charm.land/bubbles/v2/spinner"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/sphr2k/charmcli"
	"github.com/sphr2k/charmcli/render"
)

// StepState is the durable state of a workflow step.
type StepState uint8

const (
	StepPending StepState = iota
	StepRunning
	StepSucceeded
	StepFailed
	StepSkipped
	StepCancelled
)

// WorkflowStep declares one ordered, human-readable unit of work.
type WorkflowStep struct {
	ID    string
	Label string
}

// Progress is the only presentation boundary available to a workflow worker.
// Its methods are safe to call from concurrent goroutines. Workers must not
// write directly to a terminal stream.
type Progress interface {
	Start(stepID string)
	Succeed(stepID string, details ...Field)
	Fail(stepID string, err error)
	Skip(stepID string, details ...Field)
	Detail(stepID string, detail Field)
}

// WorkflowFunc performs a long-running operation and reports its structured
// progress through Progress.
type WorkflowFunc func(context.Context, Progress) error

// Workflow presents a sequence of long-running steps. Activity remains the
// smaller primitive for quiet, one-shot operations.
type Workflow struct {
	Title string
	Steps []WorkflowStep
}

// Run owns stderr for the duration of the workflow. Interactive terminals get
// one Bubble Tea model with at most one spinner; non-interactive callers get
// line-oriented progress without ANSI control sequences.
func (w Workflow) Run(ctx context.Context, runtime *charmcli.Runtime, run WorkflowFunc) error {
	if runtime == nil {
		return fmt.Errorf("nil charmcli runtime")
	}
	if run == nil {
		return fmt.Errorf("nil workflow function")
	}
	if err := w.validate(); err != nil {
		return err
	}
	if !runtime.Capabilities.Interactive {
		return w.runPlain(ctx, runtime.Streams.Err, run)
	}

	model := newWorkflowModel(ctx, w, runtime.Capabilities.Width, run, render.New(runtime.Streams.Err))
	program := tea.NewProgram(model, tea.WithInput(runtime.Streams.In), tea.WithOutput(runtime.Streams.Err), tea.WithContext(ctx))
	finished, err := program.Run()
	if err != nil {
		return err
	}
	return finished.(workflowModel).err
}

func (w Workflow) validate() error {
	if strings.TrimSpace(w.Title) == "" {
		return fmt.Errorf("workflow title is required")
	}
	if len(w.Steps) == 0 {
		return fmt.Errorf("workflow requires at least one step")
	}
	seen := make(map[string]struct{}, len(w.Steps))
	for _, step := range w.Steps {
		if strings.TrimSpace(step.ID) == "" || strings.TrimSpace(step.Label) == "" {
			return fmt.Errorf("workflow steps require an ID and label")
		}
		if _, exists := seen[step.ID]; exists {
			return fmt.Errorf("duplicate workflow step %q", step.ID)
		}
		seen[step.ID] = struct{}{}
	}
	return nil
}

func (w Workflow) runPlain(ctx context.Context, output io.Writer, run WorkflowFunc) error {
	plain := &plainProgress{state: newWorkflowState(w), output: output}
	_, _ = fmt.Fprintln(output, w.Title)
	err := run(ctx, plain)
	plain.finish(err)
	return plain.err
}

type workflowEventKind uint8

const (
	workflowStart workflowEventKind = iota
	workflowSucceed
	workflowFail
	workflowSkip
	workflowDetail
)

type workflowEvent struct {
	kind    workflowEventKind
	stepID  string
	details []Field
	err     error
}

type workflowDoneMsg struct{ err error }

type workflowState struct {
	workflow Workflow
	states   map[string]StepState
	details  map[string][]Field
	active   string
}

func newWorkflowState(workflow Workflow) workflowState {
	states := make(map[string]StepState, len(workflow.Steps))
	details := make(map[string][]Field, len(workflow.Steps))
	for _, step := range workflow.Steps {
		states[step.ID] = StepPending
	}
	return workflowState{workflow: workflow, states: states, details: details}
}

func (s *workflowState) apply(event workflowEvent) {
	if _, exists := s.states[event.stepID]; !exists {
		return
	}
	switch event.kind {
	case workflowStart:
		s.states[event.stepID] = StepRunning
		s.active = event.stepID
	case workflowSucceed:
		s.states[event.stepID] = StepSucceeded
		s.details[event.stepID] = append(s.details[event.stepID], event.details...)
		if s.active == event.stepID {
			s.active = ""
		}
	case workflowFail:
		s.states[event.stepID] = StepFailed
		if event.err != nil {
			s.details[event.stepID] = append(s.details[event.stepID], Field{Label: "error", Value: event.err.Error()})
		}
		if s.active == event.stepID {
			s.active = ""
		}
	case workflowSkip:
		s.states[event.stepID] = StepSkipped
		s.details[event.stepID] = append(s.details[event.stepID], event.details...)
		if s.active == event.stepID {
			s.active = ""
		}
	case workflowDetail:
		s.details[event.stepID] = append(s.details[event.stepID], event.details...)
	}
}

func (s *workflowState) finish(err error) error {
	for _, step := range s.workflow.Steps {
		if s.states[step.ID] != StepRunning {
			continue
		}
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			s.states[step.ID] = StepCancelled
			s.details[step.ID] = append(s.details[step.ID], Field{Label: "status", Value: "cancelled"})
		} else {
			s.states[step.ID] = StepFailed
			s.details[step.ID] = append(s.details[step.ID], Field{Label: "error", Value: "workflow ended before the step completed"})
			if err == nil {
				err = fmt.Errorf("workflow ended with unfinished step %q", step.ID)
			}
		}
	}
	s.active = ""
	return err
}

type channelProgress struct {
	ctx    context.Context
	events chan<- tea.Msg
}

func (p channelProgress) Start(id string) { p.send(workflowEvent{kind: workflowStart, stepID: id}) }
func (p channelProgress) Succeed(id string, details ...Field) {
	p.send(workflowEvent{kind: workflowSucceed, stepID: id, details: details})
}
func (p channelProgress) Fail(id string, err error) {
	p.send(workflowEvent{kind: workflowFail, stepID: id, err: err})
}
func (p channelProgress) Skip(id string, details ...Field) {
	p.send(workflowEvent{kind: workflowSkip, stepID: id, details: details})
}
func (p channelProgress) Detail(id string, detail Field) {
	p.send(workflowEvent{kind: workflowDetail, stepID: id, details: []Field{detail}})
}
func (p channelProgress) send(event workflowEvent) {
	select {
	case <-p.ctx.Done():
	case p.events <- event:
	}
}

type plainProgress struct {
	mu     sync.Mutex
	state  workflowState
	output io.Writer
	err    error
}

func (p *plainProgress) Start(id string) { p.apply(workflowEvent{kind: workflowStart, stepID: id}) }
func (p *plainProgress) Succeed(id string, details ...Field) {
	p.apply(workflowEvent{kind: workflowSucceed, stepID: id, details: details})
}
func (p *plainProgress) Fail(id string, err error) {
	p.apply(workflowEvent{kind: workflowFail, stepID: id, err: err})
}
func (p *plainProgress) Skip(id string, details ...Field) {
	p.apply(workflowEvent{kind: workflowSkip, stepID: id, details: details})
}
func (p *plainProgress) Detail(id string, detail Field) {
	p.apply(workflowEvent{kind: workflowDetail, stepID: id, details: []Field{detail}})
}
func (p *plainProgress) apply(event workflowEvent) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.state.apply(event)
	writePlainEvent(p.output, p.state, event)
}
func (p *plainProgress) finish(err error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.err = p.state.finish(err)
}

type workflowModel struct {
	ctx      context.Context
	state    workflowState
	run      WorkflowFunc
	events   chan tea.Msg
	spinner  spinner.Model
	renderer render.Renderer
	width    int
	done     bool
	err      error
}

func newWorkflowModel(ctx context.Context, workflow Workflow, width int, run WorkflowFunc, renderer render.Renderer) workflowModel {
	return workflowModel{ctx: ctx, state: newWorkflowState(workflow), run: run, events: make(chan tea.Msg, 128), spinner: spinner.New(spinner.WithSpinner(spinner.MiniDot)), renderer: renderer, width: workflowWidth(width)}
}
func (m workflowModel) Init() tea.Cmd {
	start := func() tea.Msg {
		go func() {
			progress := channelProgress{ctx: m.ctx, events: m.events}
			m.events <- workflowDoneMsg{err: m.run(m.ctx, progress)}
		}()
		return nil
	}
	return tea.Batch(m.spinner.Tick, m.nextEvent(), start)
}
func (m workflowModel) nextEvent() tea.Cmd { return func() tea.Msg { return <-m.events } }
func (m workflowModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch value := msg.(type) {
	case workflowEvent:
		m.state.apply(value)
		return m, m.nextEvent()
	case workflowDoneMsg:
		m.err = m.state.finish(value.err)
		m.done = true
		return m, tea.Quit
	case tea.WindowSizeMsg:
		m.width = workflowWidth(value.Width)
		return m, nil
	default:
		var command tea.Cmd
		m.spinner, command = m.spinner.Update(msg)
		return m, command
	}
}
func (m workflowModel) View() tea.View {
	return tea.NewView(renderWorkflow(m.state, m.width, m.renderer, func(stepID string) string {
		if !m.done && stepID == m.state.active {
			return m.spinner.View()
		}
		return ""
	}))
}

func writePlainEvent(output io.Writer, state workflowState, event workflowEvent) {
	step, ok := workflowStep(state.workflow, event.stepID)
	if !ok {
		return
	}
	_, _ = fmt.Fprintf(output, "  %-10s %s\n", plainState(state.states[step.ID]), step.Label)
	for _, detail := range eventDetails(event) {
		writePlainDetail(output, detail)
	}
}
func writePlainDetail(output io.Writer, detail Field) {
	if detail.Label == "" {
		_, _ = fmt.Fprintf(output, "    %s\n", detail.Value)
		return
	}
	_, _ = fmt.Fprintf(output, "    %-10s %s\n", detail.Label, detail.Value)
}
func eventDetails(event workflowEvent) []Field {
	details := append([]Field(nil), event.details...)
	if event.kind == workflowFail && event.err != nil {
		details = append(details, Field{Label: "error", Value: event.err.Error()})
	}
	return details
}

func renderWorkflow(state workflowState, width int, renderer render.Renderer, spinnerFor func(string) string) string {
	lines := []string{renderer.Heading(state.workflow.Title)}
	for _, step := range state.workflow.Steps {
		marker := workflowMarker(state.states[step.ID], spinnerFor(step.ID))
		lines = append(lines, wrapWorkflowLine("  "+marker+" ", step.Label, width, renderer, state.states[step.ID])...)
		for _, detail := range state.details[step.ID] {
			lines = append(lines, wrapWorkflowDetail(detail, width, renderer)...)
		}
	}
	return strings.Join(lines, "\n") + "\n"
}
func wrapWorkflowLine(prefix, value string, width int, renderer render.Renderer, state StepState) []string {
	lines := wrapText(prefix, value, width)
	for i, line := range lines {
		if i == 0 {
			lines[i] = renderer.Style(line, workflowTone(state))
		} else {
			lines[i] = renderer.Muted(line)
		}
	}
	return lines
}
func wrapWorkflowDetail(detail Field, width int, renderer render.Renderer) []string {
	prefix := "      "
	if detail.Label != "" {
		prefix += detail.Label + "  "
	}
	lines := wrapText(prefix, detail.Value, width)
	for i, line := range lines {
		if i == 0 && detail.Label != "" {
			lines[i] = renderer.Muted("      "+detail.Label+"  ") + renderer.Value(strings.TrimPrefix(line, prefix))
		} else {
			lines[i] = renderer.Muted(line)
		}
	}
	return lines
}
func wrapText(prefix, value string, width int) []string {
	if width <= lipgloss.Width(prefix)+1 {
		width = 80
	}
	continuation := strings.Repeat(" ", lipgloss.Width(prefix))
	words := strings.Fields(value)
	if len(words) == 0 {
		return []string{prefix}
	}
	line := prefix
	var lines []string
	for _, word := range words {
		separator := ""
		if line != prefix && line != continuation {
			separator = " "
		}
		if lipgloss.Width(line+separator+word) > width && line != prefix && line != continuation {
			lines = append(lines, line)
			line = continuation + word
			continue
		}
		line += separator + word
	}
	return append(lines, line)
}
func workflowWidth(width int) int {
	if width < 24 {
		return 80
	}
	return width
}
func workflowStep(workflow Workflow, id string) (WorkflowStep, bool) {
	for _, step := range workflow.Steps {
		if step.ID == id {
			return step, true
		}
	}
	return WorkflowStep{}, false
}
func workflowMarker(state StepState, current string) string {
	if current != "" {
		return current
	}
	switch state {
	case StepSucceeded:
		return "✓"
	case StepFailed, StepCancelled:
		return "×"
	case StepSkipped:
		return "–"
	case StepRunning:
		return "•"
	default:
		return "·"
	}
}
func workflowTone(state StepState) render.Tone {
	switch state {
	case StepSucceeded:
		return render.ToneSuccess
	case StepFailed, StepCancelled:
		return render.ToneError
	case StepSkipped, StepPending:
		return render.ToneMuted
	default:
		return render.ToneAccent
	}
}
func plainState(state StepState) string {
	switch state {
	case StepRunning:
		return "running"
	case StepSucceeded:
		return "succeeded"
	case StepFailed:
		return "failed"
	case StepSkipped:
		return "skipped"
	case StepCancelled:
		return "cancelled"
	default:
		return "pending"
	}
}
