package libdevfile

import (
	"context"

	"github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
	"github.com/devfile/library/v2/pkg/devfile/parser"
)

// execCommand is a command implementation for exec commands
type execCommand struct {
	command    v1alpha2.Command
	devfileObj parser.DevfileObj
}

var _ command = (*execCommand)(nil)

// newExecCommand creates a new execCommand instance, adapting the devfile-defined command to run in the target component's
// container, modifying it to add environment variables or adapting the path as needed.
func newExecCommand(devfileObj parser.DevfileObj, command v1alpha2.Command) *execCommand {
	return &execCommand{
		command:    command,
		devfileObj: devfileObj,
	}
}

func (o *execCommand) CheckValidity() error {
	return nil
}

func (o *execCommand) Execute(ctx context.Context, handler Handler) error {
	if o.isTerminating() {
		return handler.ExecuteTerminatingCommand(ctx, o.command)
	}
	return handler.ExecuteNonTerminatingCommand(ctx, o.command)
}

// isTerminating returns true if not Run or Debug command
func (o *execCommand) isTerminating() bool {
	if o.command.Exec.Group == nil {
		return true
	}
	kind := o.command.Exec.Group.Kind
	if kind == v1alpha2.RunCommandGroupKind || kind == v1alpha2.DebugCommandGroupKind {
		return false
	}
	return true
}
