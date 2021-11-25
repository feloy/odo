package registry

import (
	// Built-in packages
	"fmt"

	// Third-party packages
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/zalando/go-keyring"
	ktemplates "k8s.io/kubectl/pkg/util/templates"

	// odo packages
	registryUtil "github.com/redhat-developer/odo/pkg/odo/cli/registry/util"
	"github.com/redhat-developer/odo/pkg/odo/genericclioptions"
	"github.com/redhat-developer/odo/pkg/preference"
	"github.com/redhat-developer/odo/pkg/util"
)

const updateCommandName = "update"

// "odo registry update" command description and examples
var (
	updateLongDesc = ktemplates.LongDesc(`Update devfile registry URL`)

	updateExample = ktemplates.Examples(`# Update devfile registry URL
	%[1]s CheRegistry https://che-devfile-registry-update.openshift.io
	`)
)

// UpdateOptions encapsulates the options for the "odo registry update" command
type UpdateOptions struct {
	// Parameters
	registryName string
	registryURL  string

	// Flags
	tokenFlag string
	forceFlag bool

	operation string
	user      string
}

// NewUpdateOptions creates a new UpdateOptions instance
func NewUpdateOptions() *UpdateOptions {
	return &UpdateOptions{}
}

// Complete completes UpdateOptions after they've been created
func (o *UpdateOptions) Complete(name string, cmd *cobra.Command, args []string) (err error) {
	o.operation = "update"
	o.registryName = args[0]
	o.registryURL = args[1]
	o.user = "default"
	return nil
}

// Validate validates the UpdateOptions based on completed values
func (o *UpdateOptions) Validate() (err error) {
	err = util.ValidateURL(o.registryURL)
	if err != nil {
		return err
	}
	if registryUtil.IsGitBasedRegistry(o.registryURL) {
		registryUtil.PrintGitRegistryDeprecationWarning()
	}
	return nil
}

// Run contains the logic for "odo registry update" command
func (o *UpdateOptions) Run(cmd *cobra.Command) (err error) {
	secureBeforeUpdate := false
	secureAfterUpdate := false

	secure, err := registryUtil.IsSecure(o.registryName)
	if err != nil {
		return err
	}

	if secure {
		secureBeforeUpdate = true
	}

	if o.tokenFlag != "" {
		secureAfterUpdate = true
	}

	cfg, err := preference.New()
	if err != nil {
		return errors.Wrap(err, "unable to update registry")
	}
	err = cfg.RegistryHandler(o.operation, o.registryName, o.registryURL, o.forceFlag, secureAfterUpdate)
	if err != nil {
		return err
	}

	if secureAfterUpdate {
		err = keyring.Set(util.CredentialPrefix+o.registryName, o.user, o.tokenFlag)
		if err != nil {
			return errors.Wrap(err, "unable to store registry credential to keyring")
		}
	} else if secureBeforeUpdate && !secureAfterUpdate {
		err = keyring.Delete(util.CredentialPrefix+o.registryName, o.user)
		if err != nil {
			return errors.Wrap(err, "unable to delete registry credential from keyring")
		}
	}

	return nil
}

// NewCmdUpdate implements the "odo registry update" command
func NewCmdUpdate(name, fullName string) *cobra.Command {
	o := NewUpdateOptions()
	registryUpdateCmd := &cobra.Command{
		Use:     fmt.Sprintf("%s <registry name> <registry URL>", name),
		Short:   updateLongDesc,
		Long:    updateLongDesc,
		Example: fmt.Sprintf(fmt.Sprint(updateExample), fullName),
		Args:    cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			genericclioptions.GenericRun(o, cmd, args)
		},
	}

	registryUpdateCmd.Flags().StringVar(&o.tokenFlag, "token", "", "Token to be used to access secure registry")
	registryUpdateCmd.Flags().BoolVarP(&o.forceFlag, "force", "f", false, "Don't ask for confirmation, update the registry directly")

	return registryUpdateCmd
}
