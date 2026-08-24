# charmcli

Opinionated Go CLI infrastructure for a consistent modern command experience built on Charm without replacing Cobra.

`charmcli` combines the current Charm v2 stack around a small set of cross-CLI contracts:

- Cobra stays the command model.
- Fang v2 owns normal help, errors, version, completion and manpage UX.
- Huh v2 owns prompts and forms.
- Bubble Tea v2/Bubbles v2 own live interaction and activity rendering.
- Lip Gloss v2 is the rendering substrate for charmcli's own static human visual language.
- stdout is the requested result; prompts/progress/errors go to stderr.
- exit status is `0` success, `1` an explicitly rendered valid-negative/domain result, `2` invalid invocation or execution failure, and `130` cancellation/user abort.

The project is intentionally pre-1.0 while the reference consumers (`homelabctl`, `policyopsctl`, `resourcectl`) validate the API.

## Core

```go
package main

import (
    "context"
    "os"

    "github.com/sphr2k/charmcli"
    "github.com/spf13/cobra"
)

func main() {
    app := charmcli.New(charmcli.WithVersion("dev"))
    code := app.Run(context.Background(), os.Args[1:], charmcli.DefaultStreams(), func(rt *charmcli.Runtime) *cobra.Command {
        return &cobra.Command{
            Use: "example",
            RunE: func(cmd *cobra.Command, _ []string) error {
                _, err := cmd.OutOrStdout().Write([]byte("hello\n"))
                return err
            },
        }
    })
    os.Exit(code)
}
```

Normal returned errors are execution failures and therefore exit 2. Use the helpers when intent should be explicit:

```go
return charmcli.Failure(err)   // execution failure, exit 2
return charmcli.Usage(err)     // invalid invocation, exit 2
return charmcli.Exit(1)        // valid negative/domain result already rendered; silent exit 1
```

Cobra flag parsing, unknown subcommands and positional `Args` validators are classified as usage automatically. `context.Canceled` and Huh user aborts exit 130 without an additional Fang error banner.

## Interaction

Small interactive workflows use Huh with native Charm styling:

```go
err := ui.ConfirmExact(ctx, rt, ui.ConfirmExactOptions{
    Title:    "Replace worker-03?",
    Expected: "worker-03",
    Bypass:   yes,
})
```

`ui.Activity` and `ui.RunSteps` use Bubble Tea/Bubbles on an interactive terminal and deterministic line-oriented stderr output otherwise.

## Output

The `output` package standardizes `-o/--output` and machine encoders while leaving domain result models domain-owned:

```go
selector := output.NewSelector(output.Human, output.Human, output.JSON, output.YAML)
selector.AddFlags(cmd)
```

JSON/YAML/name are deterministic, ANSI-free result representations. Static human views use charmcli's shared visual grammar rather than Huh/Clack-style prompt timelines. The canonical shape is resource header + uppercase section/rule + dense aligned details, with semantic color and responsive one/two-column layout. List-style `get` commands use compact borderless tables.

See [`docs/visual-language.md`](docs/visual-language.md).

## Kubernetes extension

Kubernetes support is a separate Go module:

```text
github.com/sphr2k/charmcli/k8s
```

Core therefore carries no `k8s.io/*` dependency.

The extension is deliberately kubectl-native. `Scope` is backed by the real `genericclioptions.ConfigFlags`, passes its flags through directly, and exposes upstream REST interfaces instead of inventing another Kubernetes configuration model.

```go
scope := k8s.NewScope(k8s.WithAllNamespaces())
scope.AddFlags(cmd.PersistentFlags())

resolved, err := scope.Resolve()
```

Standalone and kubectl-plugin names can share one command tree:

```go
inv := k8s.Invocation{Standalone: "resourcectl", Plugin: "resources"}
use := inv.DisplayName(os.Args[0])
// resourcectl or kubectl resources
```

## Design

The active OpenSpec is under `openspec/changes/build-charmcli-platform/`.
