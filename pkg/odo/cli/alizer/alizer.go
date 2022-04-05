package alizer

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/redhat-developer/odo/pkg/alizer"
	"github.com/redhat-developer/odo/pkg/log"
	"github.com/redhat-developer/odo/pkg/odo/cmdline"
	"github.com/redhat-developer/odo/pkg/odo/genericclioptions"
	"github.com/redhat-developer/odo/pkg/odo/genericclioptions/clientset"
	"github.com/spf13/cobra"
)

const RecommendedCommandName = "alizer"

type AlizerOptions struct {
	clientset *clientset.Clientset
}

// NewAlizerOptions creates a new AlizerOptions instance
func NewAlizerOptions() *AlizerOptions {
	return &AlizerOptions{}
}

func (o *AlizerOptions) SetClientset(clientset *clientset.Clientset) {
	o.clientset = clientset
}

func (o *AlizerOptions) Complete(cmdline cmdline.Cmdline, args []string) (err error) {
	return nil
}

func (o *AlizerOptions) Validate() error {
	return nil
}

// Run contains the logic for the odo command
func (o *AlizerOptions) Run(ctx context.Context) (err error) {
	if log.IsJSON() {
		cwd, err := os.Getwd()
		if err != nil {
			return err
		}
		df, reg, err := o.clientset.AlizerClient.DetectFramework(cwd)
		if err != nil {
			return err
		}
		result := alizer.GetDevfileLocationFromDetection(df, reg)
		b, err := json.Marshal(result)
		if err != nil {
			return err
		}
		fmt.Printf("%v", string(b))
	}

	return nil
}
func NewCmdAlizer(name, fullName string) *cobra.Command {

	o := NewAlizerOptions()
	alizerCmd := &cobra.Command{
		Use:         name,
		Long:        "Detect devfile to use based on files present in current directory",
		Args:        cobra.MaximumNArgs(0),
		Annotations: map[string]string{},
		Run: func(cmd *cobra.Command, args []string) {
			genericclioptions.GenericRun(o, cmd, args)
		},
	}
	clientset.Add(alizerCmd, clientset.ALIZER)

	alizerCmd.Annotations["machineoutput"] = "json"

	return alizerCmd
}
