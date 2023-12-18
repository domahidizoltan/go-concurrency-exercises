//////////////////////////////////////////////////////////////////////
//
// Given is a mock process which runs indefinitely and blocks the
// program. Right now the only way to stop the program is to send a
// SIGINT (Ctrl-C). Killing a process like that is not graceful, so we
// want to try to gracefully stop the process first.
//
// Change the program to do the following:
//   1. On SIGINT try to gracefully stop the process using
//          `proc.Stop()`
//   2. If SIGINT is called again, just kill the program (last resort)
//

package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	shutdownSignal := make(chan os.Signal, 2)
	signal.Notify(shutdownSignal, syscall.SIGINT)

	// Create a process
	proc := MockProcess{}

	// Run the process (blocking)
	go proc.Run()

	<-shutdownSignal
	go proc.Stop()

	<-shutdownSignal
	fmt.Printf("\nKilling process!\n")
	os.Exit(1)
}
