package devfile

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/devfile/library/pkg/devfile/parser"
	"github.com/redhat-developer/odo/pkg/devfile"
	"github.com/redhat-developer/odo/pkg/devfile/location"
	"github.com/redhat-developer/odo/pkg/devfile/validate"
	"github.com/redhat-developer/odo/pkg/log"
	"github.com/redhat-developer/odo/pkg/odo/cmdline"
	"github.com/redhat-developer/odo/pkg/odo/genericclioptions"
	"github.com/redhat-developer/odo/pkg/odo/genericclioptions/clientset"
	"github.com/redhat-developer/odo/pkg/util"
	"github.com/spf13/cobra"
)

const RecommendedCommandName = "devfile"

type DevfileOptions struct {
	clientset  *clientset.Clientset
	devfileObj parser.DevfileObj
}

// NewAlizerOptions creates a new AlizerOptions instance
func NewDevfileOptions() *DevfileOptions {
	return &DevfileOptions{}
}

func (o *DevfileOptions) SetClientset(clientset *clientset.Clientset) {
	o.clientset = clientset
}

func (o *DevfileOptions) Complete(cmdline cmdline.Cmdline, args []string) (err error) {
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

func (o *DevfileOptions) Validate() error {
	return nil
}

// Run contains the logic for the odo command
func (o *DevfileOptions) Run() (err error) {
	if log.IsJSON() {
		b, err := json.Marshal(o.devfileObj.Data)
		if err != nil {
			return err
		}
		fmt.Printf("%v", string(b))

	}

	return nil
}
func NewCmdDevfile(name, fullName string) *cobra.Command {

	o := NewDevfileOptions()
	devfileCmd := &cobra.Command{
		Use:         name,
		Long:        "Get information about devfile",
		Args:        cobra.MaximumNArgs(0),
		Annotations: map[string]string{},
		Run: func(cmd *cobra.Command, args []string) {
			genericclioptions.GenericRun(o, cmd, args)
		},
	}

	devfileCmd.Annotations["machineoutput"] = "json"

	return devfileCmd
}
