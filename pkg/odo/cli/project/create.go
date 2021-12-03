package project

import (
	"fmt"

	"github.com/redhat-developer/odo/pkg/log"
	"github.com/redhat-developer/odo/pkg/machineoutput"
	"github.com/redhat-developer/odo/pkg/odo/genericclioptions"
	"github.com/redhat-developer/odo/pkg/project"
	scontext "github.com/redhat-developer/odo/pkg/segment/context"

	"github.com/spf13/cobra"

	ktemplates "k8s.io/kubectl/pkg/util/templates"
)

const createRecommendedCommandName = "create"

var (
	createExample = ktemplates.Examples(`
	# Create a new project
	%[1]s myproject
	`)

	createLongDesc = ktemplates.LongDesc(`Create a new project.
	This command directly performs actions on the cluster and doesn't require a push.
	`)

	createShortDesc = `Create a new project`
)

// ProjectCreateOptions encapsulates the options for the odo project create command
type ProjectCreateOptions struct {
	// Context
	*genericclioptions.Context

	// Parameters
	projectName string

	// Flags
	waitFlag bool
}

// NewProjectCreateOptions creates a ProjectCreateOptions instance
func NewProjectCreateOptions() *ProjectCreateOptions {
	return &ProjectCreateOptions{}
}

// Complete completes ProjectCreateOptions after they've been created
func (pco *ProjectCreateOptions) Complete(name string, cmd *cobra.Command, args []string) (err error) {
	pco.projectName = args[0]
	pco.Context, err = genericclioptions.New(genericclioptions.NewCreateParameters(cmd))
	return err
}

// Validate validates the parameters of the ProjectCreateOptions
func (pco *ProjectCreateOptions) Validate() error {
	return nil
}

// Run runs the project create command
func (pco *ProjectCreateOptions) Run(cmd *cobra.Command) (err error) {
	if scontext.GetTelemetryStatus(cmd.Context()) {
		scontext.SetClusterType(cmd.Context(), pco.KClient)
	}

	// Create the "spinner"
	s := &log.Status{}

	// If the --wait parameter has been passed, we add a spinner..
	if pco.waitFlag {
		s = log.Spinner("Waiting for project to come up")
		defer s.End(false)
	}

	// Create the project & end the spinner (if there is any..)
	err = project.Create(pco.Context.KClient, pco.projectName, pco.waitFlag)
	if err != nil {
		return err
	}
	s.End(true)

	successMessage := fmt.Sprintf(`Project %q is ready for use`, pco.projectName)
	log.Successf(successMessage)

	// Set the current project when created
	err = project.SetCurrent(pco.Context.KClient, pco.projectName)
	if err != nil {
		return err
	}

	log.Successf("New project created and now using project: %v", pco.projectName)

	// If -o json has been passed, let's output the appropriate json output.
	if log.IsJSON() {
		prj := project.NewProject(pco.projectName, true)
		machineoutput.OutputSuccess(prj)
	}

	return nil
}

// NewCmdProjectCreate creates the project create command
func NewCmdProjectCreate(name, fullName string) *cobra.Command {
	o := NewProjectCreateOptions()

	projectCreateCmd := &cobra.Command{
		Use:         name,
		Short:       createShortDesc,
		Long:        createLongDesc,
		Example:     fmt.Sprintf(createExample, fullName),
		Args:        cobra.ExactArgs(1),
		Annotations: map[string]string{"machineoutput": "json"},
		Run: func(cmd *cobra.Command, args []string) {
			genericclioptions.GenericRun(o, cmd, args)
		},
	}

	projectCreateCmd.Flags().BoolVarP(&o.waitFlag, "wait", "w", false, "Wait until the project is ready")
	return projectCreateCmd
}
