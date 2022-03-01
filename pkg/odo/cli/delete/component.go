package delete

import (
	"fmt"

	"github.com/spf13/cobra"

	componentlabels "github.com/redhat-developer/odo/pkg/component/labels"
	"github.com/redhat-developer/odo/pkg/odo/cmdline"
	"github.com/redhat-developer/odo/pkg/odo/genericclioptions"
	"github.com/redhat-developer/odo/pkg/odo/genericclioptions/clientset"
)

// ComponentRecommendedCommandName is the recommended component sub-command name
const ComponentRecommendedCommandName = "component"

type ComponentOptions struct {
	// name of the component to delete, optional
	name string

	// namespace on which to find the component to delete, optional, defaults to current namespace
	namespace string

	// forceFlag forces deletion
	forceFlag bool

	// Clients
	clientset *clientset.Clientset
}

// NewComponentOptions returns new instance of ComponentOptions
func NewComponentOptions() *ComponentOptions {
	return &ComponentOptions{}
}

func (o *ComponentOptions) SetClientset(clientset *clientset.Clientset) {
}

func (o *ComponentOptions) Complete(cmdline cmdline.Cmdline, args []string) (err error) {
	if o.namespace != "" {
		o.clientset.KubernetesClient.SetNamespace(o.namespace)
	}
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
	selector := componentlabels.GetSelector(o.name, "app")
	// TODO add managed-by=odo
	list, err := o.clientset.KubernetesClient.GetAllResourcesFromSelector(selector, o.clientset.KubernetesClient.GetCurrentNamespace())
	if err != nil {
		return err
	}
	for _, resource := range list {
		fmt.Printf("%s.%s\n", resource.GetKind(), resource.GetName())
	}
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
	componentCmd.Flags().StringVar(&o.namespace, "namespace", "", "Namespace in which to find the component to delete, optional. By default, the current namespace defined in kube config is used")
	componentCmd.Flags().BoolVarP(&o.forceFlag, "force", "f", false, "Delete component without prompting")
	clientset.Add(componentCmd, clientset.DELETE_COMPONENT, clientset.KUBERNETES)

	return componentCmd
}
