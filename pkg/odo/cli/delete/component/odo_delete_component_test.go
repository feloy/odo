package component

import (
	"bytes"
	"context"
	"path/filepath"
	"strings"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/redhat-developer/odo/pkg/devfile"
	"github.com/redhat-developer/odo/pkg/kclient"
	"github.com/redhat-developer/odo/pkg/odo/cmdline"
	"github.com/redhat-developer/odo/pkg/odo/commonflags"
	odocontext "github.com/redhat-developer/odo/pkg/odo/context"
	"github.com/redhat-developer/odo/pkg/odo/genericclioptions/clientset"
	"github.com/redhat-developer/odo/pkg/testingutil/filesystem"

	"github.com/spf13/cobra"

	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/api/meta"
)

const (
	_workdir = "/home/user/workdir"
)

func TestOdoDeleteComponent(t *testing.T) {
	type flags struct {
		name          string
		namespace     string
		withFilesFlag bool
		forceFlag     bool
		waitFlag      bool
	}

	tests := []struct {
		name           string
		flags          flags
		ctx            func(ctx context.Context) (context.Context, error)
		args           []string
		wantErr        bool
		checkErr       func(err error) bool
		checkOutputs   func(stdout, stderr string) bool
		mockKubeClient func(kclient *kclient.MockClientInterface)
	}{
		{
			name: "empty directory without --name arg should fail",
			ctx: func(ctx context.Context) (context.Context, error) {
				ctx = odocontext.WithWorkingDirectory(ctx, _workdir)
				ctx = odocontext.WithDevfileObj(ctx, nil)
				return ctx, nil
			},
			flags: flags{
				forceFlag: true,
			},
			wantErr: true,
			checkErr: func(err error) bool {
				str := err.Error()
				return strings.Contains(str, "The current directory does not represent an odo component")
			},
		},
		{
			name: "empty directory without --name arg, with --files should fail",
			ctx: func(ctx context.Context) (context.Context, error) {
				ctx = odocontext.WithWorkingDirectory(ctx, _workdir)
				ctx = odocontext.WithDevfileObj(ctx, nil)
				return ctx, nil
			},
			flags: flags{
				forceFlag:     true,
				withFilesFlag: true,
			},
			wantErr: true,
			checkErr: func(err error) bool {
				str := err.Error()
				return strings.Contains(str, "The current directory does not represent an odo component")
			},
		},
		{
			name: "using both --name and --files should fail",
			ctx: func(ctx context.Context) (context.Context, error) {
				ctx = odocontext.WithWorkingDirectory(ctx, _workdir)
				ctx = odocontext.WithDevfileObj(ctx, nil)
				return ctx, nil
			},
			flags: flags{
				forceFlag:     true,
				name:          "mycomp",
				withFilesFlag: true,
			},
			wantErr: true,
			checkErr: func(err error) bool {
				str := err.Error()
				return strings.Contains(str, "'--files' cannot be used with '--name'; '--files' must be used from a directory containing a Devfile")
			},
			mockKubeClient: func(kclient *kclient.MockClientInterface) {
				kclient.EXPECT().GetCurrentNamespace().Times(1).Return("ns1")
			},
		},
		{
			name: "with a Devfile, should delete the deployment",
			ctx: func(ctx context.Context) (context.Context, error) {
				ctx = odocontext.WithWorkingDirectory(ctx, _workdir)
				ctx = odocontext.WithComponentName(ctx, "nodejs")
				ctx = odocontext.WithApplication(ctx, "app")
				ctx = odocontext.WithNamespace(ctx, "ns1")
				// When: a devfile is present in current directory
				devfile, err := devfile.ParseAndValidateFromFile(filepath.Join("..", "..", "..", "..", "..", "tests", "examples", "source", "devfiles", "nodejs", "devfile.yaml"))
				if err != nil {
					return nil, err
				}
				ctx = odocontext.WithDevfileObj(ctx, &devfile)
				return ctx, nil
			},
			flags: flags{
				forceFlag:     true,
				withFilesFlag: true,
			},
			wantErr: false,
			checkOutputs: func(stdout, stderr string) bool {
				if !strings.Contains(stdout, `No resource found for component "nodejs-app" in namespace "ns1"`) {
					return false
				}
				if len(stderr) > 0 {
					return false
				}
				return true
			},
			mockKubeClient: func(kclient *kclient.MockClientInterface) {
				deploymentName := "nodejs-app"
				deployment := appsv1.Deployment{}
				deployment.SetName(deploymentName)
				// When: A deployment is deployed
				kclient.EXPECT().GetDeploymentByName(deploymentName).Return(&deployment, nil)
				// When: No other resources are deployed
				kclient.EXPECT().GetAllResourcesFromSelector(gomock.Any(), "ns1").Return(nil, nil)

				// Then: The deployment should be deleted
				gvr := appsv1.SchemeGroupVersion.WithResource("Deployment")
				kclient.EXPECT().GetRestMappingFromUnstructured(gomock.Any()).Times(1).Return(&meta.RESTMapping{
					Resource: gvr,
				}, nil)
				kclient.EXPECT().DeleteDynamicResource(deploymentName, gvr, false).Times(1).Return(nil)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			ctrl := gomock.NewController(t)
			cmd := &cobra.Command{}
			clientset.Add(cmd, clientset.DELETE_COMPONENT, clientset.KUBERNETES, clientset.FILESYSTEM)

			mockKubeClient := kclient.NewMockClientInterface(ctrl)
			if tt.mockKubeClient != nil {
				tt.mockKubeClient(mockKubeClient)
			}
			deps, err := clientset.Fetch(cmd, commonflags.RunOnCluster, clientset.Clientset{
				FS:               filesystem.NewFakeFs(),
				KubernetesClient: mockKubeClient,
			})
			if err != nil {
				t.Errorf("unexpected err %v", err)
			}

			err = deps.FS.MkdirAll(_workdir, 0644)
			if err != nil {
				t.Errorf("unexpected err %v", err)
			}

			o := &ComponentOptions{
				name:          tt.flags.name,
				namespace:     tt.flags.namespace,
				withFilesFlag: tt.flags.withFilesFlag,
				forceFlag:     tt.flags.forceFlag,
				waitFlag:      tt.flags.waitFlag,
				clientset:     deps,
			}

			ctx, err := tt.ctx(context.Background())
			var stdout, stderr bytes.Buffer
			ctx = odocontext.WithStdout(ctx, &stdout)
			ctx = odocontext.WithStderr(ctx, &stderr)
			if err != nil {
				t.Errorf("unexpected error when building context: %v", err)
			}
			cmdLineObj := cmdline.NewCobra(cmd)

			err = o.Complete(ctx, cmdLineObj, tt.args)
			if err != nil {
				if !tt.wantErr {
					t.Errorf("unexpected error: %v", err)
				}
				if !tt.checkErr(err) {
					t.Errorf("unexpected error: %v", err)
				}
				if tt.checkOutputs != nil {
					tt.checkOutputs(stdout.String(), stderr.String())
				}
				return
			}

			err = o.Validate(ctx)
			if err != nil {
				if !tt.wantErr {
					t.Errorf("unexpected error: %v", err)
				}
				if !tt.checkErr(err) {
					t.Errorf("unexpected error: %v", err)
				}
				if tt.checkOutputs != nil {
					tt.checkOutputs(stdout.String(), stderr.String())
				}
				return
			}

			err = o.Run(ctx)
			if err != nil {
				if !tt.wantErr {
					t.Errorf("unexpected error: %v", err)
				}
				if !tt.checkErr(err) {
					t.Errorf("unexpected error: %v", err)
				}
				if tt.checkOutputs != nil {
					tt.checkOutputs(stdout.String(), stderr.String())
				}
				return
			}

			if tt.wantErr {
				t.Errorf("no error happened, but error is expected")
			}

			if tt.checkOutputs != nil {
				tt.checkOutputs(stdout.String(), stderr.String())
			}
		})

	}
}
