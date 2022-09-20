package podmandev

import (
	"context"
	"fmt"
	"io"

	"github.com/devfile/library/pkg/devfile/parser"
	"github.com/redhat-developer/odo/pkg/dev"
	"github.com/redhat-developer/odo/pkg/podman"
)

type DevClient struct {
	podmanClient podman.Client
}

var _ dev.Client = (*DevClient)(nil)

func NewDevClient(podmanClient podman.Client) *DevClient {
	return &DevClient{
		podmanClient: podmanClient,
	}
}

func (o *DevClient) Start(
	ctx context.Context,
	devfileObj parser.DevfileObj,
	componentName string,
	path string,
	devfilePath string,
	out io.Writer,
	errOut io.Writer,
	options dev.StartOptions,
) error {
	fmt.Printf("Deploying using Podman\n\n")

	pod, err := createPodFromComponent(
		devfileObj,
		componentName,
		"app",
		options.BuildCommand,
		options.RunCommand,
		"",
	)
	if err != nil {
		return err
	}

	err = o.podmanClient.PlayKube(pod)
	if err != nil {
		return err
	}

	<-ctx.Done()

	fmt.Printf("Cleaning up resources\n")
	err = o.podmanClient.PodStop(pod.GetName())
	if err != nil {
		return err
	}
	err = o.podmanClient.PodRm(pod.GetName())
	if err != nil {
		return err
	}

	for _, volume := range pod.Spec.Volumes {
		err = o.podmanClient.VolumeRm(volume.Name)
		if err != nil {
			return err
		}
	}

	return nil
}
