package output

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/sphr2k/charmcli"
	"github.com/spf13/cobra"
	"go.yaml.in/yaml/v3"
)

// Format identifies one CLI result representation.
type Format string

const (
	Human Format = "human"
	Wide  Format = "wide"
	JSON  Format = "json"
	YAML  Format = "yaml"
	Name  Format = "name"
)

// Selector binds and validates a command output format.
type Selector struct {
	value        string
	defaultValue Format
	allowed      map[Format]struct{}
}

// NewSelector creates an output selector. If no formats are supplied, all
// standard formats are accepted.
func NewSelector(defaultFormat Format, formats ...Format) *Selector {
	if defaultFormat == "" {
		defaultFormat = Human
	}
	if len(formats) == 0 {
		formats = []Format{Human, Wide, JSON, YAML, Name}
	}
	allowed := make(map[Format]struct{}, len(formats))
	for _, format := range formats {
		allowed[format] = struct{}{}
	}
	return &Selector{value: string(defaultFormat), defaultValue: defaultFormat, allowed: allowed}
}

// AddFlags binds the conventional -o/--output flag.
func (s *Selector) AddFlags(cmd *cobra.Command) {
	cmd.Flags().StringVarP(&s.value, "output", "o", string(s.defaultValue), "Output format: "+strings.Join(s.allowedNames(), ", "))
}

// Format returns the selected validated format.
func (s *Selector) Format() (Format, error) {
	format := Format(strings.ToLower(strings.TrimSpace(s.value)))
	if _, ok := s.allowed[format]; !ok {
		return "", charmcli.Usage(fmt.Errorf("unsupported output format %q; expected one of %s", s.value, strings.Join(s.allowedNames(), ", ")))
	}
	return format, nil
}

func (s *Selector) allowedNames() []string {
	order := []Format{Human, Wide, JSON, YAML, Name}
	result := make([]string, 0, len(s.allowed))
	for _, format := range order {
		if _, ok := s.allowed[format]; ok {
			result = append(result, string(format))
		}
	}
	return result
}

// Renderers contains optional domain-owned renderers for non-structured
// formats. JSON and YAML are encoded generically from the value.
type Renderers[T any] struct {
	Human func(io.Writer, T) error
	Wide  func(io.Writer, T) error
	Name  func(io.Writer, T) error
}

// Write writes one selected result representation.
func Write[T any](w io.Writer, format Format, value T, renderers Renderers[T]) error {
	switch format {
	case JSON:
		encoder := json.NewEncoder(w)
		encoder.SetIndent("", "  ")
		return encoder.Encode(value)
	case YAML:
		data, err := yaml.Marshal(value)
		if err != nil {
			return fmt.Errorf("encode yaml: %w", err)
		}
		if _, err := w.Write(data); err != nil {
			return fmt.Errorf("write yaml: %w", err)
		}
		return nil
	case Human:
		return callRenderer("human", w, value, renderers.Human)
	case Wide:
		return callRenderer("wide", w, value, renderers.Wide)
	case Name:
		return callRenderer("name", w, value, renderers.Name)
	default:
		return fmt.Errorf("unsupported output format %q", format)
	}
}

func callRenderer[T any](name string, w io.Writer, value T, renderer func(io.Writer, T) error) error {
	if renderer == nil {
		return fmt.Errorf("%s output renderer is not configured", name)
	}
	return renderer(w, value)
}
