package main

import (
	"bufio"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/jcw/jeebus"
)

func ExampleMain() {
	fmt.Println("version", jeebus.Version)
	// Output:
	// version 0.3.0
}

// Compile and run main, and wait for it to report starting its HTTP server.
func TestRunMain(t *testing.T) {
	cmd := exec.Command("go")
	cmd.Args = append(cmd.Args, "run", "main.go")

	sout, err := cmd.StderrPipe()
	jeebus.Check(err)
	scanner := bufio.NewScanner(sout)

	err = cmd.Start()
	jeebus.Check(err)
	defer cmd.Wait()

	done := make(chan int)

	go func() {
		var pid int
		for scanner.Scan() {
			t := scanner.Text()
			if n := strings.Index(t, " pid "); n > 0 {
				pid, _ = strconv.Atoi(t[n+5:])
			}
			const startMsg = "starting HTTP server on http://localhost:3000"
			if pid > 0 && strings.Contains(t, startMsg) {
				done <- pid
			}
		}
	}()

	select {
	case <-time.After(5 * time.Second):
		// timeout must allow for the compile time as well!
		t.Errorf("HTTP server did not start")
		cmd.Process.Kill() // FIXME: this kills go, not the main process!
	case pid := <-done:
		syscall.Kill(pid, syscall.SIGINT)
	}
}
