// Package testkit contains deterministic runtime fixtures for charmcli tests
// and consumers.
package testkit

import (
	"bytes"

	"github.com/sphr2k/charmcli"
)

// Harness owns in-memory streams and an explicitly configured Runtime.
type Harness struct {
	In      *bytes.Buffer
	Out     *bytes.Buffer
	Err     *bytes.Buffer
	Runtime *charmcli.Runtime
}

// New returns a harness with explicit capabilities.
func New(capabilities charmcli.Capabilities) *Harness {
	in := &bytes.Buffer{}
	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	runtime := charmcli.NewRuntime(
		charmcli.Streams{In: in, Out: out, Err: errOut},
		charmcli.WithCapabilities(capabilities),
	)
	return &Harness{In: in, Out: out, Err: errOut, Runtime: runtime}
}
