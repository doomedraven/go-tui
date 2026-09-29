---
title: "Pretty JSON"
weight: 90
---

API:

- `PrettyJSON(w, src)`

`PrettyJSON` indents JSON and applies depth-aware ANSI coloring when writing to a terminal.

## Input types

- `string`
- `[]byte`
- Any Go value marshalable by `encoding/json`

## Example

```go
payload := map[string]any{
	"id": 42,
	"name": "service-a",
	"tags": []string{"prod", "api"},
}

if err := tui.PrettyJSON(os.Stdout, payload); err != nil {
	return err
}
```

## Behavior

- Non-terminal output: plain indented JSON.
- Terminal output: colored keys/values by nesting depth.
- Invalid JSON input returns an error from the indent/marshal stage.
