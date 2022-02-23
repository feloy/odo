package component

import (
	"fmt"
	"io"
	"os"
	"reflect"
	"strings"
	"time"

	"k8s.io/utils/pointer"

	"github.com/devfile/library/pkg/devfile/generator"
	componentlabels "github.com/redhat-developer/odo/pkg/component/labels"
	"github.com/redhat-developer/odo/pkg/devfile"
	"github.com/redhat-developer/odo/pkg/envinfo"
	"github.com/redhat-developer/odo/pkg/preference"
	"github.com/redhat-developer/odo/pkg/service"
	"github.com/redhat-developer/odo/pkg/util"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/pkg/errors"
	"k8s.io/klog"

	devfilev1 "github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
	parsercommon "github.com/devfile/library/pkg/devfile/parser/data/v2/common"
	"github.com/redhat-developer/odo/pkg/component"
	"github.com/redhat-developer/odo/pkg/devfile/adapters/common"
	"github.com/redhat-developer/odo/pkg/devfile/adapters/kubernetes/storage"
	"github.com/redhat-developer/odo/pkg/devfile/adapters/kubernetes/utils"
	"github.com/redhat-developer/odo/pkg/kclient"
	"github.com/redhat-developer/odo/pkg/log"
	storagepkg "github.com/redhat-developer/odo/pkg/storage"
	"github.com/redhat-developer/odo/pkg/sync"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
)

const supervisorDStatusWaitTimeInterval = 1

// New instantiates a component adapter
func New(adapterContext common.AdapterContext, client kclient.ClientInterface, prefClient preference.Client) Adapter {

	adapter := Adapter{Client: client, prefClient: prefClient}
	adapter.GenericAdapter = common.NewGenericAdapter(&adapter, adapterContext)
	adapter.GenericAdapter.InitWith(&adapter)
	return adapter
}

// getPod lazily records and retrieves the pod associated with the component associated with this adapter. If refresh parameter
// is true, then the pod is refreshed from the cluster regardless of its current local state
func (a *Adapter) getPod(refresh bool) (*corev1.Pod, error) {
	if refresh || a.pod == nil {
		podSelector := fmt.Sprintf("component=%s", a.ComponentName)

		// Wait for Pod to be in running state otherwise we can't sync data to it.
		pod, err := a.Client.WaitAndGetPodWithEvents(podSelector, corev1.PodRunning, "Waiting for component to start", time.Duration(a.prefClient.GetPushTimeout())*time.Second)
		if err != nil {
			return nil, errors.Wrapf(err, "error while waiting for pod %s", podSelector)
		}
		a.pod = pod
	}
	return a.pod, nil
}

func (a *Adapter) ComponentInfo(command devfilev1.Command) (common.ComponentInfo, error) {
	pod, err := a.getPod(false)
	if err != nil {
		return common.ComponentInfo{}, err
	}
	return common.ComponentInfo{
		PodName:       pod.Name,
		ContainerName: command.Exec.Component,
	}, nil
}

func (a *Adapter) SupervisorComponentInfo(command devfilev1.Command) (common.ComponentInfo, error) {
	pod, err := a.getPod(false)
	if err != nil {
		return common.ComponentInfo{}, err
	}
	for _, container := range pod.Spec.Containers {
		if container.Name == command.Exec.Component && !reflect.DeepEqual(container.Command, []string{common.SupervisordBinaryPath}) {
			return common.ComponentInfo{
				ContainerName: command.Exec.Component,
				PodName:       pod.Name,
			}, nil
		}
	}
	return common.ComponentInfo{}, nil
}

// Adapter is a component adapter implementation for Kubernetes
type Adapter struct {
	Client     kclient.ClientInterface
	prefClient preference.Client

	*common.GenericAdapter

	devfileBuildCmd  string
	devfileRunCmd    string
	devfileDebugCmd  string
	devfileDebugPort int
	pod              *corev1.Pod
	deployment       *appsv1.Deployment
}

// Push updates the component if a matching component exists or creates one if it doesn't exist
// Once the component has started, it will sync the source code to it.
func (a Adapter) Push(parameters common.PushParameters) (err error) {

	a.deployment, err = a.Client.GetOneDeployment(a.ComponentName, a.AppName)
	if err != nil {
		if _, ok := err.(*kclient.DeploymentNotFoundError); !ok {
			return errors.Wrapf(err, "unable to determine if component %s exists", a.ComponentName)
		}
	}

	componentExists := a.deployment != nil

	a.devfileBuildCmd = parameters.DevfileBuildCmd
	a.devfileRunCmd = parameters.DevfileRunCmd
	a.devfileDebugCmd = parameters.DevfileDebugCmd
	a.devfileDebugPort = parameters.DebugPort

	podChanged := false
	var podName string

	// If the component already exists, retrieve the pod's name before it's potentially updated
	if componentExists {
		// First see if the component does have a pod. it could have been scaled down to zero
		_, err = a.Client.GetOnePodFromSelector(fmt.Sprintf("component=%s", a.ComponentName))
		// If an error occurs, we don't call a.getPod (a blocking function that waits till it finds a pod in "Running" state.)
		// We would rely on a call to a.createOrUpdateComponent to reset the pod count for the component to one.
		if err == nil {
			pod, podErr := a.getPod(true)
			if podErr != nil {
				return errors.Wrapf(podErr, "unable to get pod for component %s", a.ComponentName)
			}
			podName = pod.GetName()
		}
	}

	// Validate the devfile build and run commands
	log.Info("\nValidation")
	s := log.Spinner("Validating the devfile")
	err = util.ValidateK8sResourceName("component name", a.ComponentName)
	if err != nil {
		return err
	}

	err = util.ValidateK8sResourceName("component namespace", parameters.EnvSpecificInfo.GetNamespace())
	if err != nil {
		return err
	}

	pushDevfileCommands, err := common.ValidateAndGetPushDevfileCommands(a.Devfile.Data, a.devfileBuildCmd, a.devfileRunCmd)
	if err != nil {
		s.End(false)
		return errors.Wrap(err, "failed to validate devfile build and run commands")
	}
	s.End(true)

	labels := componentlabels.GetLabels(a.ComponentName, a.AppName, true)

	previousMode := parameters.EnvSpecificInfo.GetRunMode()
	currentMode := envinfo.Run

	if parameters.Debug {
		pushDevfileDebugCommands, e := common.ValidateAndGetDebugDevfileCommands(a.Devfile.Data, a.devfileDebugCmd)
		if e != nil {
			return fmt.Errorf("debug command is not valid")
		}
		pushDevfileCommands[devfilev1.DebugCommandGroupKind] = pushDevfileDebugCommands
		currentMode = envinfo.Debug
	}

	if currentMode != previousMode {
		parameters.RunModeChanged = true
	}

	// fetch the "kubernetes inlined components" to create them on cluster
	// from odo standpoint, these components contain yaml manifest of an odo service or an odo link
	k8sComponents, err := devfile.GetKubernetesComponentsToPush(a.Devfile)
	if err != nil {
		return errors.Wrap(err, "error while trying to fetch service(s) from devfile")
	}

	log.Infof("\nCreating Services for component %s", a.ComponentName)

	// validate if the GVRs represented by Kubernetes inlined components are supported by the underlying cluster
	err = service.ValidateResourcesExist(a.Client, k8sComponents, a.Context)
	if err != nil {
		return err
	}

	// create the Kubernetes objects from the manifest and delete the ones not in the devfile
	err = service.PushKubernetesResources(a.Client, k8sComponents, labels, a.Context)
	if err != nil {
		return errors.Wrap(err, "failed to create service(s) associated with the component")
	}

	log.Infof("\nCreating Kubernetes resources for component %s", a.ComponentName)

	isMainStorageEphemeral := a.prefClient.GetEphemeralSourceVolume()
	err = a.createOrUpdateComponent(componentExists, parameters.EnvSpecificInfo, isMainStorageEphemeral)
	if err != nil {
		return errors.Wrap(err, "unable to create or update component")
	}

	a.deployment, err = a.Client.WaitForDeploymentRollout(a.deployment.Name)
	if err != nil {
		return errors.Wrap(err, "error while waiting for deployment rollout")
	}

	// Wait for Pod to be in running state otherwise we can't sync data or exec commands to it.
	pod, err := a.getPod(true)
	if err != nil {
		return errors.Wrapf(err, "unable to get pod for component %s", a.ComponentName)
	}

	// list the latest state of the PVCs
	pvcs, err := a.Client.ListPVCs(fmt.Sprintf("%v=%v", "component", a.ComponentName))
	if err != nil {
		return err
	}

	ownerReference := generator.GetOwnerReference(a.deployment)
	// update the owner reference of the PVCs with the deployment
	for i := range pvcs {
		if pvcs[i].OwnerReferences != nil || pvcs[i].DeletionTimestamp != nil {
			continue
		}
		err = a.Client.UpdateStorageOwnerReference(&pvcs[i], ownerReference)
		if err != nil {
			return err
		}
	}

	err = service.UpdateServicesWithOwnerReferences(a.Client, k8sComponents, ownerReference, a.Context)
	if err != nil {
		return err
	}

	// create the Kubernetes objects from the manifest and delete the ones not in the devfile
	needRestart, err := service.PushLinks(a.Client, k8sComponents, labels, a.deployment, a.Context)
	if err != nil {
		return errors.Wrap(err, "failed to create service(s) associated with the component")
	}

	if needRestart {
		s := log.Spinner("Restarting the component")
		defer s.End(false)
		err = a.Client.WaitForPodDeletion(pod.Name)
		if err != nil {
			return err
		}
		s.End(true)
	}

	a.deployment, err = a.Client.WaitForDeploymentRollout(a.deployment.Name)
	if err != nil {
		return errors.Wrap(err, "error while waiting for deployment rollout")
	}

	// Wait for Pod to be in running state otherwise we can't sync data or exec commands to it.
	pod, err = a.getPod(true)
	if err != nil {
		return errors.Wrapf(err, "unable to get pod for component %s", a.ComponentName)
	}

	parameters.EnvSpecificInfo.SetDevfileObj(a.Devfile)
	err = component.ApplyConfig(a.Client, parameters.EnvSpecificInfo)
	if err != nil {
		return errors.Wrapf(err, "Failed to update config to component deployed.")
	}

	// Compare the name of the pod with the one before the rollout. If they differ, it means there's a new pod and a force push is required
	if componentExists && podName != pod.GetName() {
		podChanged = true
	}

	// Find at least one pod with the source volume mounted, error out if none can be found
	containerName, syncFolder, err := getFirstContainerWithSourceVolume(pod.Spec.Containers)
	if err != nil {
		return errors.Wrapf(err, "error while retrieving container from pod %s with a mounted project volume", podName)
	}

	log.Infof("\nSyncing to component %s", a.ComponentName)
	// Get a sync adapter. Check if project files have changed and sync accordingly
	syncAdapter := sync.New(a.AdapterContext, &a)
	compInfo := common.ComponentInfo{
		ContainerName: containerName,
		PodName:       pod.GetName(),
		SyncFolder:    syncFolder,
	}
	syncParams := common.SyncParameters{
		PushParams:      parameters,
		CompInfo:        compInfo,
		ComponentExists: componentExists,
		PodChanged:      podChanged,
		Files:           common.GetSyncFilesFromAttributes(pushDevfileCommands),
	}

	execRequired, err := syncAdapter.SyncFiles(syncParams)
	if err != nil {
		return errors.Wrapf(err, "Failed to sync to component with name %s", a.ComponentName)
	}

	// PostStart events from the devfile will only be executed when the component
	// didn't previously exist
	postStartEvents := a.Devfile.Data.GetEvents().PostStart
	if !componentExists && len(postStartEvents) > 0 {
		err = a.ExecDevfileEvent(postStartEvents, common.PostStart, parameters.Show)
		if err != nil {
			return err

		}
	}

	runCommand := pushDevfileCommands[devfilev1.RunCommandGroupKind]
	if parameters.Debug {
		runCommand = pushDevfileCommands[devfilev1.DebugCommandGroupKind]
	}
	running, err := a.GetSupervisordCommandStatus(runCommand)
	if err != nil {
		return err
	}

	if !running || execRequired || parameters.RunModeChanged {
		log.Infof("\nExecuting devfile commands for component %s", a.ComponentName)
		err = a.ExecDevfile(pushDevfileCommands, componentExists, parameters)
		if err != nil {
			return err
		}

		// wait for a second
		wait := time.After(supervisorDStatusWaitTimeInterval * time.Second)
		<-wait

		err := a.CheckSupervisordCommandStatus(runCommand)
		if err != nil {
			return err
		}
	} else {
		// no file was modified/added/deleted/renamed, thus return without syncing files
		log.Success("No file changes detected, skipping build. Use the '-f' flag to force the build.")
	}

	return nil
}

// GetSupervisordCommandStatus returns true if the command is running
// based on `supervisord ctl` output and returns an error if
// the command is not known by supervisord
func (a Adapter) GetSupervisordCommandStatus(command devfilev1.Command) (bool, error) {
	statusInContainer := getSupervisordStatusInContainer(a.pod.Name, command.Exec.Component, a)

	supervisordProgramName := "devrun"

	// if the command is a debug one, we check against `debugrun`
	if command.Exec.Group.Kind == devfilev1.DebugCommandGroupKind {
		supervisordProgramName = "debugrun"
	}

	for _, status := range statusInContainer {
		if strings.EqualFold(status.program, supervisordProgramName) {
			return strings.EqualFold(status.status, "running"), nil
		}
	}
	return false, fmt.Errorf("the supervisord program %s not found", supervisordProgramName)
}

// CheckSupervisordCommandStatus checks if the command is running based on supervisord status output.
// if the command is not in a running state, we fetch the last 20 lines of the component's log and display it
func (a Adapter) CheckSupervisordCommandStatus(command devfilev1.Command) error {

	running, err := a.GetSupervisordCommandStatus(command)
	if err != nil {
		return err
	}

	if !running {
		numberOfLines := 20
		log.Warningf("devfile command %q exited with error status within %d sec", command.Id, supervisorDStatusWaitTimeInterval)
		log.Infof("Last %d lines of the component's log:", numberOfLines)

		rd, err := a.Log(false, command)
		if err != nil {
			return err
		}

		err = util.DisplayLog(false, rd, os.Stderr, a.ComponentName, numberOfLines)
		if err != nil {
			return err
		}

		log.Info("To get the full log output, please run 'odo log'")
	}
	return nil
}

// Test runs the devfile test command
func (a Adapter) Test(testCmd string, show bool) (err error) {
	pod, err := a.Client.GetOnePod(a.ComponentName, a.AppName)
	if err != nil {
		return fmt.Errorf("error occurred while getting the pod: %w", err)
	}
	if pod.Status.Phase != corev1.PodRunning {
		return fmt.Errorf("pod for component %s is not running", a.ComponentName)
	}

	log.Infof("\nExecuting devfile test command for component %s", a.ComponentName)

	testCommand, err := common.ValidateAndGetTestDevfileCommands(a.Devfile.Data, testCmd)
	if err != nil {
		return errors.Wrap(err, "failed to validate devfile test command")
	}
	err = a.ExecuteDevfileCommand(testCommand, show, false)
	if err != nil {
		return errors.Wrapf(err, "failed to execute devfile commands for component %s", a.ComponentName)
	}
	return nil
}

// DoesComponentExist returns true if a component with the specified name exists, false otherwise
func (a Adapter) DoesComponentExist(cmpName string, appName string) (bool, error) {
	return utils.ComponentExists(a.Client, cmpName, appName)
}

func (a *Adapter) createOrUpdateComponent(componentExists bool, ei envinfo.EnvSpecificInfo, isMainStorageEphemeral bool) (err error) {
	ei.SetDevfileObj(a.Devfile)

	storageClient := storagepkg.NewClient(storagepkg.ClientOptions{
		Client:              a.Client,
		LocalConfigProvider: &ei,
	})

	// handle the ephemeral storage
	err = storage.HandleEphemeralStorage(a.Client, storageClient, a.ComponentName, isMainStorageEphemeral)
	if err != nil {
		return err
	}

	// From devfile info, create PVCs and return ephemeral storages
	ephemerals, err := storagepkg.Push(storageClient, &ei)
	if err != nil {
		return err
	}

	componentName := a.ComponentName
	var componentType string
	// We insert the component type in deployment annotations because its value might be in a namespaced/versioned format,
	// since labels do not support such formats, we extract the component type from these formats before assigning its value to the corresponding label.
	// This annotated value will later be used when listing the components; we do this to list/describe and stay inline with the component type value set in the devfile.
	annotatedComponentType := component.GetComponentTypeFromDevfileMetadata(a.AdapterContext.Devfile.Data.GetMetadata())
	if annotatedComponentType != component.NotAvailable {
		componentType = strings.TrimSuffix(util.ExtractComponentType(componentType), "-")
	}

	labels := componentlabels.GetLabels(componentName, a.AppName, true)
	labels["component"] = componentName
	labels[componentlabels.ComponentTypeLabel] = componentType

	containers, err := generator.GetContainers(a.Devfile, parsercommon.DevfileOptions{})
	if err != nil {
		return err
	}
	if len(containers) == 0 {
		return fmt.Errorf("no valid components found in the devfile")
	}

	// Add the project volume before generating init containers
	utils.AddOdoProjectVolume(&containers)

	containers, err = utils.UpdateContainersWithSupervisord(a.Devfile, containers, a.devfileRunCmd, a.devfileDebugCmd, a.devfileDebugPort)
	if err != nil {
		return err
	}

	initContainers, err := generator.GetInitContainers(a.Devfile)
	if err != nil {
		return err
	}

	initContainers = append(initContainers, kclient.GetBootstrapSupervisordInitContainer())

	// list all the pvcs for the component
	pvcs, err := a.Client.ListPVCs(fmt.Sprintf("%v=%v", "component", a.ComponentName))
	if err != nil {
		return err
	}

	odoSourcePVCName, volumeNameToVolInfo, err := storage.GetVolumeInfos(pvcs)
	if err != nil {
		return err
	}

	var allVolumes []corev1.Volume

	// Get PVC volumes and Volume Mounts
	pvcVolumes, err := storage.GetPersistentVolumesAndVolumeMounts(a.Devfile, containers, initContainers, volumeNameToVolInfo, parsercommon.DevfileOptions{})
	if err != nil {
		return err
	}
	allVolumes = append(allVolumes, pvcVolumes...)

	ephemeralVolumes, err := storage.GetEphemeralVolumesAndVolumeMounts(a.Devfile, containers, initContainers, ephemerals, parsercommon.DevfileOptions{})
	if err != nil {
		return err
	}
	allVolumes = append(allVolumes, ephemeralVolumes...)

	odoMandatoryVolumes := utils.GetOdoContainerVolumes(odoSourcePVCName)
	allVolumes = append(allVolumes, odoMandatoryVolumes...)

	selectorLabels := map[string]string{
		"component": componentName,
	}

	deploymentObjectMeta, err := a.generateDeploymentObjectMeta(labels)
	if err != nil {
		return err
	}

	deployParams := generator.DeploymentParams{
		TypeMeta:          generator.GetTypeMeta(kclient.DeploymentKind, kclient.DeploymentAPIVersion),
		ObjectMeta:        deploymentObjectMeta,
		InitContainers:    initContainers,
		Containers:        containers,
		Volumes:           allVolumes,
		PodSelectorLabels: selectorLabels,
	}

	deployment, err := generator.GetDeployment(a.Devfile, deployParams)
	if err != nil {
		return err
	}
	deployment.Spec.Replicas = pointer.Int32Ptr(1)
	if deployment.Annotations == nil {
		deployment.Annotations = make(map[string]string)
	}

	// Add annotation for component type; this will later be used while listing/describing a component
	deployment.Annotations[componentlabels.ComponentTypeAnnotation] = annotatedComponentType

	if vcsUri := util.GetGitOriginPath(a.Context); vcsUri != "" {
		deployment.Annotations["app.openshift.io/vcs-uri"] = vcsUri
	}

	// add the annotations to the service for linking
	serviceAnnotations := make(map[string]string)
	serviceAnnotations["service.binding/backend_ip"] = "path={.spec.clusterIP}"
	serviceAnnotations["service.binding/backend_port"] = "path={.spec.ports},elementType=sliceOfMaps,sourceKey=name,sourceValue=port"

	serviceName, err := util.NamespaceKubernetesObjectWithTrim(componentName, a.AppName)
	if err != nil {
		return err
	}
	serviceObjectMeta := generator.GetObjectMeta(serviceName, a.Client.GetCurrentNamespace(), labels, serviceAnnotations)
	serviceParams := generator.ServiceParams{
		ObjectMeta:     serviceObjectMeta,
		SelectorLabels: selectorLabels,
	}
	svc, err := generator.GetService(a.Devfile, serviceParams, parsercommon.DevfileOptions{})

	if err != nil {
		return err
	}
	klog.V(2).Infof("Creating deployment %v", deployment.Spec.Template.GetName())
	klog.V(2).Infof("The component name is %v", componentName)
	if componentExists {
		// If the component already exists, get the resource version of the deploy before updating
		klog.V(2).Info("The component already exists, attempting to update it")
		if a.Client.IsSSASupported() {
			a.deployment, err = a.Client.ApplyDeployment(*deployment)
		} else {
			a.deployment, err = a.Client.UpdateDeployment(*deployment)
		}
		if err != nil {
			return err
		}
		klog.V(2).Infof("Successfully updated component %v", componentName)
		e := a.createOrUpdateServiceForComponent(svc, componentName)
		if e != nil {
			return e
		}
	} else {
		if a.Client.IsSSASupported() {
			a.deployment, err = a.Client.ApplyDeployment(*deployment)
		} else {
			a.deployment, err = a.Client.CreateDeployment(*deployment)
		}

		if err != nil {
			return err
		}
		klog.V(2).Infof("Successfully created component %v", componentName)
		ownerReference := generator.GetOwnerReference(a.deployment)
		svc.OwnerReferences = append(svc.OwnerReferences, ownerReference)
		if len(svc.Spec.Ports) > 0 {
			_, err = a.Client.CreateService(*svc)
			if err != nil {
				return err
			}
			klog.V(2).Infof("Successfully created Service for component %s", componentName)
		}

	}

	return nil
}

func (a *Adapter) createOrUpdateServiceForComponent(svc *corev1.Service, componentName string) error {
	oldSvc, err := a.Client.GetOneService(a.ComponentName, a.AppName)
	ownerReference := generator.GetOwnerReference(a.deployment)
	svc.OwnerReferences = append(svc.OwnerReferences, ownerReference)
	if err != nil {
		// no old service was found, create a new one
		if len(svc.Spec.Ports) > 0 {
			_, err = a.Client.CreateService(*svc)
			if err != nil {
				return err
			}
			klog.V(2).Infof("Successfully created Service for component %s", componentName)
		}
		return nil
	}
	if len(svc.Spec.Ports) > 0 {
		svc.Spec.ClusterIP = oldSvc.Spec.ClusterIP
		svc.ResourceVersion = oldSvc.GetResourceVersion()
		_, err = a.Client.UpdateService(*svc)
		if err != nil {
			return err
		}
		klog.V(2).Infof("Successfully update Service for component %s", componentName)
		return nil
	}
	// delete the old existing service if the component currently doesn't expose any ports
	return a.Client.DeleteService(oldSvc.Name)
}

// generateDeploymentObjectMeta generates a ObjectMeta object for the given deployment's name and labels
// if no deployment exists, it creates a new deployment name
func (a Adapter) generateDeploymentObjectMeta(labels map[string]string) (metav1.ObjectMeta, error) {
	if a.deployment != nil {
		return generator.GetObjectMeta(a.deployment.Name, a.Client.GetCurrentNamespace(), labels, nil), nil
	} else {
		deploymentName, err := util.NamespaceKubernetesObject(a.ComponentName, a.AppName)
		if err != nil {
			return metav1.ObjectMeta{}, err
		}
		return generator.GetObjectMeta(deploymentName, a.Client.GetCurrentNamespace(), labels, nil), nil
	}
}

// getFirstContainerWithSourceVolume returns the first container that set mountSources: true as well
// as the path to the source volume inside the container.
// Because the source volume is shared across all components that need it, we only need to sync once,
// so we only need to find one container. If no container was found, that means there's no
// container to sync to, so return an error
func getFirstContainerWithSourceVolume(containers []corev1.Container) (string, string, error) {
	for _, c := range containers {
		for _, env := range c.Env {
			if env.Name == generator.EnvProjectsSrc {
				return c.Name, env.Value, nil
			}
		}
	}

	return "", "", fmt.Errorf("in order to sync files, odo requires at least one component in a devfile to set 'mountSources: true'")
}

// Delete deletes the component
func (a Adapter) Delete(labels map[string]string, show bool, wait bool) error {
	if labels == nil {
		return fmt.Errorf("cannot delete with labels being nil")
	}
	log.Printf("Gathering information for component: %q", a.ComponentName)
	podSpinner := log.Spinner("Checking status for component")
	defer podSpinner.End(false)

	pod, err := a.Client.GetOnePod(a.ComponentName, a.AppName)
	if kerrors.IsForbidden(err) {
		klog.V(2).Infof("Resource for %s forbidden", a.ComponentName)
		// log the error if it failed to determine if the component exists due to insufficient RBACs
		podSpinner.End(false)
		log.Warningf("%v", err)
		return nil
	} else if e, ok := err.(*kclient.PodNotFoundError); ok {
		podSpinner.End(false)
		log.Warningf("%v", e)
		return nil
	} else if err != nil {
		return errors.Wrapf(err, "unable to determine if component %s exists", a.ComponentName)
	}

	podSpinner.End(true)

	// if there are preStop events, execute them before deleting the deployment
	preStopEvents := a.Devfile.Data.GetEvents().PreStop
	if len(preStopEvents) > 0 {
		if pod.Status.Phase != corev1.PodRunning {
			return fmt.Errorf("unable to execute preStop events, pod for component %s is not running", a.ComponentName)
		}

		err = a.ExecDevfileEvent(preStopEvents, common.PreStop, show)
		if err != nil {
			return err
		}
	}

	log.Infof("\nDeleting component %s", a.ComponentName)
	spinner := log.Spinner("Deleting Kubernetes resources for component")
	defer spinner.End(false)

	err = a.Client.Delete(labels, wait)
	if err != nil {
		return err
	}

	spinner.End(true)
	log.Successf("Successfully deleted component")
	return nil
}

// Log returns log from component
func (a Adapter) Log(follow bool, command devfilev1.Command) (io.ReadCloser, error) {

	pod, err := a.Client.GetOnePod(a.ComponentName, a.AppName)
	if err != nil {
		return nil, errors.Errorf("the component %s doesn't exist on the cluster", a.ComponentName)
	}

	if pod.Status.Phase != corev1.PodRunning {
		return nil, errors.Errorf("unable to show logs, component is not in running state. current status=%v", pod.Status.Phase)
	}

	containerName := command.Exec.Component

	return a.Client.GetPodLogs(pod.Name, containerName, follow)
}

// Exec executes a command in the component
func (a Adapter) Exec(command []string) error {
	exists, err := utils.ComponentExists(a.Client, a.ComponentName, a.AppName)
	if err != nil {
		return err
	}

	if !exists {
		return errors.Errorf("the component %s doesn't exist on the cluster", a.ComponentName)
	}

	runCommand, err := common.GetRunCommand(a.Devfile.Data, "")
	if err != nil {
		return err
	}
	containerName := runCommand.Exec.Component

	// get the pod
	pod, err := a.Client.GetOnePod(a.ComponentName, a.AppName)
	if err != nil {
		return errors.Wrapf(err, "unable to get pod for component %s", a.ComponentName)
	}

	if pod.Status.Phase != corev1.PodRunning {
		return fmt.Errorf("unable to exec as the component is not running. Current status=%v", pod.Status.Phase)
	}

	componentInfo := common.ComponentInfo{
		PodName:       pod.Name,
		ContainerName: containerName,
	}

	return a.ExecuteCommand(componentInfo, command, true, nil, nil)
}

func (a Adapter) ExecCMDInContainer(componentInfo common.ComponentInfo, cmd []string, stdout io.Writer, stderr io.Writer, stdin io.Reader, tty bool) error {
	return a.Client.ExecCMDInContainer(componentInfo.ContainerName, componentInfo.PodName, cmd, stdout, stderr, stdin, tty)
}

// ExtractProjectToComponent extracts the project archive(tar) to the target path from the reader stdin
func (a Adapter) ExtractProjectToComponent(componentInfo common.ComponentInfo, targetPath string, stdin io.Reader) error {
	return a.Client.ExtractProjectToComponent(componentInfo.ContainerName, componentInfo.PodName, targetPath, stdin)
}

// Deploy executes the 'deploy' command defined in a devfile
func (a Adapter) Deploy() error {
	deployCmd, err := a.getDeployCommand()
	if err != nil {
		return err
	}

	return a.ExecuteDevfileCommand(deployCmd, true, false)
}

// UnDeploy reverses the effect of the 'deploy' command defined in a devfile
func (a Adapter) UnDeploy() error {
	deployCmd, err := a.getDeployCommand()
	if err != nil {
		return err
	}
	return a.ExecuteDevfileCommand(deployCmd, true, true)
}

// ExecuteDevfileCommand executes the devfile command; if unexecute is set to true, it reverses the effect of Execute
func (a Adapter) ExecuteDevfileCommand(command devfilev1.Command, show, unexecute bool) error {
	commands, err := a.Devfile.Data.GetCommands(parsercommon.DevfileOptions{})
	if err != nil {
		return err
	}

	c, err := common.New(command, common.GetCommandsMap(commands), &a)
	if err != nil {
		return err
	}
	if unexecute {
		return c.UnExecute()
	}
	return c.Execute(show)
}

// ApplyComponent 'applies' a devfile component
func (a Adapter) ApplyComponent(componentName string) error {
	cmp, err := a.getApplyComponent(componentName)
	if err != nil {
		return err
	}

	return cmp.Apply(a.Devfile, a.Context)
}

// UnApplyComponent un-'applies' a devfile component
func (a Adapter) UnApplyComponent(componentName string) error {
	cmp, err := a.getApplyComponent(componentName)
	if err != nil {
		return err
	}
	return cmp.UnApply(a.Context)
}

// getDeployCommand validates the deploy command and returns it
func (a Adapter) getDeployCommand() (devfilev1.Command, error) {
	deployGroupCmd, err := a.Devfile.Data.GetCommands(parsercommon.DevfileOptions{
		CommandOptions: parsercommon.CommandOptions{
			CommandGroupKind: devfilev1.DeployCommandGroupKind,
		},
	})
	if err != nil {
		return devfilev1.Command{}, err
	}
	if len(deployGroupCmd) == 0 {
		return devfilev1.Command{}, &NoDefaultDeployCommandFoundError{}
	}
	if len(deployGroupCmd) > 1 {
		var found bool
		var foundGroupCmd devfilev1.Command
		for _, groupCmd := range deployGroupCmd {
			group := parsercommon.GetGroup(groupCmd)
			if group == nil {
				continue
			}
			if group.IsDefault != nil && *group.IsDefault {
				if found {
					return devfilev1.Command{}, &MoreThanOneDefaultDeployCommandFoundError{}
				}
				found = true
				foundGroupCmd = groupCmd
			}
		}
		if !found {
			return devfilev1.Command{}, &NoDefaultDeployCommandFoundError{}
		}
		return foundGroupCmd, nil
	}
	return deployGroupCmd[0], nil
}

// getApplyComponent returns the 'Apply' command's component(kubernetes/image)
func (a Adapter) getApplyComponent(componentName string) (componentToApply, error) {
	components, err := a.Devfile.Data.GetComponents(parsercommon.DevfileOptions{})
	if err != nil {
		return nil, err
	}
	var component devfilev1.Component
	var found bool
	for _, component = range components {
		if component.Name == componentName {
			found = true
			break
		}
	}
	if !found {
		return nil, fmt.Errorf("component %q not found", componentName)
	}

	cmp, err := createComponent(a, component)
	if err != nil {
		return nil, err
	}
	return cmp, nil
}
