package component

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/klog"
	ktemplates "k8s.io/kubectl/pkg/util/templates"

	"github.com/redhat-developer/odo/pkg/labels"
	"github.com/redhat-developer/odo/pkg/log"
	clierrors "github.com/redhat-developer/odo/pkg/odo/cli/errors"
	"github.com/redhat-developer/odo/pkg/odo/cli/files"
	"github.com/redhat-developer/odo/pkg/odo/cli/ui"
	"github.com/redhat-developer/odo/pkg/odo/cmdline"
	odocontext "github.com/redhat-developer/odo/pkg/odo/context"
	"github.com/redhat-developer/odo/pkg/odo/genericclioptions"
	"github.com/redhat-developer/odo/pkg/odo/genericclioptions/clientset"
	"github.com/redhat-developer/odo/pkg/testingutil/filesystem"
)

// ComponentRecommendedCommandName is the recommended component sub-command name
const ComponentRecommendedCommandName = "component"

var deleteExample = ktemplates.Examples(`
# Delete the component present in the current directory from the cluster
%[1]s

# Delete the component named 'frontend' in the currently active namespace from the cluster
%[1]s --name frontend

# Delete the component named 'frontend' in the 'myproject' namespace from the cluster
%[1]s --name frontend --namespace myproject
`)

type ComponentOptions struct {
	// name of the component to delete, optional
	name string

	// namespace on which to find the component to delete, optional, defaults to current namespace
	namespace string

	// withFilesFlag controls whether files generated by odo should be deleted as well.
	withFilesFlag bool

	// forceFlag forces deletion
	forceFlag bool

	// waitFlag waits for deletion of all resources
	waitFlag bool

	// Clients
	clientset *clientset.Clientset
}

var _ genericclioptions.Runnable = (*ComponentOptions)(nil)

// NewComponentOptions returns new instance of ComponentOptions
func NewComponentOptions() *ComponentOptions {
	return &ComponentOptions{}
}

func (o *ComponentOptions) SetClientset(clientset *clientset.Clientset) {
	o.clientset = clientset
}

func (o *ComponentOptions) Complete(ctx context.Context, cmdline cmdline.Cmdline, args []string) (err error) {
	// 1. Name is not passed, and odo has access to devfile.yaml; Name is not passed so we assume that odo has access to the devfile.yaml
	if o.name == "" {
		devfileObj := odocontext.GetDevfileObj(ctx)
		if devfileObj == nil {
			return genericclioptions.NewNoDevfileError(odocontext.GetWorkingDirectory(ctx))
		}
		return nil
	}
	// 2. Name is passed, and odo does not have access to devfile.yaml; if Name is passed, then we assume that odo does not have access to the devfile.yaml
	if o.namespace != "" {
		o.clientset.KubernetesClient.SetNamespace(o.namespace)
	} else {
		o.namespace = o.clientset.KubernetesClient.GetCurrentNamespace()
	}
	return nil
}

func (o *ComponentOptions) Validate(ctx context.Context) error {
	if o.withFilesFlag && o.name != "" {
		return errors.New("'--files' cannot be used with '--name'; '--files' must be used from a directory containing a Devfile")
	}
	return nil
}

func (o *ComponentOptions) Run(ctx context.Context) error {
	if o.name != "" {
		return o.deleteNamedComponent(ctx)
	}
	return o.deleteDevfileComponent(ctx)
}

// deleteNamedComponent deletes a component given its name
func (o *ComponentOptions) deleteNamedComponent(ctx context.Context) error {
	log.Info("Searching resources to delete, please wait...")
	list, err := o.clientset.DeleteClient.ListClusterResourcesToDelete(ctx, o.name, o.namespace)
	if err != nil {
		return err
	}
	if len(list) == 0 {
		log.Infof("No resource found for component %q in namespace %q\n", o.name, o.namespace)
		return nil
	}
	printDevfileComponents(o.name, o.namespace, list)
	if o.forceFlag || ui.Proceed("Are you sure you want to delete these resources?") {
		failed := o.clientset.DeleteClient.DeleteResources(list, o.waitFlag)
		for _, fail := range failed {
			log.Warningf("Failed to delete the %q resource: %s\n", fail.GetKind(), fail.GetName())
		}
		log.Infof("The component %q is successfully deleted from namespace %q", o.name, o.namespace)
		return nil
	}

	log.Error("Aborting deletion of component")
	return nil
}

// deleteDevfileComponent deletes all the components defined by the devfile in the current directory
// devfileObj in context must not be nil when this method is called
func (o *ComponentOptions) deleteDevfileComponent(ctx context.Context) error {
	var (
		devfileObj    = odocontext.GetDevfileObj(ctx)
		componentName = odocontext.GetComponentName(ctx)
		appName       = odocontext.GetApplication(ctx)
		namespace     string
	)

	log.Info("Searching resources to delete, please wait...")
	isInnerLoopDeployed, devfileResources, err := o.clientset.DeleteClient.ListResourcesToDeleteFromDevfile(*devfileObj, appName, componentName, labels.ComponentAnyMode)
	if err != nil {
		if clierrors.AsWarning(err) {
			log.Warning(err.Error())
		} else {
			return err
		}
	}

	var hasClusterResources bool
	if o.clientset.KubernetesClient != nil {
		namespace = odocontext.GetNamespace(ctx)
		hasClusterResources = len(devfileResources) != 0
		if hasClusterResources {
			// Print all the resources that odo will attempt to delete
			printDevfileComponents(componentName, namespace, devfileResources)
		} else {
			log.Infof("No resource found for component %q in namespace %q\n", componentName, namespace)
			if !o.withFilesFlag {
				return nil
			}
		}
	}

	var filesToDelete []string
	if o.withFilesFlag {
		filesToDelete, err = getFilesCreatedByOdo(o.clientset.FS, ctx)
		if err != nil {
			return err
		}
		printFileCreatedByOdo(filesToDelete, hasClusterResources)
	}
	hasFilesToDelete := len(filesToDelete) != 0

	if !(hasClusterResources || hasFilesToDelete) {
		klog.V(2).Info("no cluster resources and no files to delete")
		return nil
	}

	if o.forceFlag || ui.Proceed(fmt.Sprintf("Are you sure you want to delete %q and all its resources?", componentName)) {
		if hasClusterResources {
			// Get a list of component's resources present on the cluster
			clusterResources, _ := o.clientset.DeleteClient.ListClusterResourcesToDelete(ctx, componentName, namespace)
			// Get a list of component's resources absent from the devfile, but present on the cluster
			remainingResources := listResourcesMissingFromDevfilePresentOnCluster(componentName, devfileResources, clusterResources)

			// if innerloop deployment resource is present, then execute preStop events
			if isInnerLoopDeployed {
				err = o.clientset.DeleteClient.ExecutePreStopEvents(*devfileObj, appName, componentName)
				if err != nil {
					log.Errorf("Failed to execute preStop events: %v", err)
				}
			}

			// delete all the resources
			failed := o.clientset.DeleteClient.DeleteResources(devfileResources, o.waitFlag)
			for _, fail := range failed {
				log.Warningf("Failed to delete the %q resource: %s\n", fail.GetKind(), fail.GetName())
			}
			log.Infof("The component %q is successfully deleted from namespace %q\n", componentName, namespace)

			if len(remainingResources) != 0 {
				log.Printf("There are still resources left in the cluster that might be belonging to the deleted component.")
				for _, resource := range remainingResources {
					fmt.Printf("\t- %s: %s\n", resource.GetKind(), resource.GetName())
				}
				log.Infof("If you want to delete those, execute `odo delete component --name %s --namespace %s`\n", componentName, namespace)
			}
		}

		if o.withFilesFlag {
			//Delete files
			remainingFiles := o.deleteFilesCreatedByOdo(o.clientset.FS, filesToDelete)
			var listOfFiles []string
			for f, e := range remainingFiles {
				log.Warningf("Failed to delete file or directory: %s: %v\n", f, e)
				listOfFiles = append(listOfFiles, "\t- "+f)
			}
			if len(remainingFiles) != 0 {
				log.Printf("There are still files or directories that could not be deleted.")
				fmt.Println(strings.Join(listOfFiles, "\n"))
				log.Info("You need to manually delete those.")
			}
		}

		return nil
	}

	log.Error("Aborting deletion of component")

	return nil
}

// listResourcesMissingFromDevfilePresentOnCluster returns a list of resources belonging to a component name that are present on cluster, but missing from devfile
func listResourcesMissingFromDevfilePresentOnCluster(componentName string, devfileResources, clusterResources []unstructured.Unstructured) []unstructured.Unstructured {
	var remainingResources []unstructured.Unstructured
	// get resources present in k8sResources(present on the cluster) but not in devfileResources(not present in the devfile)
	for _, k8sresource := range clusterResources {
		var present bool
		for _, dresource := range devfileResources {
			//  skip if the cluster and devfile resource are same OR if the cluster resource is the component's Endpoints resource
			if reflect.DeepEqual(dresource, k8sresource) || (k8sresource.GetKind() == "Endpoints" && strings.Contains(k8sresource.GetName(), componentName)) {
				present = true
				break
			}
		}
		if !present {
			remainingResources = append(remainingResources, k8sresource)
		}
	}
	return remainingResources
}

// printDevfileResources prints the devfile components for ComponentOptions.deleteDevfileComponent
func printDevfileComponents(componentName, namespace string, k8sResources []unstructured.Unstructured) {
	log.Infof("This will delete %q from the namespace %q.", componentName, namespace)

	if len(k8sResources) != 0 {
		log.Printf("The component contains the following resources that will get deleted:")
		for _, resource := range k8sResources {
			fmt.Printf("\t- %s: %s\n", resource.GetKind(), resource.GetName())
		}
	}
	fmt.Println()
}

// getFilesCreatedByOdo gets the list of all files that were initially created by odo.
func getFilesCreatedByOdo(filesys filesystem.Filesystem, ctx context.Context) ([]string, error) {
	workingDir := odocontext.GetWorkingDirectory(ctx)
	filesToDelete, err := files.GetFilesGeneratedByOdo(filesys, workingDir)
	if err != nil {
		return nil, err
	}

	var list []string
	for _, f := range filesToDelete {
		if _, err = filesys.Stat(f); errors.Is(err, fs.ErrNotExist) {
			continue
		}
		absPath := f
		if !filepath.IsAbs(f) {
			absPath = filepath.Join(workingDir, f)
		}
		list = append(list, absPath)
	}

	return list, nil
}

func printFileCreatedByOdo(files []string, hasClusterResources bool) {
	if len(files) == 0 {
		return
	}

	m := "This will "
	if hasClusterResources {
		m += "also "
	}
	log.Info(m + "delete the following files and directories:")
	for _, f := range files {
		fmt.Println("\t- " + f)
	}
}

// deleteFilesCreatedByOdo deletes all the files that were created initially by odo.
// It returns a slice of files that could not be deleted.
func (o *ComponentOptions) deleteFilesCreatedByOdo(filesys filesystem.Filesystem, files []string) (notDeleted map[string]error) {
	notDeleted = make(map[string]error)
	for _, f := range files {
		err := filesys.RemoveAll(f)
		if err != nil {
			notDeleted[f] = err
		}
	}
	return notDeleted
}

// NewCmdComponent implements the component odo sub-command
func NewCmdComponent(name, fullName string) *cobra.Command {
	o := NewComponentOptions()

	var componentCmd = &cobra.Command{
		Use:     name,
		Short:   "Delete component",
		Long:    "Delete component",
		Args:    genericclioptions.NoArgsAndSilenceJSON,
		Example: fmt.Sprintf(deleteExample, fullName),
		RunE: func(cmd *cobra.Command, args []string) error {
			return genericclioptions.GenericRun(o, cmd, args)
		},
	}
	componentCmd.Flags().StringVar(&o.name, "name", "", "Name of the component to delete, optional. By default, the component described in the local devfile is deleted")
	componentCmd.Flags().StringVar(&o.namespace, "namespace", "", "Namespace in which to find the component to delete, optional. By default, the current namespace defined in kubeconfig is used")
	componentCmd.Flags().BoolVarP(&o.withFilesFlag, "files", "", false, "Delete all files and directories generated by odo. Use with caution.")
	componentCmd.Flags().BoolVarP(&o.forceFlag, "force", "f", false, "Delete component without prompting")
	componentCmd.Flags().BoolVarP(&o.waitFlag, "wait", "w", false, "Wait for deletion of all dependent resources")
	clientset.Add(componentCmd, clientset.DELETE_COMPONENT, clientset.KUBERNETES, clientset.FILESYSTEM)

	return componentCmd
}
