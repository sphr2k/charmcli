// Package operation provides small, domain-neutral primitives for CLIs that
// expose plan/apply workflows. It deliberately does not own orchestration,
// persistence, Cobra wiring, or any concrete infrastructure domain.
package operation

import (
	"context"
	"fmt"
	"io"
	"sync"

	"github.com/sphr2k/charmcli"
	"github.com/sphr2k/charmcli/render"
	"github.com/sphr2k/charmcli/ui"
)

// RiskLevel is the semantic severity of a planned operation risk.
type RiskLevel uint8

const (
	RiskInfo RiskLevel = iota
	RiskWarning
	RiskDanger
)

// Risk is a plain-text warning attached to a plan.
type Risk struct {
	Level   RiskLevel
	Message string
}

// Change describes one intended mutation. Resource identifies the target,
// Action is a domain verb such as create/update/replace/write, and Details are
// optional human-readable reasons or changed fields.
type Change struct {
	Resource string
	Action   string
	Details  []string
}

// Plan is the common, read-only representation of an intended mutation.
// Domain packages remain responsible for computing it.
type Plan struct {
	Summary string
	Changes []Change
	Risks   []Risk
}

// Empty reports whether applying the plan would perform no declared changes.
func (p Plan) Empty() bool { return len(p.Changes) == 0 }

// Prepared couples a plan with the exact domain-owned value that apply will
// consume. Value may be a temporary workspace, patch set, rendered tree, or
// any other implementation-specific prepared state.
type Prepared[T any] struct {
	Plan  Plan
	Value T

	closeOnce sync.Once
	closeFn   func() error
	closeErr  error
}

// NewPrepared constructs prepared state with an optional cleanup function.
func NewPrepared[T any](plan Plan, value T, closeFn func() error) *Prepared[T] {
	return &Prepared[T]{Plan: plan, Value: value, closeFn: closeFn}
}

// Close releases prepared state exactly once. A nil cleanup function is valid.
func (p *Prepared[T]) Close() error {
	if p == nil {
		return nil
	}
	p.closeOnce.Do(func() {
		if p.closeFn != nil {
			p.closeErr = p.closeFn()
		}
	})
	return p.closeErr
}

// RenderOptions controls only presentation metadata. The plan itself remains
// free of terminal styling and UI-specific types.
type RenderOptions struct {
	Title   string
	Context []ui.Field
}

// PrintPlan renders a generic plan using the shared CharmCLI visual language.
func PrintPlan(w io.Writer, renderer render.Renderer, plan Plan, options RenderOptions) {
	title := options.Title
	if title == "" {
		title = "Plan"
	}
	ui.PrintHeader(w, renderer, ui.Header{Title: title, Context: options.Context})
	if plan.Summary != "" {
		ui.PrintStatusLine(w, renderer, ui.StatusLine{Level: ui.EventInfo, Label: "summary", Value: plan.Summary})
	}
	if len(plan.Changes) == 0 {
		ui.PrintStatusLine(w, renderer, ui.StatusLine{Level: ui.EventSuccess, Label: "changes", Value: "none"})
	} else {
		rows := make([]ui.DetailRow, 0, len(plan.Changes))
		for _, change := range plan.Changes {
			label := change.Resource
			if label == "" {
				label = "change"
			}
			rows = append(rows, ui.DetailRow{Level: ui.EventInfo, Label: label, Value: change.Action, Details: append([]string(nil), change.Details...)})
		}
		ui.PrintGroupedDetails(w, renderer, ui.GroupedDetails{Title: "Changes", Groups: []ui.DetailGroup{{Title: fmt.Sprintf("%d planned", len(rows)), Rows: rows}}})
	}
	if len(plan.Risks) > 0 {
		risks := make([]ui.Risk, 0, len(plan.Risks))
		for _, risk := range plan.Risks {
			level := ui.RiskInfo
			switch risk.Level {
			case RiskWarning:
				level = ui.RiskWarning
			case RiskDanger:
				level = ui.RiskBlocking
			}
			risks = append(risks, ui.Risk{Level: level, Message: risk.Message})
		}
		ui.PrintRisks(w, renderer, risks)
	}
}

// ConfirmOptions configures the shared confirmation immediately before apply.
type ConfirmOptions struct {
	Title       string
	Description string
	Bypass      bool
}

// ConfirmApply asks for confirmation when the plan is non-empty. Empty plans
// are accepted without prompting so callers can treat noop apply uniformly.
func ConfirmApply(ctx context.Context, runtime *charmcli.Runtime, plan Plan, options ConfirmOptions) (bool, error) {
	if plan.Empty() {
		return true, nil
	}
	return ui.Confirm(ctx, runtime, ui.ConfirmOptions{
		Title:       options.Title,
		Description: options.Description,
		Bypass:      options.Bypass,
	})
}
