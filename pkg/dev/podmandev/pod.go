package podmandev

import (
	"fmt"

	"github.com/devfile/library/pkg/devfile/generator"
	"github.com/devfile/library/pkg/devfile/parser"
	"github.com/devfile/library/pkg/devfile/parser/data/v2/common"
	"github.com/redhat-developer/odo/pkg/api"
	"github.com/redhat-developer/odo/pkg/component"
	"github.com/redhat-developer/odo/pkg/devfile/adapters/kubernetes/utils"
	"github.com/redhat-developer/odo/pkg/labels"
	"github.com/redhat-developer/odo/pkg/storage"
	"github.com/redhat-developer/odo/pkg/util"

	corev1 "k8s.io/api/core/v1"
)

func createPodFromComponent(
	devfileObj parser.DevfileObj,
	componentName string,
	appName string,
	buildCommand string,
	runCommand string,
	debugCommand string,
) (*corev1.Pod, []api.ForwardedPort, error) {
	containers, err := generator.GetContainers(devfileObj, common.DevfileOptions{})
	if err != nil {
		return nil, nil, err
	}
	if len(containers) == 0 {
		return nil, nil, fmt.Errorf("no valid components found in the devfile")
	}

	containers, err = utils.UpdateContainersEntrypointsIfNeeded(devfileObj, containers, buildCommand, runCommand, debugCommand)
	if err != nil {
		return nil, nil, err
	}
	utils.AddOdoProjectVolume(&containers)
	utils.AddOdoMandatoryVolume(&containers)

	fwPorts := addHostPorts(containers)

	volumes := []corev1.Volume{
		{
			Name: storage.OdoSourceVolume,
			VolumeSource: corev1.VolumeSource{
				PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
					ClaimName: getVolumeName(componentName, appName, "source"),
				},
			},
		},
		{
			Name: storage.SharedDataVolumeName,
			VolumeSource: corev1.VolumeSource{
				PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
					ClaimName: getVolumeName(componentName, appName, "shared"),
				},
			},
		},
	}

	// TODO add labels (for GetRunningPodFromSelector)
	pod := corev1.Pod{
		Spec: corev1.PodSpec{
			Containers: containers,
			Volumes:    volumes,
		},
	}

	pod.APIVersion, pod.Kind = corev1.SchemeGroupVersion.WithKind("Pod").ToAPIVersionAndKind()
	name, err := util.NamespaceKubernetesObject(componentName, appName)
	if err != nil {
		return nil, nil, err
	}
	pod.SetName(name)

	runtime := component.GetComponentRuntimeFromDevfileMetadata(devfileObj.Data.GetMetadata())
	pod.SetLabels(labels.GetLabels(componentName, appName, runtime, labels.ComponentDevMode, true))

	return &pod, fwPorts, nil
}

func getVolumeName(componentName string, appName string, volume string) string {
	return "odo-projects-" + componentName + "-" + appName + "-" + volume
}

func addHostPorts(containers []corev1.Container) []api.ForwardedPort {
	result := []api.ForwardedPort{}
	hostPort := int32(39001)
	for i := range containers {
		for j := range containers[i].Ports {
			result = append(result, api.ForwardedPort{
				ContainerName: containers[i].Name,
				LocalAddress:  "127.0.0.1",
				LocalPort:     int(hostPort),
				ContainerPort: int(containers[i].Ports[j].ContainerPort),
			})
			containers[i].Ports[j].HostPort = hostPort
			hostPort++
		}
	}
	return result
}
