package output

import (
	"bytes"
	"io"
	"strings"
	"testing"
)

type fixture struct {
	Name string `json:"name" yaml:"name"`
}

func TestWriteJSON(t *testing.T) {
	var out bytes.Buffer
	if err := Write(&out, JSON, fixture{Name: "demo"}, Renderers[fixture]{}); err != nil {
		t.Fatal(err)
	}
	if strings.Contains(out.String(), "\x1b[") || !strings.Contains(out.String(), `"name": "demo"`) {
		t.Fatalf("json=%q", out.String())
	}
}

func TestWriteYAML(t *testing.T) {
	var out bytes.Buffer
	if err := Write(&out, YAML, fixture{Name: "demo"}, Renderers[fixture]{}); err != nil {
		t.Fatal(err)
	}
	if got := out.String(); got != "name: demo\n" {
		t.Fatalf("yaml=%q", got)
	}
}

func TestHumanRendererRemainsDomainOwned(t *testing.T) {
	var out bytes.Buffer
	err := Write(&out, Human, fixture{Name: "demo"}, Renderers[fixture]{
		Human: func(w io.Writer, value fixture) error {
			_, err := w.Write([]byte("human:" + value.Name))
			return err
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if out.String() != "human:demo" {
		t.Fatalf("out=%q", out.String())
	}
}
