# grip

A signal handling library for Go.

[![Go Reference](https://pkg.go.dev/badge/github.com/codycraven/grip.svg)](https://pkg.go.dev/github.com/codycraven/grip)

## Quickstart

Add grip to the modules in your project:

```go
package main

import (
	"fmt"
	"os"
	"syscall"

	"github.com/codycraven/grip"
)

func main() {
	// Create a channel to receive the exit code.
	ch := make(chan int)
	grip.Trap(
		// Passing to Message isn't needed, added just to show its use.
		grip.Message("received shutdown request", os.Stdout, grip.Exit(
			// The channel for Exit to send the exit code to.
			ch,
			// The io.Writer for Exit to write received errors to.
			os.Stderr,
			// ExitHandlers that should run (in sequential order).
			func() error {
				// A graceful shutdown step, if it returned an error it
				// would add 1 to the exit code.
				return nil
			},
			func() error {
				// Another graceful shutdown step.
				return fmt.Errorf("since this is the second Exit handler it'll add 2")
			},
			func() error {
				// Yet another graceful shutdown step.
				return fmt.Errorf("since this is the third Exit handler it'll add 4")
			},
			func() error {
				// Our last graceful shutdown step, if it returned an error
				// it would add 8 to the exit code.
				return nil
			},
		)),
		// Listen for these signals:
		syscall.SIGINT,
		syscall.SIGTERM,
	)
	// Wait on channel to receive int before exiting.
	os.Exit(<-ch)
}
```

## Installation

Install grip with the `go get` command:

```bash
go get -u github.com/codycraven/grip
```
