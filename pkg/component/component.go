package component

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
	"github.com/devfile/api/v2/pkg/devfile"
	"github.com/devfile/library/pkg/devfile/parser"
	parsercommon "github.com/devfile/library/pkg/devfile/parser/data/v2/common"
	dfutil "github.com/devfile/library/pkg/util"

	applabels "github.com/redhat-developer/odo/pkg/application/labels"
	componentlabels "github.com/redhat-developer/odo/pkg/component/labels"
	"github.com/redhat-developer/odo/pkg/envinfo"
	"github.com/redhat-developer/odo/pkg/kclient"
	"github.com/redhat-developer/odo/pkg/localConfigProvider"
	"github.com/redhat-developer/odo/pkg/preference"
	"github.com/redhat-developer/odo/pkg/service"
	urlpkg "github.com/redhat-developer/odo/pkg/url"
	"github.com/redhat-developer/odo/pkg/util"

	servicebinding "github.com/redhat-developer/service-binding-operator/apis/binding/v1alpha1"

	v1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog"
)

const componentRandomNamePartsMaxLen = 12
const componentNameMaxRetries = 3
const componentNameMaxLen = -1
const NotAvailable = "Not available"
const apiVersion = "odo.dev/v1alpha1"

// GetComponentDir returns source repo name
// Parameters:
//		path: source path
// Returns: directory name
func GetComponentDir(path string) (string, error) {
	retVal := ""
	if path != "" {
		retVal = filepath.Base(path)
	} else {
		currDir, err := os.Getwd()
		if err != nil {
			return "", fmt.Errorf("unable to generate a random name as getting current directory failed: %w", err)
		}
		retVal = filepath.Base(currDir)
	}
	retVal = strings.TrimSpace(util.GetDNS1123Name(strings.ToLower(retVal)))
	return retVal, nil
}

// GetDefaultComponentName generates a unique component name
// Parameters: desired default component name(w/o prefix) and slice of existing component names
// Returns: Unique component name and error if any
func GetDefaultComponentName(cfg preference.Client, componentPath string, componentType string, existingComponentList ComponentList) (string, error) {
	var prefix string
	var err error

	// Get component names from component list
	var existingComponentNames []string
	for _, component := range existingComponentList.Items {
		existingComponentNames = append(existingComponentNames, component.Name)
	}

	// Create a random generated name for the component to use within Kubernetes
	prefix, err = GetComponentDir(componentPath)
	if err != nil {
		return "", fmt.Errorf("unable to generate random component name: %w", err)
	}
	prefix = util.TruncateString(prefix, componentRandomNamePartsMaxLen)

	// Generate unique name for the component using prefix and unique random suffix
	componentName, err := dfutil.GetRandomName(
		fmt.Sprintf("%s-%s", componentType, prefix),
		componentNameMaxLen,
		existingComponentNames,
		componentNameMaxRetries,
	)
	if err != nil {
		return "", fmt.Errorf("unable to generate random component name: %w", err)
	}

	return util.GetDNS1123Name(componentName), nil
}

// ApplyConfig applies the component config onto component deployment
// Parameters:
//	client: kclient instance
//	componentConfig: Component configuration
//	envSpecificInfo: Component environment specific information, available if uses devfile
// Returns:
//	err: Errors if any else nil
func ApplyConfig(client kclient.ClientInterface, envSpecificInfo envinfo.EnvSpecificInfo) (err error) {
	isRouteSupported := false
	isRouteSupported, err = client.IsRouteSupported()
	if err != nil {
		isRouteSupported = false
	}

	urlClient := urlpkg.NewClient(urlpkg.ClientOptions{
		Client:              client,
		IsRouteSupported:    isRouteSupported,
		LocalConfigProvider: &envSpecificInfo,
	})

	return urlpkg.Push(urlpkg.PushParameters{
		LocalConfigProvider: &envSpecificInfo,
		URLClient:           urlClient,
		IsRouteSupported:    isRouteSupported,
	})
}

// Contains checks to see if the component exists in an array or not
// by checking the name, type and namespace.
// We ignore all other aspects (labels, annotations, etc.) as those are
// propagated by the Kubernetes cluster when retrieving the component and we're only
// interested if there is a duplicate entry.
func Contains(component Component, components []Component) bool {
	for _, comp := range components {
		if (component.Name == comp.Name) && (component.Spec.Type == comp.Spec.Type) && (component.Namespace == comp.Namespace) {
			return true
		}
	}
	return false
}

// RemoveDuplicateComponentsForListingOutput is for listing purposes, we remove all duplicate components which have the same
// kubernetes instance LABEL as each other.
// We are using this for LISTING purposes only (the actual Component struct doesn't have much information).
// If we encounter 2 components with the same name with Dev and Deploy modes, combines them and display "Dev, Deploy"
func RemoveDuplicateComponentsForListingOutput(components []Component) []Component {

	var outputComponents []Component

	// Loop through all the components and merge / combine all annotations + labels so it's the same.
	for _, component := range components {
		componentLabel := component.GetLabels()
		componentAnnotation := component.GetAnnotations()

		for _, check := range components {

			// Make the variable names a bit shorter for readability
			checkLabel := check.GetLabels()
			checkAnnotation := check.GetAnnotations()

			// If we find ones that match the same instance label
			if componentLabel[componentlabels.KubernetesInstanceLabel] == checkLabel[componentlabels.KubernetesInstanceLabel] {

				// Sometimes Kubernetes operators deploy an ingress / service / etc via an operator but do NOT include a
				// "app.kubernetes.io/managed-by" label. So when we "deduplicate" the component list, this may appear blank.
				// If it's blank, and there is an instance of the same name with a managed-by label, we propagate that to the output
				// components
				if componentLabel[componentlabels.KubernetesManagedByLabel] == "" && checkLabel[componentlabels.KubernetesManagedByLabel] != "" {
					component.ObjectMeta.Labels[componentlabels.KubernetesManagedByLabel] = checkLabel[componentlabels.KubernetesManagedByLabel]
				}

				// Check if there are TWO components, one in dev, one in deploy. If so, we will keep the *first* component, and
				// update the mode label to indicate it's running in both dev and deploy.
				// We check this *explicitely* in order to keep the correct order when displaying (ex. Dev, Deploy)
				// instead of some components showing Deploy, Dev depending on the order that the components are in the array.
				// We also use this long if statement in case there is something odd that went on (user modified the label outside of odo)
				componentMode := componentLabel[componentlabels.OdoModeLabel]
				checkMode := checkLabel[componentlabels.OdoModeLabel]
				if (componentMode != checkMode) &&
					(componentMode == componentlabels.ComponentDevName || componentMode == componentlabels.ComponentDeployName) &&
					(checkMode == componentlabels.ComponentDevName || checkMode == componentlabels.ComponentDeployName) {
					component.ObjectMeta.Labels[componentlabels.OdoModeLabel] = fmt.Sprintf("%s, %s", componentlabels.ComponentDevName, componentlabels.ComponentDeployName)
				}

				// Some components have no annotations (ex. knative), so must propagage the ObjectMeta with initial annotations.
				if component.ObjectMeta.Annotations == nil {
					component.ObjectMeta.Annotations = make(map[string]string)
				}

				// Check to see if the next component has a project type annotation, and update the component accordingly if it does.
				if componentAnnotation[componentlabels.OdoProjectTypeAnnotation] == "" && checkAnnotation[componentlabels.OdoProjectTypeAnnotation] != "" {
					component.ObjectMeta.Annotations[componentlabels.OdoProjectTypeAnnotation] = checkAnnotation[componentlabels.OdoProjectTypeAnnotation]
				}

			}
		}

		// Go through the output list, and add the component if it hasn't already been added.
		found := false
		for _, outputComponent := range outputComponents {
			// Go by component name instead of instance-of
			if outputComponent.Name == component.Name {
				found = true
				break
			}
		}
		if !found {
			outputComponents = append(outputComponents, component)
		}

	}
	return outputComponents
}

// ListAllClusterComponents returns a list of all "components" on a cluster
// that are both odo and non-odo components.
//
// We then return a list of "components" intended for listing / output purposes specifically for commands such as:
// `odo list`
func ListAllClusterComponents(client kclient.ClientInterface, namespace string) ([]Component, error) {

	// Get all the dynamic resources available
	resourceList, err := client.GetAllResourcesFromSelector("", namespace)
	if err != nil {
		return []Component{}, fmt.Errorf("unable to list all dynamic resources required to find components: %w", err)
	}

	var components []Component

	for _, resource := range resourceList {

		var labels, annotations map[string]string

		// Retrieve the labels and annotations from the unstructured resource output
		if resource.GetLabels() != nil {
			labels = resource.GetLabels()
		}
		if resource.GetAnnotations() != nil {
			annotations = resource.GetAnnotations()
		}

		// Figure out the correct name to use
		if labels[componentlabels.KubernetesInstanceLabel] == "" {
			continue
		}
		name := labels[componentlabels.KubernetesInstanceLabel]

		// Get the component type (if there is any..)
		var componentType string
		switch {
		case annotations[componentlabels.OdoProjectTypeAnnotation] != "":
			componentType = annotations[componentlabels.OdoProjectTypeAnnotation]
		default:
			componentType = StateTypeUnknown
		}

		// Generate the appropriate "component" with all necessary information
		component := Component{
			TypeMeta: metav1.TypeMeta{
				Kind:       resource.GetKind(),
				APIVersion: resource.GetAPIVersion(),
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:        name,
				Namespace:   namespace,
				Labels:      labels,
				Annotations: annotations,
			},
			Spec: ComponentSpec{
				Type: componentType,
			},
			Status: ComponentStatus{
				State: StateTypeUnknown,
			},
		}

		if !reflect.ValueOf(component).IsZero() {
			components = append(components, component)
		}

	}

	return components, nil
}

// ListDevfileStacks returns the devfile component matching a selector.
// The selector could be about selecting components part of an application.
// There are helpers in "applabels" package for this.
func ListDevfileStacks(client kclient.ClientInterface, selector string) (ComponentList, error) {

	var deploymentList []v1.Deployment
	var components []Component

	// retrieve all the deployments that are associated with this application
	deploymentList, err := client.GetDeploymentFromSelector(selector)
	if err != nil {
		return ComponentList{}, fmt.Errorf("unable to list components: %w", err)
	}

	// create a list of object metadata based on the component and application name (extracted from Deployment labels)
	for _, elem := range deploymentList {
		component, err := GetComponent(client, elem.Labels[componentlabels.KubernetesInstanceLabel], elem.Labels[applabels.ApplicationLabel], client.GetCurrentNamespace())
		if err != nil {
			return ComponentList{}, fmt.Errorf("Unable to get component: %w", err)
		}

		if !reflect.ValueOf(component).IsZero() {
			components = append(components, component)
		}

	}

	compoList := newComponentList(components)
	return compoList, nil
}

// List lists all the devfile components in active application
func List(client kclient.ClientInterface, applicationSelector string) (ComponentList, error) {
	devfileList, err := ListDevfileStacks(client, applicationSelector)
	if err != nil {
		return ComponentList{}, nil
	}
	return newComponentList(devfileList.Items), nil
}

// GetComponentFromDevfile converts the passed in Devfile object to a component for table / listing output purposes
// at the moment we only populate:
// objectmeta.name
// spectype
// To make it compatible with table output / component listing.
func GetComponentFromDevfile(devObj parser.DevfileObj, namespace string) (Component, error) {

	// Populates the component struct with information from devfile.yaml
	component := Component{
		ObjectMeta: metav1.ObjectMeta{
			Name:      devObj.Data.GetMetadata().Name,
			Namespace: namespace,
		},
		Spec: ComponentSpec{
			Type: GetComponentTypeFromDevfileMetadata(devObj.Data.GetMetadata()),
		},
		Status: ComponentStatus{
			LinkedServices: []SecretMount{},
			// We always assume not pushed unless otherwise
			State: StateTypeNotPushed,
		},
	}

	return component, nil
}

// GetComponentFromEnvfile extracts component's metadata from the specified env info if it exists
func GetComponentFromEnvfile(info *envinfo.EnvSpecificInfo) (Component, parser.DevfileObj, error) {
	if info.Exists() {
		devfile, err := parser.Parse(info.GetDevfilePath())
		if err != nil {
			return Component{}, parser.DevfileObj{}, err
		}
		component, err := getComponentFrom(info, GetComponentTypeFromDevfileMetadata(devfile.Data.GetMetadata()))
		if err != nil {
			return Component{}, parser.DevfileObj{}, err
		}
		components, err := devfile.Data.GetComponents(parsercommon.DevfileOptions{})
		if err != nil {
			return Component{}, parser.DevfileObj{}, err
		}
		for _, cmp := range components {
			if cmp.Container != nil {
				for _, env := range cmp.Container.Env {
					component.Spec.Env = append(component.Spec.Env, corev1.EnvVar{Name: env.Name, Value: env.Value})
				}
			}
		}

		return component, devfile, nil
	}
	return Component{}, parser.DevfileObj{}, nil
}

// GetComponentTypeFromDevfileMetadata returns component type from the devfile metadata;
// it could either be projectType or language, if neither of them are set, return 'Not available'
func GetComponentTypeFromDevfileMetadata(metadata devfile.DevfileMetadata) string {
	var componentType string
	if metadata.ProjectType != "" {
		componentType = metadata.ProjectType
	} else if metadata.Language != "" {
		componentType = metadata.Language
	} else {
		componentType = NotAvailable
	}
	return componentType
}

// GetProjectTypeFromDevfileMetadata returns component type from the devfile metadata
func GetProjectTypeFromDevfileMetadata(metadata devfile.DevfileMetadata) string {
	var projectType string
	if metadata.ProjectType != "" {
		projectType = metadata.ProjectType
	} else {
		projectType = NotAvailable
	}
	return projectType
}

// GetLanguageFromDevfileMetadata returns component type from the devfile metadata
func GetLanguageFromDevfileMetadata(metadata devfile.DevfileMetadata) string {
	var language string
	if metadata.Language != "" {
		language = metadata.Language
	} else {
		language = NotAvailable
	}
	return language
}

func getComponentFrom(info localConfigProvider.LocalConfigProvider, componentType string) (Component, error) {
	if info.Exists() {

		component := newComponentWithType(info.GetName(), componentType)

		component.Namespace = info.GetNamespace()

		component.Spec = ComponentSpec{
			App:   info.GetApplication(),
			Type:  componentType,
			Ports: []string{fmt.Sprintf("%d", info.GetDebugPort())},
		}

		urls, err := info.ListURLs()
		if err != nil {
			return Component{}, err
		}
		if len(urls) > 0 {
			for _, url := range urls {
				component.Spec.URL = append(component.Spec.URL, url.Name)
			}
		}

		return component, nil
	}
	return Component{}, nil
}

// Exists checks whether a component with the given name exists in the current application or not
// componentName is the component name to perform check for
// The first returned parameter is a bool indicating if a component with the given name already exists or not
// The second returned parameter is the error that might occurs while execution
func Exists(client kclient.ClientInterface, componentName, applicationName string) (bool, error) {
	deploymentName, err := dfutil.NamespaceOpenShiftObject(componentName, applicationName)
	if err != nil {
		return false, fmt.Errorf("unable to create namespaced name: %w", err)
	}
	deployment, _ := client.GetDeploymentByName(deploymentName)
	if deployment != nil {
		return true, nil
	}
	return false, nil
}

func GetComponentState(client kclient.ClientInterface, componentName, applicationName string) string {
	// Check to see if the deployment has been pushed or not
	c, err := GetPushedComponent(client, componentName, applicationName)
	if err != nil {
		return StateTypeUnknown
	}
	if c != nil {
		return StateTypePushed
	}
	return StateTypeNotPushed
}

// GetComponent provides component definition
func GetComponent(client kclient.ClientInterface, componentName string, applicationName string, projectName string) (component Component, err error) {
	return getRemoteComponentMetadata(client, componentName, applicationName, true, true)
}

// getRemoteComponentMetadata provides component metadata from the **cluster**, not locally.
func getRemoteComponentMetadata(client kclient.ClientInterface, componentName string, applicationName string, getUrls, getStorage bool) (Component, error) {
	fromCluster, err := GetPushedComponent(client, componentName, applicationName)
	if err != nil {
		return Component{}, fmt.Errorf("unable to get remote metadata for %s component: %w", componentName, err)
	}
	if fromCluster == nil {
		return Component{}, nil
	}

	// Component Type
	componentType, err := fromCluster.GetType()
	if err != nil {
		return Component{}, fmt.Errorf("unable to get source type: %w", err)
	}

	component := newComponentWithType(componentName, componentType)

	// URL
	if getUrls {
		urls, e := fromCluster.GetURLs()
		if e != nil {
			return Component{}, e
		}
		component.Spec.URLSpec = urls
		urlsNb := len(urls)
		if urlsNb > 0 {
			res := make([]string, 0, urlsNb)
			for _, url := range urls {
				res = append(res, url.Name)
			}
			component.Spec.URL = res
		}
	}

	// Storage
	if getStorage {
		appStore, e := fromCluster.GetStorage()
		if e != nil {
			return Component{}, fmt.Errorf("unable to get storage list: %w", e)
		}

		component.Spec.StorageSpec = appStore
		var storageList []string
		for _, store := range appStore {
			storageList = append(storageList, store.Name)
		}
		component.Spec.Storage = storageList
	}

	// Environment Variables
	envVars := fromCluster.GetEnvVars()
	var filteredEnv []corev1.EnvVar
	for _, env := range envVars {
		if !strings.Contains(env.Name, "ODO") {
			filteredEnv = append(filteredEnv, env)
		}
	}

	// Secrets
	linkedSecrets := fromCluster.GetLinkedSecrets()
	err = setLinksServiceNames(client, linkedSecrets, componentlabels.GetSelector(componentName, applicationName))
	if err != nil {
		return Component{}, fmt.Errorf("unable to get name of services: %w", err)
	}
	component.Status.LinkedServices = linkedSecrets

	// Annotations
	component.Annotations = fromCluster.GetAnnotations()

	// Mark the component status as pushed
	component.Status.State = StateTypePushed

	// Labels
	component.Labels = fromCluster.GetLabels()
	component.Namespace = client.GetCurrentNamespace()
	component.Spec.App = applicationName
	component.Spec.Env = filteredEnv

	return component, nil
}

// setLinksServiceNames sets the service name of the links from the info in ServiceBindingRequests present in the cluster
func setLinksServiceNames(client kclient.ClientInterface, linkedSecrets []SecretMount, selector string) error {
	ok, err := client.IsServiceBindingSupported()
	if err != nil {
		return fmt.Errorf("unable to check if service binding is supported: %w", err)
	}

	serviceBindings := map[string]string{}
	if ok {
		// service binding operator is installed on the cluster
		list, err := client.ListDynamicResource(kclient.ServiceBindingGroup, kclient.ServiceBindingVersion, kclient.ServiceBindingResource)
		if err != nil || list == nil {
			return err
		}

		for _, u := range list.Items {
			var sbr servicebinding.ServiceBinding
			js, err := u.MarshalJSON()
			if err != nil {
				return err
			}
			err = json.Unmarshal(js, &sbr)
			if err != nil {
				return err
			}
			services := sbr.Spec.Services
			if len(services) != 1 {
				return errors.New("the ServiceBinding resource should define only one service")
			}
			service := services[0]
			if service.Kind == "Service" {
				serviceBindings[sbr.Status.Secret] = service.Name
			} else {
				serviceBindings[sbr.Status.Secret] = service.Kind + "/" + service.Name
			}
		}
	} else {
		// service binding operator is not installed
		// get the secrets instead of the service binding objects to retrieve the link data
		secrets, err := client.ListSecrets(selector)
		if err != nil {
			return err
		}

		// get the services to get their names against the component names
		services, err := client.ListServices("")
		if err != nil {
			return err
		}

		serviceCompMap := make(map[string]string)
		for _, gotService := range services {
			serviceCompMap[gotService.Labels[componentlabels.KubernetesInstanceLabel]] = gotService.Name
		}

		for _, secret := range secrets {
			serviceName, serviceOK := secret.Labels[service.ServiceLabel]
			_, linkOK := secret.Labels[service.LinkLabel]
			serviceKind, serviceKindOK := secret.Labels[service.ServiceKind]
			if serviceKindOK && serviceOK && linkOK {
				if serviceKind == "Service" {
					if _, ok := serviceBindings[secret.Name]; !ok {
						serviceBindings[secret.Name] = serviceCompMap[serviceName]
					}
				} else {
					// service name is stored as kind-name in the labels
					parts := strings.SplitN(serviceName, "-", 2)
					if len(parts) < 2 {
						continue
					}

					serviceName = fmt.Sprintf("%v/%v", parts[0], parts[1])
					if _, ok := serviceBindings[secret.Name]; !ok {
						serviceBindings[secret.Name] = serviceName
					}
				}
			}
		}
	}

	for i, linkedSecret := range linkedSecrets {
		linkedSecrets[i].ServiceName = serviceBindings[linkedSecret.SecretName]
	}
	return nil
}

// GetOnePod gets a pod using the component and app name
func GetOnePod(client kclient.ClientInterface, componentName string, appName string) (*corev1.Pod, error) {
	return client.GetOnePodFromSelector(componentlabels.GetSelector(componentName, appName))
}

// ComponentExists checks whether a deployment by the given name exists in the given app
func ComponentExists(client kclient.ClientInterface, name string, app string) (bool, error) {
	deployment, err := client.GetOneDeployment(name, app)
	if _, ok := err.(*kclient.DeploymentNotFoundError); ok {
		klog.V(2).Infof("Deployment %s not found for belonging to the %s app ", name, app)
		return false, nil
	}
	return deployment != nil, err
}

// Log returns log from component
func Log(client kclient.ClientInterface, componentName string, appName string, follow bool, command v1alpha2.Command) (io.ReadCloser, error) {

	pod, err := GetOnePod(client, componentName, appName)
	if err != nil {
		return nil, fmt.Errorf("the component %s doesn't exist on the cluster", componentName)
	}

	if pod.Status.Phase != corev1.PodRunning {
		return nil, fmt.Errorf("unable to show logs, component is not in running state. current status=%v", pod.Status.Phase)
	}

	containerName := command.Exec.Component

	return client.GetPodLogs(pod.Name, containerName, follow)
}
