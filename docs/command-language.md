# Command language

CharmCLI consumers SHOULD use a stable verb-first command language for resource-oriented operations:

```text
<tool> <verb> <resource> [name] [flags]
```

The convention standardizes semantics, not domain models. A resource only exposes verbs that make sense for it.

## Canonical verbs

### `get`

Read-only concise state. Collections use plural resource names (`get nodes`, `get bindings`); singleton resources remain singular (`get cluster`, `get gitops`).

### `describe`

Read-only detailed state for one named resource or a singleton (`describe node worker-01`, `describe cluster`).

### `explain`

Read-only causal explanation of an existing state or decision.

### `suggest`

Read-only proposal derived from observed state. It MUST NOT persist desired state.

### `plan`

Computes the mutation that the corresponding `apply` would perform. Planning MAY read APIs, invoke resolvers, render and validate artifacts, and create temporary workspaces. It MUST NOT persist desired-state changes, modify provider or Kubernetes resources, create commits, push Git, or change secrets.

A plan SHOULD be reusable by `apply` when the domain can safely prepare exact state. `operation.Prepared[T]` exists for this purpose.

### `apply`

Converges desired state. `apply` MUST use the same planning semantics as `plan`, present the intended changes, and only then perform the mutation. Destructive domain operations may require stronger confirmation than the generic boolean confirmation helper.

### `doctor`

Read-only diagnostics across one or more resources. A non-zero exit status MAY represent hard findings rather than command execution failure when the consumer documents that contract.

## Naming

Use plural nouns for collections and singular nouns for one object:

```text
get nodes
describe node worker-01
get profiles
describe profile restricted
```

Singletons stay singular:

```text
get cluster
plan gitops
apply gitops
```

Do not turn actions into fake resources. Prefer `plan app-dependencies` over `plan update apps`.

## Utilities

Imperative utility namespaces do not have to implement this resource grammar. Consumers SHOULD NOT invent `plan`/`apply` semantics merely for symmetry. Such utilities can remain domain-specific until they acquire a real desired-state model.

## Cobra wiring

This document is a language contract, not a Cobra framework. CharmCLI intentionally does not provide a generic command registry or resource DSL. Consumers own their command trees. Shared wiring should only be extracted after multiple consumers demonstrate the same stable boilerplate.
