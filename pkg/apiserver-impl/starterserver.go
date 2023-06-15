package apiserver_impl

import (
	"context"
	"fmt"
	"net"
	"net/http"

	openapi "github.com/redhat-developer/odo/pkg/apiserver-gen/go"
	"github.com/redhat-developer/odo/pkg/kclient"
	"github.com/redhat-developer/odo/pkg/podman"
	"github.com/redhat-developer/odo/pkg/state"
	"github.com/redhat-developer/odo/pkg/util"
	"k8s.io/klog"
)

func StartServer(
	ctx context.Context,
	cancelFunc context.CancelFunc,
	port int,
	kubernetesClient kclient.ClientInterface,
	podmanClient podman.Client,
	stateClient state.Client,
) {

	DefaultApiService := NewDefaultApiService(
		cancelFunc,
		kubernetesClient,
		podmanClient,
		stateClient,
	)
	DefaultApiController := openapi.NewDefaultAPIController(DefaultApiService)

	router := openapi.NewRouter(DefaultApiController)

	var err error

	if port == 0 {
		port, err = util.NextFreePort(20000, 30001, nil, "")
		if err != nil {
			klog.V(0).Infof("Unable to start the API server; encountered error: %v", err)
			cancelFunc()
		}
	}

	err = stateClient.SetAPIServerPort(ctx, port)
	if err != nil {
		klog.V(0).Infof("Unable to start the API server; encountered error: %v", err)
		cancelFunc()
	}

	klog.V(0).Infof("API Server started at localhost:%d/api/v1", port)

	server := &http.Server{Addr: fmt.Sprintf(":%d", port), Handler: router}
	var errChan = make(chan error)
	go func() {
		server.BaseContext = func(net.Listener) context.Context {
			return ctx
		}
		err = server.ListenAndServe()
		errChan <- err
	}()
	go func() {
		select {
		case <-ctx.Done():
			klog.V(0).Infof("Shutting down the API server: %v", ctx.Err())
			err = server.Shutdown(ctx)
			if err != nil {
				klog.V(1).Infof("Error while shutting down the API server: %v", err)
			}
		case err = <-errChan:
			klog.V(0).Infof("Stopping the API server; encountered error: %v", err)
			cancelFunc()
		}
	}()
}
