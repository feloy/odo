//go:build !windows
// +build !windows

package helper

import (
	"io"
	"os/exec"

	"github.com/ActiveState/termtest/expect"
	"github.com/onsi/gomega/gexec"
)

func terminateProc(session *gexec.Session) error {
	session.Interrupt()
	return nil
}

func setSysProcAttr(command *exec.Cmd) {}

func startOnTerminal(console *expect.Console, command *exec.Cmd, outWriter io.Writer, errWriter io.Writer) (*gexec.Session, error) {
	command.Stdin = console.Tty()
	command.Stdout = console.Tty()
	command.Stderr = console.Tty()

	return gexec.Start(command, outWriter, errWriter)
}
