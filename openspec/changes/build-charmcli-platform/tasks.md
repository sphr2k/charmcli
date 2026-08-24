# Tasks: charmcli platform

## 1. Repository and module foundation

- [x] Initialize root module `github.com/sphr2k/charmcli` on current Go.
- [x] Initialize nested module `github.com/sphr2k/charmcli/k8s`.
- [x] Add `go.work` for both modules.
- [x] Add CI that tests both modules independently in the workspace.
- [x] Ensure root module has no `k8s.io/*` imports or requirements.

## 2. Core application/runtime

- [x] Add injectable `Streams` and detected/overridable terminal `Capabilities`.
- [x] Add `Runtime` as the shared command environment.
- [x] Add `App` that executes an ordinary Cobra tree through Fang v2.
- [x] Preserve direct access to Fang options rather than mirroring its full configuration API.
- [x] Add optional error transform/redaction before Fang presentation.
- [x] Add signal/context cancellation handling.

## 3. Exit and error contract

- [x] Add `ExitCoder` support.
- [x] Add typed `Usage`, `Failure`, and silent `Exit(code)` helpers.
- [x] Map cancellation/user abort to exit 130.
- [x] Ensure silent non-zero exits do not emit Fang error banners.
- [x] Add tests for 0/1/2/130 behavior and transformed errors.
- [x] Align public helpers/documentation with final convention: exit 1 = valid negative/domain result, exit 2 = invocation/execution failure.

## 4. Output

- [x] Add common output format type and values (`human`, `wide`, `json`, `yaml`, `name`).
- [x] Add Cobra `-o/--output` binding and allowed-format validation.
- [x] Add deterministic JSON/YAML encoders.
- [x] Add callback-based human/wide/name rendering without reflection-driven domain modeling.
- [x] Test that machine formats contain no ANSI and no progress/diagnostics.
- [ ] Add reusable `--no-headers` convention for tabular `get` commands.

## 5. Charm-native interaction

- [x] Add Runtime-aware Huh v2 prompt environment using injected stdin/stderr.
- [x] Add `Secret` helper with no-echo input and validation.
- [x] Add boolean `Confirm` helper.
- [x] Add `ConfirmExact` helper requiring an exact typed token/name.
- [x] Add generic `Select` and `MultiSelect` helpers where a thin wrapper is useful.
- [x] Reuse Huh ThemeCharm and accessible mode rather than creating a separate visual system.
- [x] Reject required interaction in non-interactive mode unless explicitly bypassed/defaulted.
- [x] Test user abort and exact-confirm behavior.

## 6. Progress/activity

- [x] Add reusable activity/step execution for long-running CLI operations.
- [x] Use Bubble Tea v2/Bubbles v2 for interactive terminal rendering.
- [x] Add deterministic linear stderr fallback when not interactive.
- [x] Ensure machine stdout is never touched by activity rendering.
- [x] Do not add a wrapper framework for arbitrary full-screen Bubble Tea applications.

## 7. Semantic human rendering

- [x] Add a small semantic renderer for title/accent/success/warning/error/muted/code presentation.
- [x] Derive visual vocabulary from Charm-native theme primitives.
- [x] Delegate color capability/downsampling and `NO_COLOR` behavior to Charm libraries.
- [x] Avoid embedding consumer/domain tokens or status vocabularies in charmcli.
- [ ] Extend shared renderer to cover consistent section, key/value, table-header, finding, evidence and next-command presentation.
- [ ] Add visual/golden contract tests for the shared human hierarchy.

## 8. Kubernetes extension module

- [x] Add `Scope` backed by the real `genericclioptions.ConfigFlags`.
- [x] Pass upstream ConfigFlags directly onto Cobra commands.
- [x] Expose `RESTClientGetter` and `RESTConfig` without a parallel config model.
- [x] Add optional `-A/--all-namespaces` behavior.
- [x] Add `-n`/`-A` mutual exclusion validation.
- [x] Resolve current namespace through the upstream raw kubeconfig loader.
- [x] Add reusable namespace completion plumbing.
- [x] Keep controller-runtime out of the module.

## 9. kubectl invocation identity

- [x] Add standalone/plugin invocation descriptor.
- [x] Render `policyopsctl` vs `kubectl policyops` based on executable name.
- [x] Render `resourcectl` vs `kubectl resources` based on executable name.
- [x] Keep invocation identity presentation-only; command semantics must be identical.
- [x] Add table-driven tests for standalone, plugin and unknown executable names.

## 10. Testkit and quality

- [x] Add deterministic test runtime with buffers and explicit capabilities.
- [x] Add representative examples for normal CLI and Kubernetes CLI usage.
- [x] Add gofmt/go vet/go test CI for root and k8s modules.
- [x] Add a guard that fails if root `go.mod` or root source acquires Kubernetes dependencies.
- [x] Document API stability as pre-1.0 until reference consumers have migrated.

## 11. Unified command-language helpers/contracts

- [ ] Add reusable convention helpers where they reduce drift without replacing Cobra.
- [ ] Standardize canonical resource grammar: plural nouns for `get`, singular nouns for `describe`/`apply`/`explain`/`suggest`.
- [ ] Ensure root and incomplete verb groups show help rather than execute implicit defaults.
- [ ] Standardize shared flag spellings: `-o/--output`, `-y/--yes`, `--no-color`, `--non-interactive`, `--timeout`, `--verbose`, `--debug`, `--dry-run`.
- [ ] Reserve `-v` for Fang/version semantics.
- [ ] Do not add mandatory `--quiet` in v1.
- [ ] Add convention tests proving flags are omitted rather than accepted-and-ignored when unsupported.

## 12. Apply and dry-run contract

- [ ] Add shared types/helpers for describing apply transitions without embedding consumer-specific transition logic.
- [ ] Define/test transition categories suitable for human/machine presentation (`noop`, `create`, `update`, `replace` or consumer extension).
- [ ] Ensure destructive confirmation happens only after a destructive transition is computed.
- [ ] Ensure `--yes` bypasses confirmation only.
- [ ] Ensure required confirmation in non-interactive mode fails before mutation.
- [ ] Add optional `--dry-run` binding with the hard contract: resolve + validate + compute + present, zero external mutation.
- [ ] Add tests proving dry-run never invokes registered mutation callbacks.

## 13. Unified output/visual contract tests

- [ ] Add contract fixtures for `get`, `describe`, `apply`, and `doctor` presentation.
- [ ] Verify machine output is ANSI-free under TTY and non-TTY capability configurations.
- [ ] Verify progress/debug/errors remain on stderr for all machine formats.
- [ ] Verify `wide` adds detail without changing the logical result.
- [ ] Verify `name` contains identity only.
- [ ] Verify domain states remain visible while mapping to shared severity.

## 14. policyopsctl migration contract

- [ ] Migrate from Fang v1 to charmcli/Fang v2.
- [ ] Preserve canonical verbs: `get`, `describe`, `explain`, `suggest`, `doctor`.
- [ ] Adopt shared flags, output, severity, human hierarchy, progress/error and exit contracts.
- [ ] Remove PolicyOps-local generic palette/hierarchy plumbing where charmcli provides equivalent primitives.
- [ ] Preserve typed PolicyOps result models and intentional manifest output for `suggest binding -o yaml|json`.
- [ ] Make `policyopsctl` and `kubectl policyops` use the same command tree and prove semantic parity in tests.
- [ ] Keep Kubernetes config/scope behavior native through `charmcli/k8s`.

## 15. resourcectl migration contract

- [ ] Rename canonical standalone product from `kubectl-resources` to `resourcectl`.
- [ ] Replace hand-written parser with Cobra/Fang/charmcli.
- [ ] Make `get` canonical and remove implicit list-on-empty behavior.
- [ ] Canonicalize resource-oriented `get`/`describe` grammar; keep `list` only as hidden/deprecated compatibility alias if retained.
- [ ] Adopt shared flags, output, severity, human hierarchy, error and exit contracts.
- [ ] Replace direct clientcmd plumbing with `charmcli/k8s` ConfigFlags passthrough.
- [ ] Keep `kubectl resources` as a presentation alias for the same command tree.
- [ ] Remove ResourceCLI-local generic style/TTY/parser plumbing after equivalent charmcli coverage exists.

## 16. homelabctl canonical command migration

- [ ] Replace bespoke parser with ordinary Cobra commands executed through charmcli/Fang v2.
- [ ] Canonicalize the public surface to:
  - [ ] `homelabctl get cluster`
  - [ ] `homelabctl get nodes`
  - [ ] `homelabctl describe cluster`
  - [ ] `homelabctl describe node NAME`
  - [ ] `homelabctl apply cluster`
  - [ ] `homelabctl apply node NAME`
  - [ ] `homelabctl doctor`
- [ ] Make `apply node` subsume create/update/replace/no-op behavior.
- [ ] Make `apply cluster` the canonical user-facing replacement for `cluster bootstrap` while preserving useful phase-selection controls.
- [ ] Split current node-status UX into concise `get nodes` and detailed `describe node` semantics.
- [ ] Define cluster `get`/`describe` result models and doctor scope without inventing synthetic success.
- [ ] Preserve pet/cattle lifecycle guards before mutation.

## 17. homelabctl interaction and visual migration

- [ ] Replace hand-built Clack-inspired prompt/progress code with Charm-native Huh/Bubble Tea/Bubbles/charmcli primitives.
- [ ] Preserve exact-name confirmation for destructive node replacement.
- [ ] Use `-y/--yes` as the canonical confirmation bypass.
- [ ] Add `--non-interactive` behavior and prove required unanswered interaction fails before mutation.
- [ ] Preserve Doppler secret no-echo behavior and environment fallback semantics.
- [ ] Preserve/strengthen redaction before Fang error presentation.
- [ ] Replace Homelab-local palette/spinner/tree infrastructure with shared human visual grammar where equivalent.
- [ ] Preserve qualitative UX: migration may redesign visuals but MUST NOT reduce clarity, safety, feedback or accessibility.

## 18. homelabctl apply/dry-run migration

- [ ] Refactor node application to determine transition before confirmation/mutation.
- [ ] Expose `--dry-run` for node apply only once full mutation-free transition planning is correct.
- [ ] Evaluate cluster apply for faithful dry-run; expose the flag only if all mutating phases can be prevented while meaningful planning remains possible.
- [ ] Test no-op/create/update/replace transition presentation.
- [ ] Test destructive replacement dry-run requires no confirmation and makes zero mutations.
- [ ] Test destructive real apply requires exact confirmation or `--yes`.

## 19. Compatibility and documentation

- [ ] Decide the bounded compatibility window for old Homelab command forms and resource `list`.
- [ ] If retained, implement old forms only as hidden/deprecated aliases to canonical commands.
- [ ] Remove deprecated syntax from canonical README/help/examples immediately.
- [ ] Add golden/visual tests for root help and every public verb group across all three CLIs.
- [ ] Ensure shared terminology and flag descriptions match across consumer help text.

## 20. Cross-consumer acceptance

- [ ] A user familiar with one reference CLI can predict the others' root/verb/resource grammar.
- [ ] Shared flags have identical spelling and behavior across all commands that expose them.
- [ ] Shared output formats have identical semantics.
- [ ] Shared severity and human hierarchy are visually consistent across all three CLIs.
- [ ] API/RBAC/provider/config execution failure maps to exit 2; valid negative diagnostic result maps to exit 1; abort maps to 130.
- [ ] Standalone and kubectl plugin aliases are semantically equivalent.
- [ ] No migrated consumer retains duplicate generic parser/TTY/palette/progress/Kubernetes-config infrastructure beside charmcli equivalents.
- [ ] No consumer migration weakens destructive-operation safety, machine-output cleanliness, accessibility or diagnostic clarity.
