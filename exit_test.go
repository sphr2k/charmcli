package charmcli

import (
	"context"
	"errors"
	"testing"

	"charm.land/huh/v2"
)

func TestExitCode(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want int
	}{
		{name: "success", err: nil, want: 0},
		{name: "execution failure", err: Failure(errors.New("boom")), want: 2},
		{name: "usage", err: Usage(errors.New("bad args")), want: 2},
		{name: "valid negative result", err: Exit(1), want: 1},
		{name: "silent custom", err: Exit(7), want: 7},
		{name: "cancelled", err: context.Canceled, want: 130},
		{name: "huh aborted", err: huh.ErrUserAborted, want: 130},
		{name: "plain execution error", err: errors.New("boom"), want: 2},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ExitCode(tt.err); got != tt.want {
				t.Fatalf("ExitCode() = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestExitZeroIsNil(t *testing.T) {
	if err := Exit(0); err != nil {
		t.Fatalf("Exit(0) = %v, want nil", err)
	}
}
