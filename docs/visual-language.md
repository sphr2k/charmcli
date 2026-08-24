# charmcli visual language

> **v1.1** – calmer hierarchy, Title Case sections, light header accent, more breathing room.
> Colors continue to come exclusively from Huh `ThemeCharm`.

charmcli uses Charm as its terminal rendering substrate and shared color
language, but it does not adopt a Charm/Clack prompt layout for static command
output. The visual grammar is closer to modern system tools such as `delta`,
`bat`, and `duf`: dense, resource-oriented, quiet, and optimized for scanning
operational state.

This guide is normative for shared human-output primitives. Domain CLIs keep
ownership of their data and vocabulary; charmcli owns the presentation grammar.

## Principles

1. **Information first.** Decoration establishes hierarchy; it is never the
   content.
2. **Resource-oriented, not conversational.** `get` and `describe` should look
   like operator tools, not setup wizards.
3. **Dense but calm.** Prefer aligned fields and whitespace over repeated glyphs.
   A little vertical breathing room after the header and between sections is
   preferred over packing everything to the top.
4. **Color is semantic and shared.** Static output reuses Huh `ThemeCharm`
   color tokens so prompts, help, and result views remain one product family.
   Never introduce fixed hex palettes in consumers.
5. **Progress is transient.** Live activity belongs on stderr and resolves to a
   compact result. Static stdout remains stable.
6. **Responsive by construction.** Detail views become denser on wide terminals
   and remain readable in a narrow terminal.
7. **Boxes are exceptional.** Use a box only when it represents a real semantic
   boundary such as a warning, a destructive preflight summary, or an isolated
   artifact. Do not put every section in a card.

## Reference detail view (v1.1)

The canonical `describe` shape is:

```text
▌  homelab-node-1                                          ● READY
   esxi · pet · 192.168.200.1 · 10d12h

Server
────────────────────────────────────────────────────────────
power       poweredOn           ssh          ✓ reachable
ipv4        10.0.7.20

Host
────────────────────────────────────────────────────────────
uptime      10d19h              disk         68% · 25G
load        0.33 0.62 0.92      reboot       ⚠ required
cloud-init  ✓ done

Kubernetes
────────────────────────────────────────────────────────────
kubelet     v1.34.3             cpu          4
os          Ubuntu 24.04.4      memory       11.7 GiB
kernel      7.0.0-28            runtime      containerd 2.3.2

Workloads
────────────────────────────────────────────────────────────
pods        48                  drainable    34
daemonsets  10                  replace      ✓ ok

Network
────────────────────────────────────────────────────────────
Cilium      ● ready     23m
Kilo        ● ready     1d20h
```

The exact fields are domain-owned. The hierarchy is not.

## Resource header

A detail view starts with a light left accent, identity, current state, and
compact secondary metadata.

```text
▌  homelab-node-1                                          ● READY
   esxi · pet · 192.168.200.1 · 10d12h
```

Rules:

- one left accent glyph (`▌` or a thin `│`) rendered with the shared accent tone
  from `ThemeCharm`;
- resource identity is the strongest text on the screen (title tone);
- one state marker (`●`) is enough;
- status is right-aligned when the terminal is wide enough;
- secondary metadata is muted and separated with ` · `;
- one blank line after the header for breathing room;
- do not put the resource header in a box by default.

Use `render.ResourceHeaderLines` (v1.1 adds the optional left accent and spacing).

## Sections

Static sections use **Title Case** followed by a quiet rule:

```text
Host
────────────────────────────────────────────────────────────
```

Rules:

- Title Case (not `HOST` / ALL CAPS) for a calmer look;
- no `◇`/`│` timeline for normal static detail output;
- no repeated leading icon on each section;
- use whitespace between sections;
- the rule follows the available content width rather than a fixed global size.

Use `render.SectionLines` and `render.Rule`.

## Key/value details

Keys are secondary; values are primary. On a wide terminal, details prefer two
aligned columns. On a narrow terminal, or when the actual values would overflow,
they collapse to one column.

```text
uptime      10d19h              disk         68% · 25G
load        0.33 0.62 0.92      reboot       ⚠ required
```

Default maximum density:

```text
width < 80    one detail column
width >= 80   prefer two columns when content fits
```

Use `render.DetailGridForWidth` for normal terminal-aware output. Lower-level
`render.DetailColumns` and `render.DetailGrid` exist for callers that
intentionally control layout.

## Tables

`get` remains table-oriented. Tables should be compact and borderless by
default:

```text
NAME             STATE   PROVIDER   IP              AGE
homelab-node-1   Ready   esxi       192.168.200.1   10d
homelab-node-2   Ready   hcloud     10.0.0.12       32d
```

Rules:

- headers establish the structure; borders normally do not;
- stable resource identity should be the first column;
- `wide` adds high-signal columns rather than changing the underlying result;
- truncation must be deliberate and terminal-width aware;
- structured output never reuses the human table representation.

## Semantic tone

The shared tone vocabulary remains small. Its color tokens come from the Charm
family (`huh.ThemeCharm`) rather than from consumer-local palettes.

| Tone | Intended use |
| --- | --- |
| plain | normal values |
| accent/info | headings, selected identity, left accent, informational emphasis |
| success | healthy/ready/completed state |
| warning | attention required, degraded state, planned risky transition |
| error | failed/invalid/hard failure state |
| muted | keys, evidence, metadata, rules |

Domain state is never replaced by generic severity. Render `Drift`, `Ready`,
`Unmanaged`, or `poweredOn` as those values and apply the appropriate tone.

## Symbols

Symbols are scarce and semantic. Preferred vocabulary:

```text
●  state
✓  successful check / reachable / satisfied
⚠  warning / attention required
✕  failed check
→  transition
▌  header accent (v1.1)
```

Do not decorate every line. Do not introduce emoji into the canonical grammar.
A textual value must remain understandable if symbols or color are removed.

## Boxes and panels

Boxes are allowed, but not as the default section primitive. Good uses include:

- destructive-operation preflight summaries;
- a warning that must be visually isolated from normal evidence;
- a generated artifact or result that has an actual boundary;
- rare dashboard-like views whose contents are semantically independent cards.

Avoid "box soup": a `describe` command with `Server`, `Host`, `Kubernetes`,
`Workloads`, and `Network` each enclosed in its own border is usually harder to
scan than section rules and aligned details.

## Dynamic operations

Live operations use `ui.Activity` / `ui.RunSteps` on stderr. They may use
spinners, progress, and live updates while work is running. On completion they
resolve to a compact stable line.

```text
⠹ Draining homelab-node-1  17/34 pods
```

then:

```text
✓ Drained homelab-node-1  34 pods · 8.2s
```

This is deliberately a different grammar from static `describe` output.

## Terminal and machine behavior

- Lip Gloss v2 is the static rendering and layout substrate.
- Huh remains the interaction/form layer and its `ThemeCharm` styles provide
  the shared color-token source for static semantic tones.
- Huh's prompt structure does **not** define static `get`/`describe` layout.
- Bubble Tea/Bubbles remain the live interaction layer.
- Color-profile detection controls ANSI/color behavior.
- Human output must remain meaningful with no ANSI.
- JSON/YAML/name output is stable, ANSI-free machine data and does not contain
  these decorations.
- stdout contains requested results; prompts, progress, and errors use stderr.

## Anti-patterns

Do not use a Clack-style timeline as the universal layout:

```text
◇ Server
│ status  poweredOn
│
◇ Host
│ uptime  10d19h
```

Do not use borders merely because the renderer can draw them:

```text
╭─ Server ─╮  ╭─ SSH ─╮  ╭─ Host ─╮
│ ...      │  │ ...   │  │ ...    │
╰──────────╯  ╰───────╯  ╰────────╯
```

Do not give each consumer its own palette or equivalent local renderer once a
shared charmcli primitive exists.

## Implementation boundary

The intended stack is:

```text
domain result
    ↓
charmcli semantic primitives
    ↓
Huh ThemeCharm color tokens + Lip Gloss v2 layout
    ↓
terminal
```

Charm provides the coherent color and terminal substrate. charmcli owns the
static visual grammar and presentation contract.

## Changelog (v1.1)

- Resource header: optional left accent (`▌`) using accent tone + breathing room.
- Sections: Title Case instead of ALL CAPS.
- Symbols: prefer `⚠` for warnings (still falls back to plain text without color).
- Spacing: one blank line after the header and between major sections.
- Colors: unchanged – still derived only from `huh.ThemeCharm` via `render.Renderer`.
