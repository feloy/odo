package component

import (
	"errors"
	"os"
	"path"
	"path/filepath"
	"reflect"
	"testing"

	devfilepkg "github.com/devfile/api/v2/pkg/devfile"
	"github.com/devfile/library/pkg/devfile"
	"github.com/devfile/library/pkg/devfile/parser"
	devfileCtx "github.com/devfile/library/pkg/devfile/parser/context"
	"github.com/devfile/library/pkg/devfile/parser/data"
	"github.com/devfile/library/pkg/testingutil/filesystem"
	dfutil "github.com/devfile/library/pkg/util"
	"github.com/golang/mock/gomock"
	"github.com/kylelemons/godebug/pretty"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	"github.com/redhat-developer/odo/pkg/kclient"
	"github.com/redhat-developer/odo/pkg/labels"
	"github.com/redhat-developer/odo/pkg/testingutil"
	"github.com/redhat-developer/odo/pkg/util"

	"github.com/redhat-developer/odo/pkg/api"
)

func TestListAllClusterComponents(t *testing.T) {
	const odoVersion = "v3.0.0-beta3"
	res1 := getUnstructured("dep1", "deployment", "v1", "Unknown", "", "Unknown", "my-ns")
	res2 := getUnstructured("svc1", "service", "v1", "odo", odoVersion, "nodejs", "my-ns")
	res3 := getUnstructured("dep1", "deployment", "v1", "Unknown", "", "Unknown", "my-ns")
	res3.SetLabels(map[string]string{})

	commonLabels := labels.Builder().WithComponentName("comp1").WithManager("odo").WithManagedByVersion(odoVersion)

	resDev := getUnstructured("depDev", "deployment", "v1", "odo", odoVersion, "nodejs", "my-ns")
	labelsDev := commonLabels.WithMode("Dev").Labels()
	resDev.SetLabels(labelsDev)

	resDeploy := getUnstructured("depDeploy", "deployment", "v1", "odo", odoVersion, "nodejs", "my-ns")
	labelsDeploy := commonLabels.WithMode("Deploy").Labels()
	resDeploy.SetLabels(labelsDeploy)

	type fields struct {
		kubeClient func(ctrl *gomock.Controller) kclient.ClientInterface
	}
	type args struct {
		namespace string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []api.ComponentAbstract
		wantErr bool
	}{
		{
			name: "1 non-odo resource returned with Unknown",
			fields: fields{
				kubeClient: func(ctrl *gomock.Controller) kclient.ClientInterface {
					var resources []unstructured.Unstructured
					resources = append(resources, res1)
					client := kclient.NewMockClientInterface(ctrl)
					selector := ""
					client.EXPECT().GetAllResourcesFromSelector(selector, "my-ns").Return(resources, nil)
					return client
				},
			},
			args: args{
				namespace: "my-ns",
			},
			want: []api.ComponentAbstract{{
				Name:             "dep1",
				ManagedBy:        "Unknown",
				ManagedByVersion: "",
				RunningIn:        nil,
				Type:             "Unknown",
			}},
			wantErr: false,
		},
		{
			name: "0 non-odo resource without instance label is not returned",
			fields: fields{
				kubeClient: func(ctrl *gomock.Controller) kclient.ClientInterface {
					var resources []unstructured.Unstructured
					resources = append(resources, res3)
					client := kclient.NewMockClientInterface(ctrl)
					client.EXPECT().GetAllResourcesFromSelector(gomock.Any(), "my-ns").Return(resources, nil)
					return client
				},
			},
			args: args{
				namespace: "my-ns",
			},
			want:    nil,
			wantErr: false,
		},
		{
			name: "1 non-odo resource returned with Unknown, and 1 odo resource returned with odo",
			fields: fields{
				kubeClient: func(ctrl *gomock.Controller) kclient.ClientInterface {
					var resources []unstructured.Unstructured
					resources = append(resources, res1, res2)
					client := kclient.NewMockClientInterface(ctrl)
					client.EXPECT().GetAllResourcesFromSelector(gomock.Any(), "my-ns").Return(resources, nil)
					return client
				},
			},
			args: args{
				namespace: "my-ns",
			},
			want: []api.ComponentAbstract{{
				Name:             "dep1",
				ManagedBy:        "Unknown",
				ManagedByVersion: "",
				RunningIn:        nil,
				Type:             "Unknown",
			}, {
				Name:             "svc1",
				ManagedBy:        "odo",
				ManagedByVersion: "v3.0.0-beta3",
				RunningIn:        nil,
				Type:             "nodejs",
			}},
			wantErr: false,
		},
		{
			name: "one resource in Dev and Deploy modes",
			fields: fields{
				kubeClient: func(ctrl *gomock.Controller) kclient.ClientInterface {
					var resources []unstructured.Unstructured
					resources = append(resources, resDev, resDeploy)
					client := kclient.NewMockClientInterface(ctrl)
					client.EXPECT().GetAllResourcesFromSelector(gomock.Any(), "my-ns").Return(resources, nil)
					return client
				},
			},
			args: args{
				namespace: "my-ns",
			},
			want: []api.ComponentAbstract{{
				Name:             "comp1",
				ManagedBy:        "odo",
				ManagedByVersion: "v3.0.0-beta3",
				RunningIn: api.RunningModeList{
					"dev":    true,
					"deploy": true,
				},
				Type: "nodejs",
			}},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			got, err := ListAllClusterComponents(tt.fields.kubeClient(ctrl), tt.args.namespace)
			if (err != nil) != tt.wantErr {
				t.Errorf("ListAllClusterComponents error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ListAllClusterComponents got = %+v\nwant = %+v\ncomparison:\n %v", got, tt.want, pretty.Compare(got, tt.want))
			}
		})
	}
}

func TestGetComponentTypeFromDevfileMetadata(t *testing.T) {
	tests := []devfilepkg.DevfileMetadata{
		{
			Name:        "ReturnProject",
			ProjectType: "Maven",
			Language:    "Java",
		},
		{
			Name:     "ReturnLanguage",
			Language: "Java",
		},
		{
			Name: "ReturnNA",
		},
	}
	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			var want string
			got := GetComponentTypeFromDevfileMetadata(tt)
			switch tt.Name {
			case "ReturnProject":
				want = tt.ProjectType
			case "ReturnLanguage":
				want = tt.Language
			case "ReturnNA":
				want = NotAvailable
			}
			if got != want {
				t.Errorf("Incorrect component type returned; got: %q, want: %q", got, want)
			}
		})
	}
}

// getUnstructured returns an unstructured.Unstructured object
func getUnstructured(name, kind, apiVersion, managed, managedByVersion, componentType, namespace string) (u unstructured.Unstructured) {
	u.SetName(name)
	u.SetKind(kind)
	u.SetAPIVersion(apiVersion)
	u.SetNamespace(namespace)
	u.SetLabels(labels.Builder().
		WithComponentName(name).
		WithManager(managed).
		WithManagedByVersion(managedByVersion).
		Labels())
	u.SetAnnotations(labels.Builder().
		WithProjectType(componentType).
		Labels())
	return
}

func TestGetRunningModes(t *testing.T) {

	resourceDev1 := unstructured.Unstructured{}
	resourceDev1.SetLabels(labels.Builder().WithMode(labels.ComponentDevMode).Labels())

	resourceDev2 := unstructured.Unstructured{}
	resourceDev2.SetLabels(labels.Builder().WithMode(labels.ComponentDevMode).Labels())

	resourceDeploy1 := unstructured.Unstructured{}
	resourceDeploy1.SetLabels(labels.Builder().WithMode(labels.ComponentDeployMode).Labels())

	resourceDeploy2 := unstructured.Unstructured{}
	resourceDeploy2.SetLabels(labels.Builder().WithMode(labels.ComponentDeployMode).Labels())

	otherResource := unstructured.Unstructured{}

	packageManifestResource := unstructured.Unstructured{}
	packageManifestResource.SetKind("PackageManifest")
	packageManifestResource.SetLabels(labels.Builder().WithMode(labels.ComponentDevMode).Labels())

	type args struct {
		client func(ctrl *gomock.Controller) kclient.ClientInterface
		name   string
	}
	tests := []struct {
		name    string
		args    args
		want    api.RunningModeList
		wantErr bool
	}{
		{
			name: "No resources",
			args: args{
				client: func(ctrl *gomock.Controller) kclient.ClientInterface {
					c := kclient.NewMockClientInterface(ctrl)
					c.EXPECT().GetCurrentNamespace().Return("a-namespace").AnyTimes()
					c.EXPECT().GetAllResourcesFromSelector(gomock.Any(), gomock.Any()).Return([]unstructured.Unstructured{}, nil)
					return c
				},
				name: "aname",
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "Only PackageManifest resource",
			args: args{
				client: func(ctrl *gomock.Controller) kclient.ClientInterface {
					c := kclient.NewMockClientInterface(ctrl)
					c.EXPECT().GetCurrentNamespace().Return("a-namespace").AnyTimes()
					c.EXPECT().GetAllResourcesFromSelector(gomock.Any(), gomock.Any()).Return([]unstructured.Unstructured{packageManifestResource}, nil)
					return c
				},
				name: "aname",
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "No dev/deploy resources",
			args: args{
				client: func(ctrl *gomock.Controller) kclient.ClientInterface {
					c := kclient.NewMockClientInterface(ctrl)
					c.EXPECT().GetCurrentNamespace().Return("a-namespace").AnyTimes()
					c.EXPECT().GetAllResourcesFromSelector(gomock.Any(), gomock.Any()).Return([]unstructured.Unstructured{packageManifestResource, otherResource}, nil)
					return c
				},
				name: "aname",
			},
			want: api.RunningModeList{"dev": false, "deploy": false},
		},
		{
			name: "Only Dev resources",
			args: args{
				client: func(ctrl *gomock.Controller) kclient.ClientInterface {
					c := kclient.NewMockClientInterface(ctrl)
					c.EXPECT().GetCurrentNamespace().Return("a-namespace").AnyTimes()
					c.EXPECT().GetAllResourcesFromSelector(gomock.Any(), gomock.Any()).Return([]unstructured.Unstructured{packageManifestResource, otherResource, resourceDev1, resourceDev2}, nil)
					return c
				},
				name: "aname",
			},
			want: api.RunningModeList{"dev": true, "deploy": false},
		},
		{
			name: "Only Deploy resources",
			args: args{
				client: func(ctrl *gomock.Controller) kclient.ClientInterface {
					c := kclient.NewMockClientInterface(ctrl)
					c.EXPECT().GetCurrentNamespace().Return("a-namespace").AnyTimes()
					c.EXPECT().GetAllResourcesFromSelector(gomock.Any(), gomock.Any()).Return([]unstructured.Unstructured{packageManifestResource, otherResource, resourceDeploy1, resourceDeploy2}, nil)
					return c
				},
				name: "aname",
			},
			want: api.RunningModeList{"dev": false, "deploy": true},
		},
		{
			name: "Dev and Deploy resources",
			args: args{
				client: func(ctrl *gomock.Controller) kclient.ClientInterface {
					c := kclient.NewMockClientInterface(ctrl)
					c.EXPECT().GetCurrentNamespace().Return("a-namespace").AnyTimes()
					c.EXPECT().GetAllResourcesFromSelector(gomock.Any(), gomock.Any()).Return([]unstructured.Unstructured{packageManifestResource, otherResource, resourceDev1, resourceDev2, resourceDeploy1, resourceDeploy2}, nil)
					return c
				},
				name: "aname",
			},
			want: api.RunningModeList{"dev": true, "deploy": true},
		},
		{
			name: "Unknown",
			args: args{
				client: func(ctrl *gomock.Controller) kclient.ClientInterface {
					c := kclient.NewMockClientInterface(ctrl)
					c.EXPECT().GetCurrentNamespace().Return("a-namespace").AnyTimes()
					c.EXPECT().GetAllResourcesFromSelector(gomock.Any(), gomock.Any()).Return(nil, errors.New("error"))
					return c
				},
				name: "aname",
			},
			want: api.RunningModeList{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			got, err := GetRunningModes(tt.args.client(ctrl), tt.args.name)
			if (err != nil) != tt.wantErr {
				t.Errorf("error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetRunningModes() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGatherName(t *testing.T) {
	type devfileProvider func() (*parser.DevfileObj, string, error)
	fakeDevfileWithNameProvider := func(name string) devfileProvider {
		return func() (*parser.DevfileObj, string, error) {
			dData, err := data.NewDevfileData(string(data.APISchemaVersion220))
			if err != nil {
				return nil, "", err
			}
			dData.SetMetadata(devfilepkg.DevfileMetadata{Name: name})
			return &parser.DevfileObj{
				Ctx:  devfileCtx.FakeContext(filesystem.NewFakeFs(), parser.OutputDevfileYamlPath),
				Data: dData,
			}, "", nil
		}
	}

	fs := filesystem.DefaultFs{}
	//realDevfileWithNameProvider creates a real temporary directory and writes a devfile with the given name to it.
	//It is the responsibility of the caller to remove the directory.
	realDevfileWithNameProvider := func(name string) devfileProvider {
		return func() (*parser.DevfileObj, string, error) {
			dir, err := fs.TempDir("", "Component_GatherName_")
			if err != nil {
				return nil, dir, err
			}

			originalDevfile := testingutil.GetTestDevfileObjFromFile("devfile.yaml")
			originalDevfilePath := originalDevfile.Ctx.GetAbsPath()

			stat, err := os.Stat(originalDevfilePath)
			if err != nil {
				return nil, dir, err
			}
			dPath := path.Join(dir, "devfile.yaml")
			err = dfutil.CopyFile(originalDevfilePath, dPath, stat)
			if err != nil {
				return nil, dir, err
			}

			var d parser.DevfileObj
			d, _, err = devfile.ParseDevfileAndValidate(parser.ParserArgs{Path: dPath})
			if err != nil {
				return nil, dir, err
			}

			err = d.SetMetadataName(name)

			return &d, dir, err
		}
	}

	wantDevfileDirectoryName := func(contextDir string, d *parser.DevfileObj) string {
		return util.GetDNS1123Name(filepath.Base(filepath.Dir(d.Ctx.GetAbsPath())))
	}

	for _, tt := range []struct {
		name                string
		devfileProviderFunc devfileProvider
		wantErr             bool
		want                func(contextDir string, d *parser.DevfileObj) string
	}{
		{
			name:                "compliant name",
			devfileProviderFunc: fakeDevfileWithNameProvider("my-component-name"),
			want:                func(contextDir string, d *parser.DevfileObj) string { return "my-component-name" },
		},
		{
			name:                "un-sanitized name",
			devfileProviderFunc: fakeDevfileWithNameProvider("name with spaces"),
			want:                func(contextDir string, d *parser.DevfileObj) string { return "name-with-spaces" },
		},
		{
			name:                "all numeric name",
			devfileProviderFunc: fakeDevfileWithNameProvider("123456789"),
			// "x" prefix added by util.GetDNS1123Name
			want: func(contextDir string, d *parser.DevfileObj) string { return "x123456789" },
		},
		{
			name:                "no name",
			devfileProviderFunc: realDevfileWithNameProvider(""),
			want:                wantDevfileDirectoryName,
		},
		{
			name:                "blank name",
			devfileProviderFunc: realDevfileWithNameProvider("   "),
			want:                wantDevfileDirectoryName,
		},
		{
			name: "passing no devfile should use the context directory name",
			devfileProviderFunc: func() (*parser.DevfileObj, string, error) {
				dir, err := fs.TempDir("", "Component_GatherName_")
				if err != nil {
					return nil, dir, err
				}
				return nil, dir, nil
			},
			want: func(contextDir string, _ *parser.DevfileObj) string {
				return util.GetDNS1123Name(filepath.Base(contextDir))
			},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			d, dir, dErr := tt.devfileProviderFunc()
			if dir != "" {
				defer func(fs filesystem.Filesystem, path string) {
					if err := fs.RemoveAll(path); err != nil {
						t.Logf("error while attempting to remove temporary directory %q: %v", path, err)
					}
				}(fs, dir)
			}
			if dErr != nil {
				t.Errorf("error when building test Devfile object: %v", dErr)
				return
			}

			got, err := GatherName(dir, d)
			if (err != nil) != tt.wantErr {
				t.Errorf("error = %v, wantErr %v", err, tt.wantErr)
			}
			want := tt.want(dir, d)
			if !reflect.DeepEqual(got, want) {
				t.Errorf("GatherName() = %q, want = %q", got, want)
			}
		})
	}
}
