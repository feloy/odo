//go:build !windows
// +build !windows

package helper

import (
	"io"
	"os/exec"

	"github.com/onsi/gomega/gexec"
)

func terminateProc(session *gexec.Session) error {
	session.Interrupt()
	return nil
}

func setSysProcAttr(command *exec.Cmd) {}

func startOnTerminal(command *exec.Cmd, outWriter io.Writer, errWriter io.Writer) (*gexec.Session, error) {
	return gexec.Start(command, outWriter, errWriter)
}
