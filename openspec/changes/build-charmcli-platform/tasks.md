# Tasks: charmcli platform

## 1. Repository and module foundation

- [ ] Initialize root module `github.com/sphr2k/charmcli` on current Go.
- [ ] Initialize nested module `github.com/sphr2k/charmcli/k8s`.
- [ ] Add `go.work` for both modules.
- [ ] Add CI that tests both modules independently in the workspace.
- [ ] Ensure root module has no `k8s.io/*` imports or requirements.

## 2. Core application/runtime

- [ ] Add injectable `Streams` and detected/overridable terminal `Capabilities`.
- [ ] Add `Runtime` as the shared command environment.
- [ ] Add `App` that executes an ordinary Cobra tree through Fang v2.
- [ ] Preserve direct access to Fang options rather than mirroring its full configuration API.
- [ ] Add optional error transform/redaction before Fang presentation.
- [ ] Add signal/context cancellation handling.

## 3. Exit and error contract

- [ ] Add `ExitCoder` support.
- [ ] Add typed `Usage`, `Failure`, and silent `Exit(code)` helpers.
- [ ] Map cancellation/user abort to exit 130.
- [ ] Ensure silent non-zero exits do not emit Fang error banners.
- [ ] Add tests for 0/1/2/130 behavior and transformed errors.

## 4. Output

- [ ] Add common output format type and values (`human`, `wide`, `json`, `yaml`, `name`).
- [ ] Add Cobra `-o/--output` binding and allowed-format validation.
- [ ] Add deterministic JSON/YAML encoders.
- [ ] Add callback-based human/wide/name rendering without reflection-driven domain modeling.
- [ ] Test that machine formats contain no ANSI and no progress/diagnostics.

## 5. Charm-native interaction

- [ ] Add Runtime-aware Huh v2 prompt environment using injected stdin/stderr.
- [ ] Add `Secret` helper with no-echo input and validation.
- [ ] Add boolean `Confirm` helper.
- [ ] Add `ConfirmExact` helper requiring an exact typed token/name.
- [ ] Add generic `Select` and `MultiSelect` helpers where a thin wrapper is useful.
- [ ] Reuse Huh ThemeCharm and accessible mode rather than creating a separate visual system.
- [ ] Reject required interaction in non-interactive mode unless explicitly bypassed/defaulted.
- [ ] Test user abort and exact-confirm behavior.

## 6. Progress/activity

- [ ] Add reusable activity/step execution for long-running CLI operations.
- [ ] Use Bubble Tea v2/Bubbles v2 for interactive terminal rendering.
- [ ] Add deterministic linear stderr fallback when not interactive.
- [ ] Ensure machine stdout is never touched by activity rendering.
- [ ] Do not add a wrapper framework for arbitrary full-screen Bubble Tea applications.

## 7. Semantic human rendering

- [ ] Add a small semantic renderer for title/accent/success/warning/error/muted/code presentation.
- [ ] Derive visual vocabulary from Charm-native theme primitives.
- [ ] Delegate color capability/downsampling and `NO_COLOR` behavior to Charm libraries.
- [ ] Avoid embedding consumer/domain tokens or status vocabularies in charmcli.

## 8. Kubernetes extension module

- [ ] Add `Scope` backed by the real `genericclioptions.ConfigFlags`.
- [ ] Pass upstream ConfigFlags directly onto Cobra commands.
- [ ] Expose `RESTClientGetter` and `RESTConfig` without a parallel config model.
- [ ] Add optional `-A/--all-namespaces` behavior.
- [ ] Add `-n`/`-A` mutual exclusion validation.
- [ ] Resolve current namespace through the upstream raw kubeconfig loader.
- [ ] Add reusable namespace completion plumbing.
- [ ] Keep controller-runtime out of the module.

## 9. kubectl invocation identity

- [ ] Add standalone/plugin invocation descriptor.
- [ ] Render `policyopsctl` vs `kubectl policyops` based on executable name.
- [ ] Render `resourcectl` vs `kubectl resources` based on executable name.
- [ ] Keep invocation identity presentation-only; command semantics must be identical.
- [ ] Add table-driven tests for standalone, plugin and unknown executable names.

## 10. Testkit and quality

- [ ] Add deterministic test runtime with buffers and explicit capabilities.
- [ ] Add representative examples for normal CLI, interactive prompt and Kubernetes CLI usage.
- [ ] Add gofmt/go vet/go test CI for root and k8s modules.
- [ ] Add a guard that fails if root `go.mod` or root source acquires Kubernetes dependencies.
- [ ] Document API stability as pre-1.0 until reference consumers have migrated.

## 11. Reference-consumer follow-up contracts

- [ ] `policyopsctl`: migrate from Fang v1 to charmcli/Fang v2 while preserving command behavior and kubectl alias equivalence.
- [ ] `resourcectl`: replace hand-written parser with Cobra/Fang/charmcli, make `get` canonical, keep `list` only as compatibility alias if desired, and retain `kubectl resources`.
- [ ] `homelabctl`: replace bespoke parser and manual terminal UX with charmcli while preserving strict command surface, exact destructive confirmation, redaction, and machine-output cleanliness.
- [ ] Remove consumer-local CLI infrastructure only after equivalent charmcli coverage exists; no UX/safety regression is accepted.
