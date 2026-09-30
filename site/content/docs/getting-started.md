---
title: "Getting Started"
weight: 10
---

Install:

```shell
go get github.com/nfx/go-tui
```

Minimal dropdown:

```go
package main

import (
	"log"

	tui "github.com/nfx/go-tui"
)

func main() {
	choice, err := tui.Dropdown("Select environment", []string{"dev", "staging", "prod"})
	if err != nil {
		log.Fatal(err)
	}
	log.Println("selected:", choice)
}
```

Most widgets accept shared options such as `WithContext`, `WithTimeout`, `WithInput`, and `WithOutput`.
Start with [Shared Options and Helpers](options-and-helpers.md), then see each widget page.
