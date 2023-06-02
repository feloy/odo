package kubedev

import (
	"context"
	"fmt"
	"path/filepath"
	"time"

	devfilev1 "github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
	parsercommon "github.com/devfile/library/v2/pkg/devfile/parser/data/v2/common"

	"github.com/redhat-developer/odo/pkg/component"
	"github.com/redhat-developer/odo/pkg/dev/common"
	"github.com/redhat-developer/odo/pkg/devfile/image"
	"github.com/redhat-developer/odo/pkg/libdevfile"
	"github.com/redhat-developer/odo/pkg/log"
	odocontext "github.com/redhat-developer/odo/pkg/odo/context"
	"github.com/redhat-developer/odo/pkg/port"
	"github.com/redhat-developer/odo/pkg/sync"
	"github.com/redhat-developer/odo/pkg/watch"

	"k8s.io/klog"
)

func (o *DevClient) innerloop(ctx context.Context, parameters common.PushParameters, componentStatus *watch.ComponentStatus) error {
	var (
		componentName = odocontext.GetComponentName(ctx)
		devfilePath   = odocontext.GetDevfilePath(ctx)
		path          = filepath.Dir(devfilePath)
	)

	// Now the Deployment has a Ready replica, we can get the Pod to work inside it
	pod, err := o.kubernetesClient.GetPodUsingComponentName(componentName)
	if err != nil {
		return fmt.Errorf("unable to get pod for component %s: %w", componentName, err)
	}

	// Find at least one pod with the source volume mounted, error out if none can be found
	containerName, syncFolder, err := common.GetFirstContainerWithSourceVolume(pod.Spec.Containers)
	if err != nil {
		return fmt.Errorf("error while retrieving container from pod %s with a mounted project volume: %w", pod.GetName(), err)
	}

	s := log.Spinner("Syncing files into the container")
	defer s.End(false)

	// Get commands
	pushDevfileCommands, err := o.getPushDevfileCommands(parameters)
	if err != nil {
		return fmt.Errorf("failed to validate devfile build and run commands: %w", err)
	}

	podChanged := componentStatus.GetState() == watch.StateWaitDeployment

	// Get a sync adapter. Check if project files have changed and sync accordingly
	compInfo := sync.ComponentInfo{
		ComponentName: componentName,
		ContainerName: containerName,
		PodName:       pod.GetName(),
		SyncFolder:    syncFolder,
	}

	cmdKind := devfilev1.RunCommandGroupKind
	cmdName := parameters.StartOptions.RunCommand
	if parameters.StartOptions.Debug {
		cmdKind = devfilev1.DebugCommandGroupKind
		cmdName = parameters.StartOptions.DebugCommand
	}

	syncParams := sync.SyncParameters{
		Path:                     path,
		WatchFiles:               parameters.WatchFiles,
		WatchDeletedFiles:        parameters.WatchDeletedFiles,
		IgnoredFiles:             parameters.StartOptions.IgnorePaths,
		DevfileScanIndexForWatch: parameters.DevfileScanIndexForWatch,

		CompInfo:  compInfo,
		ForcePush: !o.deploymentExists || podChanged,
		Files:     common.GetSyncFilesFromAttributes(pushDevfileCommands[cmdKind]),
	}

	execRequired, err := o.syncClient.SyncFiles(ctx, syncParams)
	if err != nil {
		componentStatus.SetState(watch.StateReady)
		return fmt.Errorf("failed to sync to component with name %s: %w", componentName, err)
	}
	s.End(true)

	if !componentStatus.PostStartEventsDone && libdevfile.HasPostStartEvents(parameters.Devfile) {
		// PostStart events from the devfile will only be executed when the component
		// didn't previously exist
		handler := component.NewRunHandler(
			ctx,
			o.kubernetesClient,
			o.execClient,
			o.configAutomountClient,
			// TODO(feloy) set these values when we want to support Apply Image/Kubernetes/OpenShift commands for PostStart commands
			nil, nil,
			component.HandlerOptions{
				PodName:           pod.Name,
				ContainersRunning: component.GetContainersNames(pod),
				Msg:               "Executing post-start command in container",
			},
		)
		err = libdevfile.ExecPostStartEvents(ctx, parameters.Devfile, handler)
		if err != nil {
			return err
		}
	}
	componentStatus.PostStartEventsDone = true

	cmd, err := libdevfile.ValidateAndGetCommand(parameters.Devfile, cmdName, cmdKind)
	if err != nil {
		return err
	}

	commandType, err := parsercommon.GetCommandType(cmd)
	if err != nil {
		return err
	}
	var running bool
	var isComposite bool

	cmdHandler := component.NewRunHandler(
		ctx,
		o.kubernetesClient,
		o.execClient,
		o.configAutomountClient,
		o.filesystem,
		image.SelectBackend(ctx),
		component.HandlerOptions{
			PodName:           pod.GetName(),
			ContainersRunning: component.GetContainersNames(pod),
			Devfile:           parameters.Devfile,
			Path:              path,
		},
	)

	if commandType == devfilev1.ExecCommandType {
		running, err = cmdHandler.IsRemoteProcessForCommandRunning(ctx, cmd, pod.Name)
		if err != nil {
			return err
		}
	} else if commandType == devfilev1.CompositeCommandType {
		// this handler will run each command in this composite command individually,
		// and will determine whether each command is running or not.
		isComposite = true
	} else {
		return fmt.Errorf("unsupported type %q for Devfile command %s, only exec and composite are handled",
			commandType, cmd.Id)
	}

	cmdHandler.ComponentExists = running || isComposite

	klog.V(4).Infof("running=%v, execRequired=%v",
		running, execRequired)

	if isComposite || !running || execRequired {
		// Invoke the build command once (before calling libdevfile.ExecuteCommandByNameAndKind), as, if cmd is a composite command,
		// the handler we pass will be called for each command in that composite command.
		doExecuteBuildCommand := func() error {
			execHandler := component.NewRunHandler(
				ctx,
				o.kubernetesClient,
				o.execClient,
				o.configAutomountClient,
				// TODO(feloy) set these values when we want to support Apply Image/Kubernetes/OpenShift commands for PostStart commands
				nil, nil,
				component.HandlerOptions{
					PodName:           pod.Name,
					ComponentExists:   running,
					ContainersRunning: component.GetContainersNames(pod),
					Msg:               "Building your application in container",
				},
			)
			return libdevfile.Build(ctx, parameters.Devfile, parameters.StartOptions.BuildCommand, execHandler)
		}
		if err = doExecuteBuildCommand(); err != nil {
			componentStatus.SetState(watch.StateReady)
			return err
		}

		err = libdevfile.ExecuteCommandByNameAndKind(ctx, parameters.Devfile, cmdName, cmdKind, cmdHandler, false)
		if err != nil {
			return err
		}
	}

	if podChanged || o.portsChanged {
		o.portForwardClient.StopPortForwarding(ctx, componentName)
	}

	// Check that the application is actually listening on the ports declared in the Devfile, so we are sure that port-forwarding will work
	appReadySpinner := log.Spinner("Waiting for the application to be ready")
	err = o.checkAppPorts(ctx, pod.Name, o.portsToForward)
	appReadySpinner.End(err == nil)
	if err != nil {
		log.Warningf("Port forwarding might not work correctly: %v", err)
		log.Warning("Running `odo logs --follow` might help in identifying the problem.")
		fmt.Fprintln(log.GetStdout())
	}

	err = o.portForwardClient.StartPortForwarding(ctx, parameters.Devfile, componentName, parameters.StartOptions.Debug, parameters.StartOptions.RandomPorts, log.GetStdout(), parameters.StartOptions.ErrOut, parameters.StartOptions.CustomForwardedPorts, parameters.StartOptions.CustomAddress)
	if err != nil {
		return common.NewErrPortForward(err)
	}
	componentStatus.EndpointsForwarded = o.portForwardClient.GetForwardedPorts()

	componentStatus.SetState(watch.StateReady)
	return nil
}

func (o *DevClient) getPushDevfileCommands(parameters common.PushParameters) (map[devfilev1.CommandGroupKind]devfilev1.Command, error) {
	pushDevfileCommands, err := libdevfile.ValidateAndGetPushCommands(parameters.Devfile, parameters.StartOptions.BuildCommand, parameters.StartOptions.RunCommand)
	if err != nil {
		return nil, fmt.Errorf("failed to validate devfile build and run commands: %w", err)
	}

	if parameters.StartOptions.Debug {
		pushDevfileDebugCommands, e := libdevfile.ValidateAndGetCommand(parameters.Devfile, parameters.StartOptions.DebugCommand, devfilev1.DebugCommandGroupKind)
		if e != nil {
			return nil, fmt.Errorf("debug command is not valid: %w", e)
		}
		pushDevfileCommands[devfilev1.DebugCommandGroupKind] = pushDevfileDebugCommands
	}

	return pushDevfileCommands, nil
}

func (o *DevClient) checkAppPorts(ctx context.Context, podName string, portsToFwd map[string][]devfilev1.Endpoint) error {
	containerPortsMapping := make(map[string][]int)
	for c, ports := range portsToFwd {
		for _, p := range ports {
			containerPortsMapping[c] = append(containerPortsMapping[c], p.TargetPort)
		}
	}
	return port.CheckAppPortsListening(ctx, o.execClient, podName, containerPortsMapping, 1*time.Minute)
}
