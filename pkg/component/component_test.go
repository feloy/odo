package component

import (
	"errors"
	"reflect"
	"testing"

	devfilepkg "github.com/devfile/api/v2/pkg/devfile"
	"github.com/golang/mock/gomock"
	"github.com/kylelemons/godebug/pretty"

	"github.com/redhat-developer/odo/pkg/kclient"
	"github.com/redhat-developer/odo/pkg/labels"

	"github.com/redhat-developer/odo/pkg/api"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func TestListAllClusterComponents(t *testing.T) {
	res1 := getUnstructured("dep1", "deployment", "v1", "Unknown", "Unknown", "my-ns")
	res2 := getUnstructured("svc1", "service", "v1", "odo", "nodejs", "my-ns")

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
		want    []OdoComponent
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
			want: []OdoComponent{{
				Name:      "dep1",
				ManagedBy: "Unknown",
				Modes:     map[string]bool{},
				Type:      "Unknown",
			}},
			wantErr: false,
		},
		{
			name: "1 non-odo resource returned with Unknown, and 1 odo resource returned with odo",
			fields: fields{
				kubeClient: func(ctrl *gomock.Controller) kclient.ClientInterface {
					var resources []unstructured.Unstructured
					resources = append(resources, res1, res2)
					client := kclient.NewMockClientInterface(ctrl)
					selector := ""
					client.EXPECT().GetAllResourcesFromSelector(selector, "my-ns").Return(resources, nil)
					return client
				},
			},
			args: args{
				namespace: "my-ns",
			},
			want: []OdoComponent{{
				Name:      "dep1",
				ManagedBy: "Unknown",
				Modes:     map[string]bool{},
				Type:      "Unknown",
			}, {
				Name:      "svc1",
				ManagedBy: "odo",
				Modes:     map[string]bool{},
				Type:      "nodejs",
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

func Test_getMachineReadableFormat(t *testing.T) {
	type args struct {
		componentName string
		componentType string
	}
	tests := []struct {
		name string
		args args
		want Component
	}{
		{
			name: "Test: Machine Readable Output",
			args: args{componentName: "frontend", componentType: "nodejs"},
			want: Component{
				TypeMeta: metav1.TypeMeta{
					Kind:       "Component",
					APIVersion: "odo.dev/v1alpha1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: "frontend",
				},
				Spec: ComponentSpec{
					Type: "nodejs",
				},
				Status: ComponentStatus{},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := newComponentWithType(tt.args.componentName, tt.args.componentType); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getMachineReadableFormat() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_getMachineReadableFormatForList(t *testing.T) {
	type args struct {
		components []Component
	}
	tests := []struct {
		name string
		args args
		want ComponentList
	}{
		{
			name: "Test: machine readable output for list",
			args: args{
				components: []Component{
					{
						TypeMeta: metav1.TypeMeta{
							Kind:       "Component",
							APIVersion: "odo.dev/v1alpha1",
						},
						ObjectMeta: metav1.ObjectMeta{
							Name: "frontend",
						},
						Spec: ComponentSpec{
							Type: "nodejs",
						},
						Status: ComponentStatus{},
					},
					{
						TypeMeta: metav1.TypeMeta{
							Kind:       "Component",
							APIVersion: "odo.dev/v1alpha1",
						},
						ObjectMeta: metav1.ObjectMeta{
							Name: "backend",
						},
						Spec: ComponentSpec{
							Type: "wildfly",
						},
						Status: ComponentStatus{},
					},
				},
			},
			want: ComponentList{
				TypeMeta: metav1.TypeMeta{
					Kind:       "List",
					APIVersion: "odo.dev/v1alpha1",
				},
				ListMeta: metav1.ListMeta{},
				Items: []Component{
					{
						TypeMeta: metav1.TypeMeta{
							Kind:       "Component",
							APIVersion: "odo.dev/v1alpha1",
						},
						ObjectMeta: metav1.ObjectMeta{
							Name: "frontend",
						},
						Spec: ComponentSpec{
							Type: "nodejs",
						},
						Status: ComponentStatus{},
					},
					{
						TypeMeta: metav1.TypeMeta{
							Kind:       "Component",
							APIVersion: "odo.dev/v1alpha1",
						},
						ObjectMeta: metav1.ObjectMeta{
							Name: "backend",
						},
						Spec: ComponentSpec{
							Type: "wildfly",
						},
						Status: ComponentStatus{},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := newComponentList(tt.args.components); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getMachineReadableFormatForList() = %v, want %v", got, tt.want)
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
func getUnstructured(name, kind, apiVersion, managed, componentType, namespace string) (u unstructured.Unstructured) {
	u.SetName(name)
	u.SetKind(kind)
	u.SetAPIVersion(apiVersion)
	u.SetNamespace(namespace)
	u.SetLabels(labels.Builder().
		WithComponentName(name).
		WithManager(managed).
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
		client    func(ctrl *gomock.Controller) kclient.ClientInterface
		name      string
		namespace string
	}
	tests := []struct {
		name    string
		args    args
		want    []api.RunningMode
		wantErr bool
	}{
		{
			name: "No resources",
			args: args{
				client: func(ctrl *gomock.Controller) kclient.ClientInterface {
					c := kclient.NewMockClientInterface(ctrl)
					c.EXPECT().GetAllResourcesFromSelector(gomock.Any(), gomock.Any()).Return([]unstructured.Unstructured{packageManifestResource, otherResource}, nil)
					return c
				},
				name:      "aname",
				namespace: "anamespace",
			},
			want: []api.RunningMode{},
		},
		{
			name: "Only Dev resources",
			args: args{
				client: func(ctrl *gomock.Controller) kclient.ClientInterface {
					c := kclient.NewMockClientInterface(ctrl)
					c.EXPECT().GetAllResourcesFromSelector(gomock.Any(), gomock.Any()).Return([]unstructured.Unstructured{packageManifestResource, otherResource, resourceDev1, resourceDev2}, nil)
					return c
				},
				name:      "aname",
				namespace: "anamespace",
			},
			want: []api.RunningMode{api.RunningModeDev},
		},
		{
			name: "Only Deploy resources",
			args: args{
				client: func(ctrl *gomock.Controller) kclient.ClientInterface {
					c := kclient.NewMockClientInterface(ctrl)
					c.EXPECT().GetAllResourcesFromSelector(gomock.Any(), gomock.Any()).Return([]unstructured.Unstructured{packageManifestResource, otherResource, resourceDeploy1, resourceDeploy2}, nil)
					return c
				},
				name:      "aname",
				namespace: "anamespace",
			},
			want: []api.RunningMode{api.RunningModeDeploy},
		},
		{
			name: "Dev and Deploy resources",
			args: args{
				client: func(ctrl *gomock.Controller) kclient.ClientInterface {
					c := kclient.NewMockClientInterface(ctrl)
					c.EXPECT().GetAllResourcesFromSelector(gomock.Any(), gomock.Any()).Return([]unstructured.Unstructured{packageManifestResource, otherResource, resourceDev1, resourceDev2, resourceDeploy1, resourceDeploy2}, nil)
					return c
				},
				name:      "aname",
				namespace: "anamespace",
			},
			want: []api.RunningMode{api.RunningModeDev, api.RunningModeDeploy},
		},
		{
			name: "Unknown",
			args: args{
				client: func(ctrl *gomock.Controller) kclient.ClientInterface {
					c := kclient.NewMockClientInterface(ctrl)
					c.EXPECT().GetAllResourcesFromSelector(gomock.Any(), gomock.Any()).Return(nil, errors.New("error"))
					return c
				},
				name:      "aname",
				namespace: "anamespace",
			},
			want: []api.RunningMode{api.RunningModeUnknown},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			got, err := GetRunningModes(tt.args.client(ctrl), tt.args.name, tt.args.namespace)
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
