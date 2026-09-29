---
title: "Input and Password"
weight: 40
---

APIs:

- `Input(label, opts...)`
- `Password(label, opts...)`

## Basic example

```go
value, err := tui.Input("Project name")
password, err := tui.Password("API token")
```

`Password` masks entered text and does not print an answer line.

## WithDefault

Sets initial text and cursor position.

```go
value, err := tui.Input("Project name",
	tui.WithDefault("demo-service"),
)
```

## WithContext

Sets interaction context.

```go
value, err := tui.Input("Project name",
	tui.WithContext(ctx),
)
```

## WithTimeout

Stops interaction when timeout elapses.

```go
value, err := tui.Input("Project name",
	tui.WithTimeout(20*time.Second),
)
```

## WithInput

Overrides input source.

```go
value, err := tui.Input("Project name",
	tui.WithInput(in),
)
```

## WithOutput

Overrides output destination.

```go
value, err := tui.Input("Project name",
	tui.WithOutput(out),
)
```

## WithOptions

Bundles common options.

```go
common := tui.WithOptions(
	tui.WithTimeout(20*time.Second),
	tui.WithOutput(os.Stderr),
)

value, err := tui.Input("Project name", common)
```

## Notes

- Arrow-left/right and backspace editing are supported.
- Paste buffers are handled as one input event.
- Timeout/cancel returns context error.
