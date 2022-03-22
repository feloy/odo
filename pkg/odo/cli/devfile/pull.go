package devfile

import (
	"errors"
	"fmt"

	"github.com/devfile/library/pkg/devfile/parser"
	"github.com/redhat-developer/odo/pkg/devfile"
	"github.com/redhat-developer/odo/pkg/devfile/location"
	"github.com/redhat-developer/odo/pkg/devfile/validate"
	"github.com/redhat-developer/odo/pkg/odo/cmdline"
	"github.com/redhat-developer/odo/pkg/odo/genericclioptions"
	"github.com/redhat-developer/odo/pkg/odo/genericclioptions/clientset"
	"github.com/redhat-developer/odo/pkg/util"
	"github.com/spf13/cobra"
)

const PullRecommendedCommandName = "pull"

type PullOptions struct {
	clientset  *clientset.Clientset
	devfileObj parser.DevfileObj
}

// NewAlizerOptions creates a new AlizerOptions instance
func NewPullOptions() *PullOptions {
	return &PullOptions{}
}

func (o *PullOptions) SetClientset(clientset *clientset.Clientset) {
	o.clientset = clientset
}

func (o *PullOptions) Complete(cmdline cmdline.Cmdline, args []string) (err error) {
	devfilePath := location.DevfileLocation("")
	isDevfile := util.CheckPathExists(devfilePath)
	if !isDevfile {
		return errors.New("no devfile found")
	}

	o.devfileObj, err = devfile.ParseAndValidateFromFile(devfilePath)
	if err != nil {
		return fmt.Errorf("failed to parse the devfile %s: %w", devfilePath, err)
	}
	err = validate.ValidateDevfileData(o.devfileObj.Data)
	if err != nil {
		return err
	}

	return nil
}

func (o *PullOptions) Validate() error {
	return nil
}

// Run contains the logic for the odo command
func (o *PullOptions) Run() (err error) {
	return nil
}

func NewCmdPull(name, fullName string) *cobra.Command {
	o := NewPullOptions()
	devfileCmd := &cobra.Command{
		Use:         name,
		Long:        "Pull a starter project defined in a devfile",
		Args:        cobra.MaximumNArgs(0),
		Annotations: map[string]string{},
		Run: func(cmd *cobra.Command, args []string) {
			genericclioptions.GenericRun(o, cmd, args)
		},
	}
	return devfileCmd
}
