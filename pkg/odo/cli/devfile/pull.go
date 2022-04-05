package devfile

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
	"github.com/devfile/library/pkg/devfile/parser"
	"github.com/devfile/library/pkg/devfile/parser/data/v2/common"
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
	cwd            string
	clientset      *clientset.Clientset
	devfilePath    string
	devfileObj     parser.DevfileObj
	projectName    string
	starterProject v1alpha2.StarterProject
}

// NewAlizerOptions creates a new AlizerOptions instance
func NewPullOptions() *PullOptions {
	return &PullOptions{}
}

func (o *PullOptions) SetClientset(clientset *clientset.Clientset) {
	o.clientset = clientset
}

func (o *PullOptions) Complete(cmdline cmdline.Cmdline, args []string) (err error) {
	o.devfilePath = location.DevfileLocation("")
	isDevfile := util.CheckPathExists(o.devfilePath)
	if !isDevfile {
		return errors.New("no devfile found")
	}

	o.devfileObj, err = devfile.ParseAndValidateFromFile(o.devfilePath)
	if err != nil {
		return fmt.Errorf("failed to parse the devfile %s: %w", o.devfilePath, err)
	}
	err = validate.ValidateDevfileData(o.devfileObj.Data)
	if err != nil {
		return err
	}

	if len(args) != 1 {
		return errors.New("you should specify a starter project name")
	}
	o.projectName = args[0]

	o.cwd, err = os.Getwd()
	return err
}

func (o *PullOptions) Validate() error {
	projects, err := o.devfileObj.Data.GetStarterProjects(common.DevfileOptions{})
	if err != nil {
		return err
	}
	found := false
	for _, project := range projects {
		if project.Name == o.projectName {
			o.starterProject = project
			found = true
			break
		}
	}
	if !found {
		return fmt.Errorf("project %q not found in devfile", o.projectName)
	}
	return nil
}

// Run contains the logic for the odo command
func (o *PullOptions) Run(ctx context.Context) (err error) {
	err = o.clientset.InitClient.DownloadStarterProject(&o.starterProject, o.cwd)
	if err != nil {
		return err
	}
	if !util.CheckPathExists(o.devfilePath) {
		// devfile has been erased during starter project download
		err = o.devfileObj.WriteYamlDevfile()
		if err != nil {
			return err
		}
	} else {
		// devfile has been replaced or has not been erased during starter project download
		name := o.devfileObj.GetMetadataName()
		o.devfileObj, err = devfile.ParseAndValidateFromFile(o.devfilePath)
		if err != nil {
			return fmt.Errorf("failed to parse the devfile %s: %w", o.devfilePath, err)
		}
		err = validate.ValidateDevfileData(o.devfileObj.Data)
		if err != nil {
			return err
		}
		err = o.devfileObj.SetMetadataName(name)
		if err != nil {
			return err
		}
	}
	return nil
}

func NewCmdPull(name, fullName string) *cobra.Command {
	o := NewPullOptions()
	devfileCmd := &cobra.Command{
		Use:         name,
		Long:        "Pull a starter project defined in a devfile",
		Args:        cobra.ExactArgs(1),
		Annotations: map[string]string{},
		Run: func(cmd *cobra.Command, args []string) {
			genericclioptions.GenericRun(o, cmd, args)
		},
	}
	clientset.Add(devfileCmd, clientset.INIT)
	devfileCmd.Annotations["machineoutput"] = "json"
	return devfileCmd
}
