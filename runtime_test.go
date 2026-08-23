package charmcli

import (
	"bytes"
	"testing"
)

func TestRuntimeCapabilitiesCanBeOverridden(t *testing.T) {
	streams := Streams{In: &bytes.Buffer{}, Out: &bytes.Buffer{}, Err: &bytes.Buffer{}}
	runtime := NewRuntime(streams, WithCapabilities(Capabilities{Interactive: true, Accessible: true}))
	if !runtime.Capabilities.Interactive || !runtime.Capabilities.Accessible {
		t.Fatalf("capabilities = %#v", runtime.Capabilities)
	}
}

func TestRuntimeNormalizesNilStreams(t *testing.T) {
	runtime := NewRuntime(Streams{})
	if runtime.Streams.In == nil || runtime.Streams.Out == nil || runtime.Streams.Err == nil {
		t.Fatalf("streams not normalized: %#v", runtime.Streams)
	}
}
