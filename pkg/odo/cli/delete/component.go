package delete

import (
	"github.com/spf13/cobra"

	"github.com/redhat-developer/odo/pkg/odo/cmdline"
	"github.com/redhat-developer/odo/pkg/odo/genericclioptions"
	"github.com/redhat-developer/odo/pkg/odo/genericclioptions/clientset"
)

// ComponentRecommendedCommandName is the recommended component sub-command name
const ComponentRecommendedCommandName = "component"

type ComponentOptions struct {
	// name of the component to delete, optional
	name string
}

// NewComponentOptions returns new instance of ComponentOptions
func NewComponentOptions() *ComponentOptions {
	return &ComponentOptions{}
}

func (o *ComponentOptions) SetClientset(clientset *clientset.Clientset) {
}

func (o *ComponentOptions) Complete(cmdline cmdline.Cmdline, args []string) (err error) {
	return nil
}

func (o *ComponentOptions) Validate() (err error) {
	return nil

}

func (o *ComponentOptions) Run() error {
	if o.name != "" {
		return o.deleteNamedComponent()
	}
	return o.deleteDevfileComponent()
}

// deleteNamedComponent deletes a component given its name
func (o *ComponentOptions) deleteNamedComponent() error {
	return nil
}

// deleteDevfileComponent deletes a component defined by the devfile in the current directory
func (o *ComponentOptions) deleteDevfileComponent() error {
	return nil
}

// NewCmdComponent implements the component odo sub-command
func NewCmdComponent(name, fullName string) *cobra.Command {
	o := NewComponentOptions()

	var componentCmd = &cobra.Command{
		Use:   name,
		Short: "Delete component",
		Long:  "Delete component",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			genericclioptions.GenericRun(o, cmd, args)
		},
	}
	componentCmd.Flags().StringVar(&o.name, "name", "", "Name of the component to delete, optional. By default, the component described in the local devfile is deleted")

	return componentCmd
}
