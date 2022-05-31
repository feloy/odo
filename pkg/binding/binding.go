package binding

import (
	"fmt"
	"path/filepath"

	bindingApi "github.com/redhat-developer/service-binding-operator/apis/binding/v1alpha1"
	specApi "github.com/redhat-developer/service-binding-operator/apis/spec/v1alpha3"

	"gopkg.in/yaml.v2"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	devfilev1alpha2 "github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
	"github.com/devfile/library/pkg/devfile/parser"
	parsercommon "github.com/devfile/library/pkg/devfile/parser/data/v2/common"
	devfilefs "github.com/devfile/library/pkg/testingutil/filesystem"

	"github.com/redhat-developer/odo/pkg/api"
	"github.com/redhat-developer/odo/pkg/binding/asker"
	backendpkg "github.com/redhat-developer/odo/pkg/binding/backend"
	"github.com/redhat-developer/odo/pkg/kclient"
	"github.com/redhat-developer/odo/pkg/libdevfile"
)

type BindingClient struct {
	// Backends
	flagsBackend       *backendpkg.FlagsBackend
	interactiveBackend *backendpkg.InteractiveBackend

	// Clients
	kubernetesClient kclient.ClientInterface
}

func NewBindingClient(kubernetesClient kclient.ClientInterface) *BindingClient {
	// We create the asker client and the backends here and not at the CLI level, as we want to hide these details to the CLI
	askerClient := asker.NewSurveyAsker()
	return &BindingClient{
		flagsBackend:       backendpkg.NewFlagsBackend(),
		interactiveBackend: backendpkg.NewInteractiveBackend(askerClient),
		kubernetesClient:   kubernetesClient,
	}
}

// GetFlags gets the flag specific to add binding operation so that it can correctly decide on the backend to be used
// It ignores all the flags except the ones specific to add binding operation, for e.g. verbosity flag
func (o *BindingClient) GetFlags(flags map[string]string) map[string]string {
	bindingFlags := map[string]string{}
	for flag, value := range flags {
		if flag == backendpkg.FLAG_NAME || flag == backendpkg.FLAG_SERVICE || flag == backendpkg.FLAG_BIND_AS_FILES {
			bindingFlags[flag] = value
		}
	}
	return bindingFlags
}

// Validate calls Validate method of the adequate backend
func (o *BindingClient) Validate(flags map[string]string) error {
	var backend backendpkg.AddBindingBackend
	if len(flags) == 0 {
		backend = o.interactiveBackend
	} else {
		backend = o.flagsBackend
	}
	return backend.Validate(flags)
}

func (o *BindingClient) SelectServiceInstance(flags map[string]string, serviceMap map[string]unstructured.Unstructured) (string, error) {
	var backend backendpkg.AddBindingBackend
	if len(flags) == 0 {
		backend = o.interactiveBackend
	} else {
		backend = o.flagsBackend
	}
	return backend.SelectServiceInstance(flags[backendpkg.FLAG_SERVICE], serviceMap)
}

func (o *BindingClient) AskBindingName(serviceName, componentName string, flags map[string]string) (string, error) {
	var backend backendpkg.AddBindingBackend
	if len(flags) == 0 {
		backend = o.interactiveBackend
	} else {
		backend = o.flagsBackend
	}
	defaultBindingName := fmt.Sprintf("%v-%v", componentName, serviceName)
	return backend.AskBindingName(defaultBindingName, flags)
}

func (o *BindingClient) AskBindAsFiles(flags map[string]string) (bool, error) {
	var backend backendpkg.AddBindingBackend
	if len(flags) == 0 {
		backend = o.interactiveBackend
	} else {
		backend = o.flagsBackend
	}
	return backend.AskBindAsFiles(flags)
}

func (o *BindingClient) AddBinding(bindingName string, bindAsFiles bool, unstructuredService unstructured.Unstructured, obj parser.DevfileObj, componentContext string) (parser.DevfileObj, error) {
	service, err := o.kubernetesClient.NewServiceBindingServiceObject(unstructuredService, bindingName)
	if err != nil {
		return obj, err
	}

	deploymentName := fmt.Sprintf("%s-app", obj.GetMetadataName())
	deploymentGVR, err := o.kubernetesClient.GetDeploymentAPIVersion()
	if err != nil {
		return obj, err
	}

	serviceBinding := kclient.NewServiceBindingObject(bindingName, bindAsFiles, deploymentName, deploymentGVR, []bindingApi.Mapping{}, []bindingApi.Service{service})

	// Note: we cannot directly marshal the serviceBinding object to yaml because it doesn't do that in the correct k8s manifest format
	serviceBindingUnstructured, err := kclient.ConvertK8sResourceToUnstructured(serviceBinding)
	if err != nil {
		return obj, err
	}
	yamlDesc, err := yaml.Marshal(serviceBindingUnstructured.UnstructuredContent())
	if err != nil {
		return obj, err
	}

	return libdevfile.AddKubernetesComponentToDevfile(string(yamlDesc), serviceBinding.Name, obj)
}

func (o *BindingClient) GetServiceInstances() (map[string]unstructured.Unstructured, error) {
	// Get the BindableKinds/bindable-kinds object
	bindableKind, err := o.kubernetesClient.GetBindableKinds()
	if err != nil {
		return nil, err
	}

	// get a list of restMappings of all the GVKs present in bindableKind's Status
	bindableKindRestMappings, err := o.kubernetesClient.GetBindableKindStatusRestMapping(bindableKind.Status)
	if err != nil {
		return nil, err
	}

	var bindableObjectMap = map[string]unstructured.Unstructured{}
	for _, restMapping := range bindableKindRestMappings {
		// TODO: Debug into why List returns all the versions instead of the GVR version
		// List all the instances of the restMapping object
		resources, err := o.kubernetesClient.ListDynamicResources(restMapping.Resource)
		if err != nil {
			return nil, err
		}

		for _, item := range resources.Items {
			// format: `<name> (<kind>.<group>)`
			serviceName := fmt.Sprintf("%s (%s.%s)", item.GetName(), item.GetKind(), item.GroupVersionKind().Group)
			bindableObjectMap[serviceName] = item
		}

	}

	return bindableObjectMap, nil
}

// GetBindingsFromDevfile returns all ServiceBinding resources declared as Kubernertes component from a Devfile
// from group binding.operators.coreos.com/v1alpha1 or servicebinding.io/v1alpha3
func (o *BindingClient) GetBindingsFromDevfile(devfileObj parser.DevfileObj, context string) ([]api.ServiceBinding, error) {
	result := []api.ServiceBinding{}
	kubeComponents, err := devfileObj.Data.GetComponents(parsercommon.DevfileOptions{
		ComponentOptions: parsercommon.ComponentOptions{
			ComponentType: devfilev1alpha2.KubernetesComponentType,
		},
	})
	if err != nil {
		return nil, err
	}

	for _, component := range kubeComponents {
		strCRD, err := libdevfile.GetK8sManifestWithVariablesSubstituted(devfileObj, component.Name, context, devfilefs.DefaultFs{})
		if err != nil {
			return nil, err
		}

		u := unstructured.Unstructured{}
		if err := yaml.Unmarshal([]byte(strCRD), &u.Object); err != nil {
			return nil, err
		}

		switch u.GetObjectKind().GroupVersionKind() {
		case bindingApi.GroupVersionKind:

			var sbo bindingApi.ServiceBinding
			err := o.kubernetesClient.ConvertUnstructuredToResource(u, &sbo)
			if err != nil {
				return nil, err
			}

			sb, err := api.ServiceBindingFromBinding(sbo)
			if err != nil {
				return nil, err
			}

			sb.Status, err = o.getStatusFromBinding(sb.Name)
			if err != nil {
				return nil, err
			}

			result = append(result, sb)

		case specApi.GroupVersion.WithKind("ServiceBinding"):

			var sbc specApi.ServiceBinding
			err := o.kubernetesClient.ConvertUnstructuredToResource(u, &sbc)
			if err != nil {
				return nil, err
			}

			sb, err := api.ServiceBindingFromSpec(sbc)
			if err != nil {
				return nil, err
			}

			sb.Status, err = o.getStatusFromSpec(sb.Name)
			if err != nil {
				return nil, err
			}

			result = append(result, sb)

		}
	}
	return result, nil
}

// GetBinding returns the ServiceBinding retource with the given name
// from the cluster, from group binding.operators.coreos.com/v1alpha1 or servicebinding.io/v1alpha3
func (o *BindingClient) GetBinding(name string) (api.ServiceBinding, error) {

	bindingSB, err := o.kubernetesClient.GetBindingServiceBinding(name)
	if err == nil {
		var sb api.ServiceBinding
		sb, err = api.ServiceBindingFromBinding(bindingSB)
		if err != nil {
			return api.ServiceBinding{}, err
		}
		sb.Status, err = o.getStatusFromBinding(bindingSB.Name)
		if err != nil {
			return api.ServiceBinding{}, err
		}
		return sb, nil
	}
	if err != nil && !kerrors.IsNotFound(err) {
		return api.ServiceBinding{}, err
	}

	specSB, err := o.kubernetesClient.GetSpecServiceBinding(name)
	if err == nil {
		var sb api.ServiceBinding
		sb, err = api.ServiceBindingFromSpec(specSB)
		if err != nil {
			return api.ServiceBinding{}, err
		}
		sb.Status, err = o.getStatusFromSpec(specSB.Name)
		if err != nil {
			return api.ServiceBinding{}, err
		}
		return sb, nil
	}

	// In case of notFound error, this time we return the error
	if kerrors.IsNotFound(err) {
		return api.ServiceBinding{}, fmt.Errorf("ServiceBinding %q not found", name)
	}
	return api.ServiceBinding{}, err
}

// getStatusFromBinding returns status information from a ServiceBinding in the cluster
// from group binding.operators.coreos.com/v1alpha1
func (o *BindingClient) getStatusFromBinding(name string) (*api.ServiceBindingStatus, error) {
	bindingSB, err := o.kubernetesClient.GetBindingServiceBinding(name)
	if err != nil {
		if kerrors.IsNotFound(err) {
			return nil, nil
		}
		return nil, err
	}
	secretName := bindingSB.Status.Secret
	secret, err := o.kubernetesClient.GetSecret(secretName, o.kubernetesClient.GetCurrentNamespace())
	if err != nil {
		return nil, err
	}

	if bindingSB.Spec.BindAsFiles {
		bindings := make([]string, 0, len(secret.Data))
		for k := range secret.Data {
			bindingName := filepath.ToSlash(filepath.Join("${SERVICE_BINDING_ROOT}", name, k))
			bindings = append(bindings, bindingName)
		}
		return &api.ServiceBindingStatus{
			BindingFiles: bindings,
		}, nil
	}

	bindings := make([]string, 0, len(secret.Data))
	for k := range secret.Data {
		bindings = append(bindings, k)
	}
	return &api.ServiceBindingStatus{
		BindingEnvVars: bindings,
	}, nil
}

// getStatusFromSpec returns status information from a ServiceBinding in the cluster
// from group servicebinding.io/v1alpha3
func (o *BindingClient) getStatusFromSpec(name string) (*api.ServiceBindingStatus, error) {
	specSB, err := o.kubernetesClient.GetSpecServiceBinding(name)
	if err != nil {
		if kerrors.IsNotFound(err) {
			return nil, nil
		}
		return nil, err
	}
	if specSB.Status.Binding == nil {
		return nil, nil
	}
	secretName := specSB.Status.Binding.Name
	secret, err := o.kubernetesClient.GetSecret(secretName, o.kubernetesClient.GetCurrentNamespace())
	if err != nil {
		return nil, err
	}
	bindingFiles := make([]string, 0, len(secret.Data))
	bindingEnvVars := make([]string, 0, len(specSB.Spec.Env))
	for k := range secret.Data {
		bindingName := filepath.ToSlash(filepath.Join("${SERVICE_BINDING_ROOT}", name, k))
		bindingFiles = append(bindingFiles, bindingName)
	}
	for _, env := range specSB.Spec.Env {
		bindingEnvVars = append(bindingEnvVars, env.Name)
	}
	return &api.ServiceBindingStatus{
		BindingFiles:   bindingFiles,
		BindingEnvVars: bindingEnvVars,
	}, nil
}
