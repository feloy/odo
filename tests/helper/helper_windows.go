//go:build windows
// +build windows

package helper

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"syscall"

	"github.com/ActiveState/termtest/expect"
	"github.com/onsi/gomega/gexec"
	"golang.org/x/sys/windows"
)

func terminateProc(session *gexec.Session) error {
	pid := session.Command.Process.Pid
	dll, err := windows.LoadDLL("kernel32.dll")
	if err != nil {
		return fmt.Errorf("loading DLL: %w", err)
	}
	defer dll.Release()
	generateConsoleCtrlEvent, err := dll.FindProc("GenerateConsoleCtrlEvent")
	if err != nil {
		return fmt.Errorf("finding GenerateConsoleCtrlEvent: %w", err)
	}
	r1, _, err := generateConsoleCtrlEvent.Call(uintptr(syscall.CTRL_BREAK_EVENT), uintptr(pid))
	if r1 == 0 {
		return fmt.Errorf("calling GenerateConsoleCtrlEvent: %w", err)
	}
	return nil
}

func setSysProcAttr(command *exec.Cmd) {
	command.SysProcAttr = &syscall.SysProcAttr{
		CreationFlags: syscall.CREATE_NEW_PROCESS_GROUP,
	}
}

func startOnTerminal(console *expect.Console, command *exec.Cmd, outWriter io.Writer, errWriter io.Writer) (*gexec.Session, error) {
	var argv []string
	if len(command.Args) > 0 {
		argv = command.Args
	} else {
		argv = []string{command.Path}
	}

	var envv []string
	if command.Env != nil {
		envv = command.Env
	} else {
		envv = os.Environ()
	}
	pid, _, err := console.Pty.Spawn(command.Path, argv, &syscall.ProcAttr{
		Dir: command.Dir,
		Env: envv,
	})
	if err != nil {
		return nil, fmt.Errorf("Failed to spawn process in terminal: %w", err)
	}

	// Let's pray that this always works.  Unfortunately we cannot create our process from a process handle.
	command.Process, err = os.FindProcess(pid)
	if err != nil {
		return nil, fmt.Errorf("Failed to create an os.Process struct: %w", err)
	}

	exited := make(chan struct{})

	session := &Session{
		Command:  command,
		Out:      gbytes.NewBuffer(),
		Err:      gbytes.NewBuffer(),
		Exited:   exited,
		lock:     &sync.Mutex{},
		exitCode: -1,
	}
	// return gexec.Start(command, outWriter, errWriter)
}
