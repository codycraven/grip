// Package grip provides abstraction for signal handling.
package grip

import (
	"fmt"
	"io"
	"os"
	"os/signal"
)

// A SignalHandler receives a signal and does something with it.
type SignalHandler func(os.Signal)

// An ExitHandler performs actions and returns an error if a problem occurs.
type ExitHandler func() error

// Trap listens for provided os.Signals and executes a SignalHandler callback
// function when one is received.
func Trap(fn SignalHandler, s ...os.Signal) {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, s...)
	go func() {
		fn(<-ch)
	}()
}

// Message creates a SignalHandler that writes to an io.Writer and then chains
// to another SignalHandler.
//
//	grip.Trap(
//		grip.Message("received shutdown request", os.Stdout, func(_ os.Signal) {
//			fmt.Println("our signal handler")
//		}),
//		syscall.SIGINT, syscall.SIGTERM,
//	)
func Message(m string, w io.Writer, fn SignalHandler) SignalHandler {
	return func(s os.Signal) {
		fmt.Fprintf(w, "%s: %s\n", m, s)
		fn(s)
	}
}

// Exit creates a SignalHandler that passes exit codes to a channel.
//
// Exit calls each provided ExitHandler. For each ExitHandler, the integer sent
// to the channel is incremented in a base-2 manner so that when receiving the
// exit code you can determine which ExitHandler(s) failed.
//
// If you receive 6 as your exit code then you can determine which step failed
// based on bitmasking:
//
//  - 1 & 6 == 0 -> first ExitHandler passed
//  - 2 & 6 == 1 -> second ExitHandler failed
//  - 4 & 6 == 1 -> third ExitHandler failed
//  - 8 & 6 == 0 -> fourth ExitHandler passed
//
// A complete example:
//
//	package main
//
//	import (
//		"fmt"
//		"os"
//		"syscall"
//
//		"github.com/codycraven/grip"
//	)
//
//	func main() {
//		// Create a channel to receive the exit code.
//		ch := make(chan int)
//		grip.Trap(
//			// Passing to Message isn't needed, added just to show its use.
//			grip.Message("received shutdown request", os.Stdout, grip.Exit(
//				// The channel for Exit to send the exit code to.
//				ch,
//				// The io.Writer for Exit to write received errors to.
//				os.Stderr,
//				// ExitHandlers that should run (in sequential order).
//				func() error {
//					// A graceful shutdown step, if it returned an error it
//					// would add 1 to the exit code.
//					return nil
//				},
//				func() error {
//					// Another graceful shutdown step.
//					return fmt.Errorf("since this is the second Exit handler it'll add 2")
//				},
//				func() error {
//					// Yet another graceful shutdown step.
//					return fmt.Errorf("since this is the third Exit handler it'll add 4")
//				},
//				func() error {
//					// Our last graceful shutdown step, if it returned an error
//					// it would add 8 to the exit code.
//					return nil
//				},
//			)),
//			// Listen for these signals:
//			syscall.SIGINT,
//			syscall.SIGTERM,
//		)
//		// Wait on channel to receive int before exiting.
//		os.Exit(<-ch)
//	}
func Exit(ch chan int, errWriter io.Writer, fn ...ExitHandler) SignalHandler {
	return func(s os.Signal) {
		exit := 0
		errBit := 1
		for _, f := range fn {
			err := f()
			if err != nil {
				exit += errBit
				fmt.Fprintf(errWriter, "added %d to exit code for error: %s\n", errBit, err)
			}
			errBit *= 2
		}
		ch <- exit
	}
}
