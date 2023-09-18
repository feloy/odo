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
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/klog"
	ktemplates "k8s.io/kubectl/pkg/util/templates"

	"github.com/redhat-developer/odo/pkg/api"
	"github.com/redhat-developer/odo/pkg/labels"
	"github.com/redhat-developer/odo/pkg/log"
	clierrors "github.com/redhat-developer/odo/pkg/odo/cli/errors"
	"github.com/redhat-developer/odo/pkg/odo/cli/feature"
	"github.com/redhat-developer/odo/pkg/odo/cli/files"
	"github.com/redhat-developer/odo/pkg/odo/cli/ui"
	"github.com/redhat-developer/odo/pkg/odo/cmdline"
	"github.com/redhat-developer/odo/pkg/odo/commonflags"
	fcontext "github.com/redhat-developer/odo/pkg/odo/commonflags/context"
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

	// runningInFlag limits the scope of deletion to resources created for the specified running mode.
	runningInFlag string

	// runningIn translates runningInFlag into a usable label that indicates which running mode we should consider
	// when listing resources candidate for deletion.
	// It can be either Dev, Deploy or Any (using constant labels.Component*Mode).
	runningIn string

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

func (o *ComponentOptions) UseDevfile(ctx context.Context, cmdline cmdline.Cmdline, args []string) bool {
	return o.name == ""
}

func (o *ComponentOptions) Complete(ctx context.Context, cmdline cmdline.Cmdline, args []string) (err error) {
	switch api.RunningMode(o.runningInFlag) {
	case api.RunningModeDev:
		o.runningIn = labels.ComponentDevMode
	case api.RunningModeDeploy:
		o.runningIn = labels.ComponentDeployMode
	case "":
		o.runningIn = labels.ComponentAnyMode
	default:
		return fmt.Errorf("invalid value for --running-in: %q. Acceptable values are: %s, %s",
			o.runningInFlag, api.RunningModeDev, api.RunningModeDeploy)
	}

	// Limit access to platforms if necessary
	if !feature.IsEnabled(ctx, feature.GenericPlatformFlag) {
		o.clientset.PodmanClient = nil
	}
	switch fcontext.GetPlatform(ctx, "") {
	case commonflags.PlatformCluster:
		o.clientset.PodmanClient = nil
	case commonflags.PlatformPodman:
		o.clientset.KubernetesClient = nil
	}

	// 1. Name is not passed, and odo has access to devfile.yaml; Name is not passed so we assume that odo has access to the devfile.yaml
	if o.name == "" {
		devfileObj := odocontext.GetEffectiveDevfileObj(ctx)
		if devfileObj == nil {
			return genericclioptions.NewNoDevfileError(odocontext.GetWorkingDirectory(ctx))
		}
		return nil
	}
	// 2. Name is passed, and odo does not have access to devfile.yaml; if Name is passed, then we assume that odo does not have access to the devfile.yaml
	if o.clientset.KubernetesClient != nil {
		if o.namespace != "" {
			o.clientset.KubernetesClient.SetNamespace(o.namespace)
		} else {
			o.namespace = o.clientset.KubernetesClient.GetCurrentNamespace()
		}
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
	remainingResources, err := o.deleteDevfileComponent(ctx)
	if err == nil {
		o.printRemainingResources(ctx, remainingResources)
	}
	return err
}

// deleteNamedComponent deletes a component given its name
func (o *ComponentOptions) deleteNamedComponent(ctx context.Context) error {
	var (
		appName = odocontext.GetApplication(ctx)

		clusterResources []unstructured.Unstructured
		podmanResources  []*corev1.Pod
		err              error
	)
	log.Finfof(o.clientset.Stdout, "Searching resources to delete, please wait...")
	if o.clientset.KubernetesClient != nil {
		clusterResources, err = o.clientset.DeleteClient.ListClusterResourcesToDelete(ctx, o.name, o.namespace, o.runningIn)
		if err != nil {
			return err
		}
	}

	if o.clientset.PodmanClient != nil {
		_, podmanResources, err = o.clientset.DeleteClient.ListPodmanResourcesToDelete(appName, o.name, o.runningIn)
		if err != nil {
			return err
		}
	}

	if len(clusterResources) == 0 && len(podmanResources) == 0 {
		log.Finfof(o.clientset.Stdout, messageWithPlatforms(
			o.clientset.KubernetesClient != nil,
			o.clientset.PodmanClient != nil,
			o.name, o.namespace,
		))
		return nil
	}
	o.printDevfileComponents(o.name, o.namespace, clusterResources, podmanResources)

	proceed := o.forceFlag
	if !proceed {
		proceed, err = ui.Proceed("Are you sure you want to delete these resources?")
		if err != nil {
			return err
		}
	}
	if proceed {

		if len(clusterResources) > 0 {
			spinner := log.Fspinnerf(o.clientset.Stdout, "Deleting resources from cluster")
			failed := o.clientset.DeleteClient.DeleteResources(clusterResources, o.waitFlag)
			for _, fail := range failed {
				log.Fwarningf(o.clientset.Stderr, "Failed to delete the %q resource: %s\n", fail.GetKind(), fail.GetName())
			}
			spinner.End(true)
			successMsg := fmt.Sprintf("The component %q is successfully deleted from namespace %q", o.name, o.namespace)
			if o.runningIn != "" {
				successMsg = fmt.Sprintf("The component %q running in the %s mode is successfully deleted from namespace %q", o.name, o.runningIn, o.namespace)
			}
			log.Finfof(o.clientset.Stdout, successMsg)
		}

		if len(podmanResources) > 0 {
			spinner := log.Fspinnerf(o.clientset.Stdout, "Deleting resources from podman")
			for _, pod := range podmanResources {
				err = o.clientset.PodmanClient.CleanupPodResources(pod, true)
				if err != nil {
					log.Fwarningf(o.clientset.Stderr, "Failed to delete the pod %q from podman: %s\n", pod.GetName(), err)
				}
			}
			spinner.End(true)
			successMsg := fmt.Sprintf("The component %q is successfully deleted from podman", o.name)
			if o.runningIn != "" {
				successMsg = fmt.Sprintf("The component %q running in the %s mode is successfully deleted podman", o.name, o.runningIn)
			}
			log.Finfof(o.clientset.Stdout, successMsg)
		}

		return nil
	}

	log.Ferror(o.clientset.Stderr, "Aborting deletion of component")
	return nil
}

func messageWithPlatforms(cluster, podman bool, name, namespace string) string {
	details := []string{}
	if cluster {
		details = append(details, fmt.Sprintf(" in namespace %q", namespace))
	}
	if podman {
		details = append(details, " on podman")
	}
	return fmt.Sprintf("No resource found for component %q%s\n", name, strings.Join(details, " or"))
}

// printRemainingResources lists the remaining cluster resources that are not found in the devfile.
func (o *ComponentOptions) printRemainingResources(ctx context.Context, remainingResources []unstructured.Unstructured) {
	if len(remainingResources) == 0 {
		return
	}
	componentName := odocontext.GetComponentName(ctx)
	namespace := odocontext.GetNamespace(ctx)
	log.Fprintf(o.clientset.Stdout, "There are still resources left in the cluster that might be belonging to the deleted component.")
	for _, resource := range remainingResources {
		fmt.Fprintf(o.clientset.Stdout, "\t- %s: %s\n", resource.GetKind(), resource.GetName())
	}
	log.Finfof(o.clientset.Stdout, "If you want to delete those, execute `odo delete component --name %s --namespace %s`\n", componentName, namespace)
}

// deleteDevfileComponent deletes all the components defined by the devfile in the current directory
// devfileObj in context must not be nil when this method is called
func (o *ComponentOptions) deleteDevfileComponent(ctx context.Context) ([]unstructured.Unstructured, error) {
	var (
		devfileObj    = odocontext.GetEffectiveDevfileObj(ctx)
		componentName = odocontext.GetComponentName(ctx)
		appName       = odocontext.GetApplication(ctx)

		namespace                  string
		isClusterInnerLoopDeployed bool
		hasClusterResources        bool
		clusterResources           []unstructured.Unstructured
		remainingResources         []unstructured.Unstructured

		isPodmanInnerLoopDeployed bool
		hasPodmanResources        bool
		podmanPods                []*corev1.Pod

		err error
	)

	log.Finfof(o.clientset.Stdout, "Searching resources to delete, please wait...")

	if o.clientset.KubernetesClient != nil {
		isClusterInnerLoopDeployed, clusterResources, err = o.clientset.DeleteClient.ListClusterResourcesToDeleteFromDevfile(
			*devfileObj, appName, componentName, o.runningIn)
		if err != nil {
			if clierrors.AsWarning(err) {
				log.Fwarning(o.clientset.Stderr, err.Error())
			} else {
				return nil, err
			}
		}

		namespace = odocontext.GetNamespace(ctx)
		hasClusterResources = len(clusterResources) != 0
		// Get a list of component's resources present on the cluster
		deployedResources, _ := o.clientset.DeleteClient.ListClusterResourcesToDelete(ctx, componentName, namespace, o.runningIn)
		// Get a list of component's resources absent from the devfile, but present on the cluster
		remainingResources = listResourcesMissingFromDevfilePresentOnCluster(componentName, clusterResources, deployedResources)
	}

	// 2. get podman resources
	if o.clientset.PodmanClient != nil {
		isPodmanInnerLoopDeployed, podmanPods, err = o.clientset.DeleteClient.ListPodmanResourcesToDelete(appName, componentName, o.runningIn)
		if err != nil {
			if clierrors.AsWarning(err) {
				log.Fwarning(o.clientset.Stderr, err.Error())
			} else {
				return nil, err
			}
		}
		hasPodmanResources = len(podmanPods) != 0
	}

	if !(hasClusterResources || hasPodmanResources) {
		log.Finfof(o.clientset.Stdout, messageWithPlatforms(o.clientset.KubernetesClient != nil, o.clientset.PodmanClient != nil, componentName, namespace))
		if !o.withFilesFlag {
			// check for resources here
			return remainingResources, nil
		}
	}

	o.printDevfileComponents(componentName, namespace, clusterResources, podmanPods)

	var filesToDelete []string
	if o.withFilesFlag {
		filesToDelete, err = getFilesCreatedByOdo(o.clientset.FS, ctx)
		if err != nil {
			return nil, err
		}
	}

	orphans, err := o.getOrphanDevstateFiles(o.clientset.FS, ctx)
	if err != nil {
		return nil, err
	}
	filesToDelete = append(filesToDelete, orphans...)

	hasFilesToDelete := len(filesToDelete) != 0

	if hasFilesToDelete {
		o.printFileCreatedByOdo(filesToDelete, hasClusterResources)
	}

	if !(hasClusterResources || hasPodmanResources || hasFilesToDelete) {
		klog.V(2).Info("no cluster resources and no files to delete")
		return remainingResources, nil
	}

	msg := fmt.Sprintf("Are you sure you want to delete %q and all its resources?", componentName)
	if o.runningIn != "" {
		msg = fmt.Sprintf("Are you sure you want to delete %q and all its resources running in the %s mode?", componentName, o.runningIn)
	}
	proceed := o.forceFlag
	if !proceed {
		proceed, err = ui.Proceed(msg)
		if err != nil {
			return nil, err
		}
	}
	if proceed {

		if hasClusterResources {
			spinner := log.Fspinnerf(o.clientset.Stdout, "Deleting resources from cluster")

			// if innerloop deployment resource is present, then execute preStop events
			if isClusterInnerLoopDeployed {
				err = o.clientset.DeleteClient.ExecutePreStopEvents(ctx, *devfileObj, appName, componentName)
				if err != nil {
					log.Ferrorf(o.clientset.Stderr, "Failed to execute preStop events: %v", err)
				}
			}

			// delete all the resources
			failed := o.clientset.DeleteClient.DeleteResources(clusterResources, o.waitFlag)
			for _, fail := range failed {
				log.Fwarningf(o.clientset.Stderr, "Failed to delete the %q resource: %s\n", fail.GetKind(), fail.GetName())
			}

			spinner.End(true)
			log.Finfof(o.clientset.Stdout, "The component %q is successfully deleted from namespace %q\n", componentName, namespace)

		}

		if hasPodmanResources {
			spinner := log.Fspinnerf(o.clientset.Stdout, "Deleting resources from podman")
			if isPodmanInnerLoopDeployed {
				// TODO(feloy) #6424
				_ = isPodmanInnerLoopDeployed
			}
			for _, pod := range podmanPods {
				err = o.clientset.PodmanClient.CleanupPodResources(pod, true)
				if err != nil {
					log.Fwarningf(o.clientset.Stderr, "Failed to delete the pod %q from podman: %s\n", pod.GetName(), err)
				}
			}
			spinner.End(true)
			log.Finfof(o.clientset.Stdout, "The component %q is successfully deleted from podman", componentName)
		}

		if o.withFilesFlag || len(orphans) > 0 {
			// Delete files
			remainingFiles := o.deleteFilesCreatedByOdo(o.clientset.FS, filesToDelete)
			var listOfFiles []string
			for f, e := range remainingFiles {
				log.Fwarningf(o.clientset.Stderr, "Failed to delete file or directory: %s: %v\n", f, e)
				listOfFiles = append(listOfFiles, "\t- "+f)
			}
			if len(remainingFiles) != 0 {
				log.Fprintf(o.clientset.Stdout, "There are still files or directories that could not be deleted.")
				fmt.Fprintln(o.clientset.Stdout, strings.Join(listOfFiles, "\n"))
				log.Finfof(o.clientset.Stdout, "You need to manually delete those.")
			}
		}
		return remainingResources, nil
	}

	log.Ferror(o.clientset.Stderr, "Aborting deletion of component")
	return remainingResources, nil
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
func (o *ComponentOptions) printDevfileComponents(
	componentName, namespace string,
	k8sResources []unstructured.Unstructured,
	podmanResources []*corev1.Pod,
) {
	log.Finfof(o.clientset.Stdout, infoMsg(
		len(k8sResources) != 0,
		len(podmanResources) != 0,
		componentName,
		namespace,
	))

	if len(k8sResources) != 0 {
		log.Fprintf(o.clientset.Stdout, "The following resources will get deleted from cluster:")
		for _, resource := range k8sResources {
			log.Fprintf(o.clientset.Stdout, "\t- %s: %s", resource.GetKind(), resource.GetName())
		}
		log.Fprintln(o.clientset.Stdout)
	}

	if len(podmanResources) != 0 {
		log.Fprintf(o.clientset.Stdout, "The following pods and associated volumes will get deleted from podman:")
		for _, pod := range podmanResources {
			log.Fprintf(o.clientset.Stdout, "\t- %s", pod.GetName())
		}
		log.Fprintln(o.clientset.Stdout)
	}
}

func infoMsg(
	cluster, podman bool,
	componentName, namespace string,
) string {
	froms := []string{}
	if cluster {
		froms = append(froms, fmt.Sprintf("from the namespace %q", namespace))
	}
	if podman {
		froms = append(froms, "from podman")
	}
	return fmt.Sprintf("This will delete %q %s.", componentName, strings.Join(froms, " and "))

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

// getOrphanDevstateFiles gets the list of all Devstate files for which no odo process exists
func (o *ComponentOptions) getOrphanDevstateFiles(filesys filesystem.Filesystem, ctx context.Context) ([]string, error) {
	var list []string
	if o.runningIn != labels.ComponentDeployMode {
		var orphanDevstates []string
		orphanDevstates, err := o.clientset.StateClient.GetOrphanFiles(ctx)
		if err != nil {
			return nil, err
		}
		list = append(list, orphanDevstates...)
	}
	return list, nil
}

func (o *ComponentOptions) printFileCreatedByOdo(files []string, hasClusterResources bool) {
	if len(files) == 0 {
		return
	}

	m := "This will "
	if hasClusterResources {
		m += "also "
	}
	log.Finfof(o.clientset.Stdout, m+"delete the following files and directories:")
	for _, f := range files {
		fmt.Fprintln(o.clientset.Stdout, "\t- "+f)
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
func NewCmdComponent(ctx context.Context, name, fullName string, testClientset clientset.Clientset) *cobra.Command {
	o := NewComponentOptions()

	var componentCmd = &cobra.Command{
		Use:     name,
		Short:   "Delete component",
		Long:    "Delete component",
		Args:    genericclioptions.NoArgsAndSilenceJSON,
		Example: fmt.Sprintf(deleteExample, fullName),
		RunE: func(cmd *cobra.Command, args []string) error {
			return genericclioptions.GenericRun(o, testClientset, cmd, args)
		},
	}
	componentCmd.Flags().StringVar(&o.name, "name", "", "Name of the component to delete, optional. By default, the component described in the local devfile is deleted")
	componentCmd.Flags().StringVar(&o.namespace, "namespace", "", "Namespace in which to find the component to delete, optional. By default, the current namespace defined in kubeconfig is used")
	componentCmd.Flags().StringVar(&o.runningInFlag, "running-in", "",
		"Delete resources running in the specified mode, optional. By default, all resources created by odo for the component are deleted.")
	componentCmd.Flags().BoolVarP(&o.withFilesFlag, "files", "", false, "Delete all files and directories generated by odo. Use with caution.")
	componentCmd.Flags().BoolVarP(&o.forceFlag, "force", "f", false, "Delete component without prompting")
	componentCmd.Flags().BoolVarP(&o.waitFlag, "wait", "w", false, "Wait for deletion of all dependent resources")
	clientset.Add(componentCmd, clientset.DELETE_COMPONENT, clientset.KUBERNETES, clientset.FILESYSTEM, clientset.STATE)
	if feature.IsEnabled(ctx, feature.GenericPlatformFlag) {
		clientset.Add(componentCmd, clientset.PODMAN_NULLABLE)
	}
	commonflags.UsePlatformFlag(componentCmd)

	return componentCmd
}
