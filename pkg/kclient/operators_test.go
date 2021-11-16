package kclient

import (
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"testing"

	devfile "github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
	v1alpha2 "github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
	"github.com/devfile/library/pkg/devfile/parser"
	devfileCtx "github.com/devfile/library/pkg/devfile/parser/context"
	"github.com/devfile/library/pkg/devfile/parser/data"
	devfileFileSystem "github.com/devfile/library/pkg/testingutil/filesystem"

	"github.com/ghodss/yaml"
	"github.com/go-openapi/spec"
	gomock "github.com/golang/mock/gomock"
	olm "github.com/operator-framework/api/pkg/operators/v1alpha1"
	"github.com/pkg/errors"

	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func TestGetResourceSpecDefinitionFromSwagger(t *testing.T) {
	tests := []struct {
		name    string
		swagger []byte
		group   string
		version string
		kind    string
		want    *spec.Schema
		wantErr bool
	}{
		{
			name:    "not found CRD",
			swagger: []byte("{}"),
			group:   "aGroup",
			version: "aVersion",
			kind:    "aKind",
			want:    nil,
			wantErr: true,
		},
		{
			name: "found CRD without spec",
			swagger: []byte(`{
  "definitions": {
	"com.dev4devs.postgresql.v1alpha1.Database": {
		"type": "object",
		"x-kubernetes-group-version-kind": [
			{
				"group": "postgresql.dev4devs.com",
				"kind": "Database",
				"version": "v1alpha1"
			}
		]
	}
  }
}`),
			group:   "postgresql.dev4devs.com",
			version: "v1alpha1",
			kind:    "Database",
			want:    nil,
			wantErr: false,
		},
		{
			name: "found CRD with spec",
			swagger: []byte(`{
  "definitions": {
	"com.dev4devs.postgresql.v1alpha1.Database": {
		"type": "object",
		"x-kubernetes-group-version-kind": [
			{
				"group": "postgresql.dev4devs.com",
				"kind": "Database",
				"version": "v1alpha1"
			}
		],
		"properties": {
			"spec": {
				"type": "object"
			}
		}
	}
  }
}`),
			group:   "postgresql.dev4devs.com",
			version: "v1alpha1",
			kind:    "Database",
			want: &spec.Schema{
				SchemaProps: spec.SchemaProps{
					Type: []string{"object"},
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, gotErr := getResourceSpecDefinitionFromSwagger(tt.swagger, tt.group, tt.version, tt.kind)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("expected %+v\n\ngot %+v", tt.want, got)
			}
			if (gotErr != nil) != tt.wantErr {
				t.Errorf("Expected error %v, got %v", tt.wantErr, gotErr)
			}
		})
	}
}

func TestToOpenAPISpec(t *testing.T) {
	tests := []struct {
		name string
		repr olm.CRDDescription
		want spec.Schema
	}{
		{
			name: "one-level property",
			repr: olm.CRDDescription{
				SpecDescriptors: []olm.SpecDescriptor{
					{
						Path:        "path1",
						DisplayName: "name to display 1",
						Description: "description 1",
					},
				},
			},
			want: spec.Schema{
				SchemaProps: spec.SchemaProps{
					Type: []string{"object"},
					Properties: map[string]spec.Schema{
						"path1": {
							SchemaProps: spec.SchemaProps{
								Type:        []string{"string"},
								Description: "description 1",
								Title:       "name to display 1",
							},
						},
					},
					AdditionalProperties: &spec.SchemaOrBool{
						Allows: false,
					},
				},
			},
		},

		{
			name: "multiple-levels property",
			repr: olm.CRDDescription{
				SpecDescriptors: []olm.SpecDescriptor{
					{
						Path:        "subpath1.path1",
						DisplayName: "name to display 1.1",
						Description: "description 1.1",
					},
					{
						Path:        "subpath1.path2",
						DisplayName: "name to display 1.2",
						Description: "description 1.2",
					},
					{
						Path:        "subpath2.path1",
						DisplayName: "name to display 2.1",
						Description: "description 2.1",
					},
				},
			},
			want: spec.Schema{
				SchemaProps: spec.SchemaProps{
					Type: []string{"object"},
					Properties: map[string]spec.Schema{
						"subpath1": {
							SchemaProps: spec.SchemaProps{
								Type: []string{"object"},
								Properties: map[string]spec.Schema{
									"path1": {
										SchemaProps: spec.SchemaProps{
											Type:        []string{"string"},
											Description: "description 1.1",
											Title:       "name to display 1.1",
										},
									},
									"path2": {
										SchemaProps: spec.SchemaProps{
											Type:        []string{"string"},
											Description: "description 1.2",
											Title:       "name to display 1.2",
										},
									},
								},
							},
						},
						"subpath2": {
							SchemaProps: spec.SchemaProps{
								Type: []string{"object"},
								Properties: map[string]spec.Schema{
									"path1": {
										SchemaProps: spec.SchemaProps{
											Type:        []string{"string"},
											Description: "description 2.1",
											Title:       "name to display 2.1",
										},
									},
								},
							},
						},
					},
					AdditionalProperties: &spec.SchemaOrBool{
						Allows: false,
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := toOpenAPISpec(&tt.repr)
			if !reflect.DeepEqual(*result, tt.want) {
				t.Errorf("Failed %s:\n\ngot: %+v\n\nwant: %+v", t.Name(), result, tt.want)
			}
		})
	}
}

/* TODO refactor with service_test */

const UriFolder = "kubernetes"

func setup(testFolderName string, fs devfileFileSystem.Filesystem) (devfileFileSystem.File, error) {
	err := fs.MkdirAll(testFolderName, os.ModePerm)
	if err != nil {
		return nil, err
	}
	err = fs.MkdirAll(filepath.Join(testFolderName, UriFolder), os.ModePerm)
	if err != nil {
		return nil, err
	}
	testFileName, err := fs.Create(filepath.Join(testFolderName, UriFolder, "example.yaml"))
	if err != nil {
		return nil, err
	}
	return testFileName, nil
}

type inlinedComponent struct {
	name    string
	inlined string
}

type uriComponent struct {
	name string
	uri  string
}

func getDevfileData(t *testing.T, inlined []inlinedComponent, uriComp []uriComponent) data.DevfileData {
	devfileData, err := data.NewDevfileData(string(data.APISchemaVersion200))
	if err != nil {
		t.Error(err)
	}
	for _, component := range inlined {
		err = devfileData.AddComponents([]v1alpha2.Component{{
			Name: component.name,
			ComponentUnion: devfile.ComponentUnion{
				Kubernetes: &devfile.KubernetesComponent{
					K8sLikeComponent: devfile.K8sLikeComponent{
						BaseComponent: devfile.BaseComponent{},
						K8sLikeComponentLocation: devfile.K8sLikeComponentLocation{
							Inlined: component.inlined,
						},
					},
				},
			},
		},
		})
		if err != nil {
			t.Error(err)
		}
	}
	for _, component := range uriComp {
		err = devfileData.AddComponents([]v1alpha2.Component{{
			Name: component.name,
			ComponentUnion: devfile.ComponentUnion{
				Kubernetes: &devfile.KubernetesComponent{
					K8sLikeComponent: devfile.K8sLikeComponent{
						BaseComponent: devfile.BaseComponent{},
						K8sLikeComponentLocation: devfile.K8sLikeComponentLocation{
							Uri: component.uri,
						},
					},
				},
			},
		},
		})
		if err != nil {
			t.Error(err)
		}
	}
	return devfileData
}

/* /TODO */

func TestListDevfileServices(t *testing.T) {
	fs := devfileFileSystem.NewFakeFs()

	testFolderName := "someFolder"
	testFileName, err := setup(testFolderName, fs)
	if err != nil {
		t.Errorf("unexpected error : %v", err)
		return
	}

	uriData := `
apiVersion: redis.redis.opstreelabs.in/v1beta1
kind: Redis
metadata:
  name: redis
spec:
  kubernetesConfig:
    image: quay.io/opstree/redis:v6.2`

	err = fs.WriteFile(testFileName.Name(), []byte(uriData), os.ModePerm)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	tests := []struct {
		name             string
		devfileObj       parser.DevfileObj
		wantKeys         []string
		wantErr          error
		csvSupport       bool
		csvSupportErr    error
		gvrList          []meta.RESTMapping
		gvrListErr       error
		restMapping      *meta.RESTMapping
		restMappingErr   error
		u                unstructured.Unstructured
		inlinedComponent string
	}{
		{
			name: "No service in devfile",
			devfileObj: parser.DevfileObj{
				Data: getDevfileData(t, nil, nil),
				Ctx:  devfileCtx.FakeContext(fs, parser.OutputDevfileYamlPath),
			},
			wantKeys:         []string{},
			wantErr:          nil,
			csvSupport:       true,
			csvSupportErr:    nil,
			gvrList:          []meta.RESTMapping{},
			gvrListErr:       nil,
			restMapping:      &meta.RESTMapping{},
			restMappingErr:   nil,
			u:                unstructured.Unstructured{},
			inlinedComponent: "",
		},
		{
			name: "Services including service bindings in devfile",
			devfileObj: parser.DevfileObj{
				Data: getDevfileData(t, []inlinedComponent{
					{
						name: "link1",
						inlined: `
apiVersion: binding.operators.coreos.com/v1alpha1
kind: ServiceBinding
metadata:
  name: nodejs-prj1-api-vtzg-redis-redis
spec:
  application:
    group: apps
    name: nodejs-prj1-api-vtzg-app
    resource: deployments
    version: v1
  bindAsFiles: false
  detectBindingResources: true
  services:
  - group: redis.redis.opstreelabs.in
    kind: Redis
    name: redis
    version: v1beta1`,
					},
				}, nil),
			},
			wantKeys:       []string{"ServiceBinding/link1"},
			wantErr:        nil,
			csvSupport:     true,
			csvSupportErr:  nil,
			gvrList:        []meta.RESTMapping{},
			gvrListErr:     nil,
			restMapping:    &meta.RESTMapping{},
			restMappingErr: errors.New("some error"), // because SBO is not installed
			u:              unstructured.Unstructured{},
			inlinedComponent: `
apiVersion: binding.operators.coreos.com/v1alpha1
kind: ServiceBinding
metadata:
  name: nodejs-prj1-api-vtzg-redis-redis
spec:
  application:
    group: apps
    name: nodejs-prj1-api-vtzg-app
    resource: deployments
    version: v1
  bindAsFiles: false
  detectBindingResources: true
  services:
  - group: redis.redis.opstreelabs.in
    kind: Redis
    name: redis
    version: v1beta1`,
		},
		{
			name: "URI reference in devfile",
			devfileObj: parser.DevfileObj{
				Data: getDevfileData(t, nil, []uriComponent{
					{
						name: "service1",
						uri:  filepath.Join(UriFolder, filepath.Base(testFileName.Name())),
					},
				}),
			},
			wantKeys:         []string{"Redis/service1"},
			wantErr:          nil,
			csvSupport:       false,
			csvSupportErr:    nil,
			gvrList:          nil,
			gvrListErr:       nil,
			restMapping:      nil,
			restMappingErr:   errors.New("some error"), // because Redis Operator is not installed
			u:                unstructured.Unstructured{},
			inlinedComponent: uriData,
		},
	}

	getKeys := func(m map[string]unstructured.Unstructured) []string {
		keys := make([]string, len(m))
		i := 0
		for key := range m {
			keys[i] = key
			i += 1
		}
		sort.Strings(keys)
		return keys
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockCtrl := gomock.NewController(t)
			defer mockCtrl.Finish()

			fkClient := NewMockClientInterface(mockCtrl)
			fkClient.EXPECT().IsCSVSupported().Return(tt.csvSupport, tt.csvSupportErr).AnyTimes()
			fkClient.EXPECT().GetOperatorGVRList().Return(tt.gvrList, tt.gvrListErr).AnyTimes()
			_ = yaml.Unmarshal([]byte(tt.inlinedComponent), &tt.u)

			fkClient.EXPECT().GetRestMappingFromUnstructured(tt.u).Return(tt.restMapping, tt.restMappingErr).AnyTimes()
			//fkClient.EXPECT().GetRestMappingFromUnstructured(tt.u).Return(tt.restMapping, tt.restMappingErr).Times(2)

			got, gotErr := listDevfileServices(fkClient, tt.devfileObj, testFolderName, fs)
			gotKeys := getKeys(got)
			if !reflect.DeepEqual(gotKeys, tt.wantKeys) {
				t.Errorf("%s: got %v, expect %v", t.Name(), gotKeys, tt.wantKeys)
			}
			if gotErr != tt.wantErr {
				t.Errorf("%s: got %v, expect %v", t.Name(), gotErr, tt.wantErr)
			}
		})
	}
}
