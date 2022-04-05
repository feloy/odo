package registry

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/redhat-developer/odo/pkg/log"
	"github.com/redhat-developer/odo/pkg/odo/cmdline"
	"github.com/redhat-developer/odo/pkg/odo/genericclioptions"
	"github.com/redhat-developer/odo/pkg/odo/genericclioptions/clientset"
	"github.com/spf13/cobra"
)

const RecommendedCommandName = "registry"

type RegistryOptions struct {
	clientset *clientset.Clientset
}

// NewRegistryOptions creates a new RegistryOptions instance
func NewRegistryOptions() *RegistryOptions {
	return &RegistryOptions{}
}

func (o *RegistryOptions) SetClientset(clientset *clientset.Clientset) {
	o.clientset = clientset
}

func (o *RegistryOptions) Complete(cmdline cmdline.Cmdline, args []string) (err error) {
	return nil
}

func (o *RegistryOptions) Validate() error {
	return nil
}

// Run contains the logic for the odo command
func (o *RegistryOptions) Run(ctx context.Context) (err error) {

	devfileEntries, _ := o.clientset.RegistryClient.ListDevfileStacks("")
	if log.IsJSON() {
		b, err := json.Marshal(devfileEntries.Items)
		if err != nil {
			return err
		}
		fmt.Printf("%v", string(b))
	}

	return nil
}
func NewCmdRegistry(name, fullName string) *cobra.Command {

	o := NewRegistryOptions()
	registryCmd := &cobra.Command{
		Use:         name,
		Long:        "List devfile in registries",
		Args:        cobra.MaximumNArgs(0),
		Annotations: map[string]string{},
		Run: func(cmd *cobra.Command, args []string) {
			genericclioptions.GenericRun(o, cmd, args)
		},
	}
	clientset.Add(registryCmd, clientset.REGISTRY)

	registryCmd.Annotations["machineoutput"] = "json"

	return registryCmd
}
