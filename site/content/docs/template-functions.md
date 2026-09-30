---
title: "Template Functions"
weight: 25
---

This page lists predefined helper functions available in widget templates.

## Foreground colors

- `black`
- `red`
- `green`
- `yellow`
- `blue`
- `purple`
- `cyan`
- `white`

Example:

```go
tui.WithLabelTemplate(`{{ green "Select:" }} {{ . | bold }}`)
```

## Bright foreground colors

- `brightBlack`
- `brightRed`
- `brightGreen`
- `brightYellow`
- `brightBlue`
- `brightPurple`
- `brightCyan`
- `brightWhite`

Example:

```go
tui.WithActiveItemTemplate(`{{ brightCyan "> " (label .) }}`)
```

## Background colors

- `bgBlack`
- `bgRed`
- `bgGreen`
- `bgYellow`
- `bgBlue`
- `bgPurple`
- `bgCyan`
- `bgWhite`

Example:

```go
tui.WithActiveItemTemplate(`{{ bgBlue " " }} {{ label . }}`)
```

## Text styles

- `bold`
- `dim`
- `italic`
- `underline`
- `strike`

Example:

```go
tui.WithAnswerTemplate(`{{ dim "selected:" }} {{ bold (label .Answer) }}`)
```

## Time formatter

- `ago`

Renders relative time such as `12sec ago`, `3hr ago`, or `in 2min`.

Example:

```go
tui.WithActiveItemTemplate(`{{ .Name }} {{ dim (ago .UpdatedAt) }}`)
```

## Dropdown-specific helper

Dropdown templates also expose:

- `label`

`label` resolves a display label using tags, common field names, `fmt.Stringer`, then `fmt.Sprint`.

Example:

```go
tui.WithActiveItemTemplate(`{{ green "> " (label .) }}`)
```

## Adding custom functions

Use `WithFn(name, fn)` to register your own global helper.

```go
tui.WithFn("upper", strings.ToUpper)
```
