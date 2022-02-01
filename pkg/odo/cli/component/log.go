package component

import (
	"fmt"

	projectCmd "github.com/redhat-developer/odo/pkg/odo/cli/project"
	"github.com/redhat-developer/odo/pkg/odo/cmdline"
	"github.com/redhat-developer/odo/pkg/odo/genericclioptions"
	"github.com/redhat-developer/odo/pkg/odo/util/completion"
	ktemplates "k8s.io/kubectl/pkg/util/templates"

	odoutil "github.com/redhat-developer/odo/pkg/odo/util"

	"github.com/spf13/cobra"
)

// LogRecommendedCommandName is the recommended watch command name
const LogRecommendedCommandName = "log"

var logExample = ktemplates.Examples(`  # Get the logs for the nodejs component
%[1]s nodejs
`)

// LogOptions contains log options
type LogOptions struct {
	// Component context
	*ComponentOptions

	// Flags
	followFlag  bool
	debugFlag   bool
	contextFlag string
}

// NewLogOptions returns new instance of LogOptions
func NewLogOptions() *LogOptions {
	return &LogOptions{
		ComponentOptions: &ComponentOptions{},
	}
}

// Complete completes log args
func (lo *LogOptions) Complete(cmdline cmdline.Cmdline, args []string) (err error) {
	lo.ComponentOptions.Context, err = genericclioptions.New(genericclioptions.NewCreateParameters(cmdline).NeedDevfile(lo.contextFlag))
	return err
}

// Validate validates the log parameters
func (lo *LogOptions) Validate() (err error) {
	return
}

// Run has the logic to perform the required actions as part of command
func (lo *LogOptions) Run() (err error) {
	err = lo.DevfileComponentLog()
	return
}

// NewCmdLog implements the log odo command
func NewCmdLog(name, fullName string) *cobra.Command {
	o := NewLogOptions()

	var logCmd = &cobra.Command{
		Use:         fmt.Sprintf("%s [component_name]", name),
		Short:       "Retrieve the log for the given component",
		Long:        `Retrieve the log for the given component`,
		Example:     fmt.Sprintf(logExample, fullName),
		Args:        cobra.RangeArgs(0, 1),
		Annotations: map[string]string{"command": "component"},
		Run: func(cmd *cobra.Command, args []string) {
			genericclioptions.GenericRun(o, cmd, args)
		},
	}

	logCmd.Flags().BoolVarP(&o.followFlag, "follow", "f", false, "Follow logs")
	logCmd.Flags().BoolVar(&o.debugFlag, "debug", false, "Show logs for debug command")

	logCmd.SetUsageTemplate(odoutil.CmdUsageTemplate)
	completion.RegisterCommandHandler(logCmd, completion.ComponentNameCompletionHandler)
	// Adding `--context` flag
	odoutil.AddContextFlag(logCmd, &o.contextFlag)

	//Adding `--project` flag
	projectCmd.AddProjectFlag(logCmd)

	return logCmd
}
