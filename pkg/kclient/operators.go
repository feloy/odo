package kclient

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	applabels "github.com/openshift/odo/pkg/application/labels"
	componentlabels "github.com/openshift/odo/pkg/component/labels"
	"github.com/openshift/odo/pkg/log"
	"github.com/openshift/odo/pkg/service/utils"

	kerrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/discovery/cached/memory"
	"k8s.io/client-go/restmapper"
	"k8s.io/klog"

	devfile "github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
	"github.com/devfile/library/pkg/devfile/parser"
	"github.com/devfile/library/pkg/devfile/parser/data/v2/common"
	devfilefs "github.com/devfile/library/pkg/testingutil/filesystem"

	"github.com/go-openapi/spec"
	olm "github.com/operator-framework/api/pkg/operators/v1alpha1"
	"github.com/pkg/errors"
)

const (
	apiVersion = "odo.dev/v1alpha1"
)

// IsServiceBindingSupported checks if resource of type service binding request present on the cluster
func (c *Client) IsServiceBindingSupported() (bool, error) {
	// Detection of SBO has been removed from issue https://github.com/openshift/odo/issues/5084
	return false, nil
	//	return c.IsResourceSupported("binding.operators.coreos.com", "v1alpha1", "servicebindings")
}

// IsCSVSupported checks if resource of type service binding request present on the cluster
func (c *Client) IsCSVSupported() (bool, error) {
	return c.IsResourceSupported("operators.coreos.com", "v1alpha1", "clusterserviceversions")
}

// ListClusterServiceVersions returns a list of CSVs in the cluster
// It is equivalent to doing `oc get csvs` using oc cli
func (c *Client) ListClusterServiceVersions() (*olm.ClusterServiceVersionList, error) {
	klog.V(3).Infof("Fetching list of operators installed in cluster")
	csvs, err := c.OperatorClient.ClusterServiceVersions(c.Namespace).List(context.TODO(), v1.ListOptions{})
	if err != nil {
		if kerrors.IsNotFound(err) {
			return &olm.ClusterServiceVersionList{}, nil
		}
		return &olm.ClusterServiceVersionList{}, err
	}
	return csvs, nil
}

// GetClusterServiceVersion returns a particular CSV from a list of CSVs
func (c *Client) GetClusterServiceVersion(name string) (olm.ClusterServiceVersion, error) {
	csv, err := c.OperatorClient.ClusterServiceVersions(c.Namespace).Get(context.TODO(), name, v1.GetOptions{})
	if err != nil {
		return olm.ClusterServiceVersion{}, err
	}
	return *csv, nil
}

// GetCustomResourcesFromCSV returns a list of CRs provided by an operator/CSV.
func (c *Client) GetCustomResourcesFromCSV(csv *olm.ClusterServiceVersion) *[]olm.CRDDescription {
	// we will return a list of CRs owned by the csv
	return &csv.Spec.CustomResourceDefinitions.Owned
}

// CheckCustomResourceInCSV checks if the custom resource is present in the CSV.
func (c *Client) CheckCustomResourceInCSV(customResource string, csv *olm.ClusterServiceVersion) (bool, *olm.CRDDescription) {
	var cr *olm.CRDDescription
	hasCR := false
	CRs := c.GetCustomResourcesFromCSV(csv)
	for _, custRes := range *CRs {
		c := custRes
		if c.Kind == customResource {
			cr = &c
			hasCR = true
			break
		}
	}
	return hasCR, cr
}

// SearchClusterServiceVersionList searches for whether the operator/CSV contains
// given keyword then return it
func (c *Client) SearchClusterServiceVersionList(name string) (*olm.ClusterServiceVersionList, error) {
	var result []olm.ClusterServiceVersion
	csvs, err := c.ListClusterServiceVersions()
	if err != nil {
		return &olm.ClusterServiceVersionList{}, errors.Wrap(err, "unable to list services")
	}

	// do a partial search in all the services
	for _, service := range csvs.Items {
		if strings.Contains(service.ObjectMeta.Name, name) {
			result = append(result, service)
		} else {
			for _, crd := range service.Spec.CustomResourceDefinitions.Owned {
				if name == crd.Kind {
					result = append(result, service)
				}
			}
		}
	}

	return &olm.ClusterServiceVersionList{
		TypeMeta: v1.TypeMeta{
			Kind:       "List",
			APIVersion: apiVersion,
		},
		Items: result,
	}, nil
}

// GetCustomResource returns the CR matching the name
func (c *Client) GetCustomResource(customResource string) (*olm.CRDDescription, error) {
	// Get all csvs in the namespace
	csvs, err := c.ListClusterServiceVersions()
	if err != nil {
		return &olm.CRDDescription{}, err
	}

	// iterate of csvs to find if CR of our interest is provided by any of those
	for _, csv := range csvs.Items {
		clusSerVer := csv
		crs := c.GetCustomResourcesFromCSV(&clusSerVer)

		for _, cr := range *crs {
			if cr.Kind == customResource {
				return &cr, nil
			}
		}
	}

	return &olm.CRDDescription{}, fmt.Errorf("could not find a Custom Resource named %q in the namespace", customResource)
}

// GetCSVWithCR returns the CSV (Operator) that contains the CR (service)
func (c *Client) GetCSVWithCR(name string) (*olm.ClusterServiceVersion, error) {
	csvs, err := c.ListClusterServiceVersions()
	if err != nil {
		return &olm.ClusterServiceVersion{}, errors.Wrap(err, "unable to list services")
	}

	for _, csv := range csvs.Items {
		clusterServiceVersion := csv
		for _, cr := range *c.GetCustomResourcesFromCSV(&clusterServiceVersion) {
			if cr.Kind == name {
				return &csv, nil
			}
		}
	}
	return &olm.ClusterServiceVersion{}, fmt.Errorf("could not find any Operator containing requested CR: %s", name)
}

// GetResourceSpecDefinition returns the OpenAPI v2 definition of the Kubernetes resource of a given group/version/kind
func (c *Client) GetResourceSpecDefinition(group, version, kind string) (*spec.Schema, error) {
	data, err := c.KubeClient.Discovery().RESTClient().Get().AbsPath("/openapi/v2").SetHeader("Accept", "application/json").Do(context.TODO()).Raw()
	if err != nil {
		return nil, err
	}
	return getResourceSpecDefinitionFromSwagger(data, group, version, kind)
}

// getResourceSpecDefinitionFromSwagger returns the OpenAPI v2 definition of the Kubernetes resource of a given group/version/kind, for a given swagger data
func getResourceSpecDefinitionFromSwagger(data []byte, group, version, kind string) (*spec.Schema, error) {
	schema := new(spec.Schema)
	err := json.Unmarshal([]byte(data), schema)
	if err != nil {
		return nil, err
	}

	var crd spec.Schema
	found := false
loopDefinitions:
	for _, definition := range schema.Definitions {
		extensions := definition.Extensions
		gvkI, ok := extensions["x-kubernetes-group-version-kind"]
		if !ok {
			continue
		}
		// The concrete type of this extension is expected to be an array of interface{}
		// If not, we ignore it
		gvkA, ok := gvkI.([]interface{})
		if !ok {
			continue
		}

		for i := range gvkA {
			// The concrete type of each element is expected to be a map[string]interface{}
			// If not, we ignore it
			gvk, ok := gvkA[i].(map[string]interface{})
			if !ok {
				continue
			}
			gvkGroup := gvk["group"].(string)
			gvkVersion := gvk["version"].(string)
			gvkKind := gvk["kind"].(string)
			if strings.HasSuffix(group, gvkGroup) && version == gvkVersion && kind == gvkKind {
				crd = definition
				found = true
				break loopDefinitions
			}
		}

	}
	if !found {
		return nil, errors.New("no definition found")
	}

	spec, ok := crd.Properties["spec"]
	if ok {
		return &spec, nil
	}
	return nil, nil
}

// GetCRDSpec returns the specs of a resource in an openAPIv2 format
func (c *Client) GetCRDSpec(cr *olm.CRDDescription, resourceType string, resourceName string) (*spec.Schema, error) {

	crd, err := c.GetResourceSpecDefinition(cr.Name, cr.Version, resourceName)

	if err != nil {
		log.Warning("Unable to get CRD specifications:", err)
	}

	if crd == nil {
		crd = toOpenAPISpec(cr)
	}

	return crd, nil
}

// toOpenAPISpec transforms Spec descriptors from a CRD description to an OpenAPI schema
func toOpenAPISpec(repr *olm.CRDDescription) *spec.Schema {
	if len(repr.SpecDescriptors) == 0 {
		return nil
	}
	schema := new(spec.Schema).Typed("object", "")
	schema.AdditionalProperties = &spec.SchemaOrBool{
		Allows: false,
	}
	for _, param := range repr.SpecDescriptors {
		addParam(schema, param)
	}
	return schema
}

// addParam adds a Spec Descriptor parameter to an OpenAPI schema
func addParam(schema *spec.Schema, param olm.SpecDescriptor) {
	parts := strings.SplitN(param.Path, ".", 2)
	if len(parts) == 1 {
		child := spec.StringProperty()
		if len(param.XDescriptors) == 1 {
			switch param.XDescriptors[0] {
			case "urn:alm:descriptor:com.tectonic.ui:podCount":
				child = spec.Int32Property()
				// TODO(feloy) more cases, based on
				// - https://github.com/openshift/console/blob/master/frontend/packages/operator-lifecycle-manager/src/components/descriptors/reference/reference.md
				// - https://docs.google.com/document/d/17Tdmpu4R6pA5UC4LumyJ2EP6AcotMWM127Jy728hYCk
			}
		}
		child = child.WithTitle(param.DisplayName).WithDescription(param.Description)
		schema.SetProperty(parts[0], *child)
	} else {
		var child *spec.Schema
		if _, ok := schema.Properties[parts[0]]; ok {
			c := schema.Properties[parts[0]]
			child = &c
		} else {
			child = new(spec.Schema).Typed("object", "")
		}
		param.Path = parts[1]
		addParam(child, param)
		schema.SetProperty(parts[0], *child)
	}
}

// GetRestMappingFromUnstructured returns rest mappings from unstructured data
func (client *Client) GetRestMappingFromUnstructured(u unstructured.Unstructured) (*meta.RESTMapping, error) {
	gvk := u.GroupVersionKind()

	cfg := client.GetClientConfig()

	dc, err := discovery.NewDiscoveryClientForConfig(cfg)
	if err != nil {
		return &meta.RESTMapping{}, err
	}
	mapper := restmapper.NewDeferredDiscoveryRESTMapper(memory.NewMemCacheClient(dc))

	return mapper.RESTMapping(gvk.GroupKind(), gvk.Version)
}

// GetOperatorGVRList creates a slice of rest mappings that are provided by Operators (CSV)
func (client *Client) GetOperatorGVRList() ([]meta.RESTMapping, error) {
	var operatorGVRList []meta.RESTMapping

	// ignoring the error because
	csvs, err := client.ListClusterServiceVersions()
	if err != nil {
		return operatorGVRList, err
	}
	for _, c := range csvs.Items {
		owned := c.Spec.CustomResourceDefinitions.Owned
		for i := range owned {
			g, v, r := GetGVRFromCR(&owned[i])
			operatorGVRList = append(operatorGVRList, meta.RESTMapping{
				Resource: schema.GroupVersionResource{
					Group:    g,
					Version:  v,
					Resource: r,
				},
			})
		}
	}
	return operatorGVRList, nil
}

// IsOperatorBackedService checks if the GVR of the CRD belongs to any of the CRs provided by any of the Operators
// if yes, it is an Operator backed service.
// if no, it is likely a Kubernetes built-in resource.
func (client *Client) IsOperatorBackedService(u unstructured.Unstructured) (bool, error) {
	restMapping, err := client.GetRestMappingFromUnstructured(u)
	if err != nil {
		return false, err
	}

	operatorGVRList, err := client.GetOperatorGVRList()
	if err != nil {
		return false, err
	}

	for _, i := range operatorGVRList {
		if i.Resource == restMapping.Resource {
			return true, nil
		}
	}
	return false, nil
}

// createOperatorService creates the given operator on the cluster
// it returns the CR,Kind and errors
func (client *Client) CreateOperatorService(u unstructured.Unstructured) error {
	gvr, err := client.GetRestMappingFromUnstructured(u)
	if err != nil {
		return err
	}

	// create the service on cluster
	err = client.CreateDynamicResource(u, gvr)
	if err != nil {
		return err
	}
	return err
}

// GetCRInstances fetches and returns instances of the CR provided in the
// "customResource" field. It also returns error (if any)
func (client *Client) GetCRInstances(customResource *olm.CRDDescription) (*unstructured.UnstructuredList, error) {
	klog.V(4).Infof("Getting instances of: %s\n", customResource.Name)
	group, version, resource := GetGVRFromCR(customResource)
	return client.ListDynamicResource(group, version, resource)
}

// ListOperatorServices lists all operator backed services.
// It returns list of services, slice of services that it failed (if any) to list and error (if any)
func (client *Client) ListOperatorServices() ([]unstructured.Unstructured, []string, error) {
	klog.V(4).Info("Getting list of services")

	// First let's get the list of all the operators in the namespace
	csvs, err := client.ListClusterServiceVersions()
	if err != nil {
		return nil, nil, err
	}

	if err != nil {
		return nil, nil, errors.Wrap(err, "Unable to list operator backed services")
	}

	var allCRInstances []unstructured.Unstructured
	var failedListingCR []string

	// let's get the Services a.k.a Custom Resources (CR) defined by each operator, one by one
	for _, csv := range csvs.Items {
		clusterServiceVersion := csv
		klog.V(4).Infof("Getting services started from operator: %s", clusterServiceVersion.Name)
		customResources := client.GetCustomResourcesFromCSV(&clusterServiceVersion)

		// list and write active instances of each service/CR
		var instances []unstructured.Unstructured
		for _, cr := range *customResources {
			customResource := cr

			list, err := client.GetCRInstances(&customResource)
			if err != nil {
				crName := strings.Join([]string{csv.Name, cr.Kind}, "/")
				klog.V(4).Infof("Failed to list instances of %q with error: %s", crName, err.Error())
				failedListingCR = append(failedListingCR, crName)
				break
			}

			if len(list.Items) > 0 {
				instances = append(instances, list.Items...)
			}
		}

		// assuming there are more than one instances of a CR
		allCRInstances = append(allCRInstances, instances...)
	}

	return allCRInstances, failedListingCR, nil
}

// DeployedInfo holds information about the services present on the cluster
type DeployedInfo struct {
	Kind           string
	Name           string
	IsLinkResource bool
}

// ListDeployedServices lists the services deployed in the cluster accessible by client
// managed by odo and matching the component label
func (client *Client) ListDeployedServices(labels map[string]string) (map[string]DeployedInfo, error) {
	deployed := map[string]DeployedInfo{}

	deployedServices, _, err := client.ListOperatorServices()
	if err != nil {
		// We ignore ErrNoSuchOperator error as we can deduce Operator Services are not installed
		return nil, err
	}
	for _, svc := range deployedServices {
		name := svc.GetName()
		kind := svc.GetKind()
		deployedLabels := svc.GetLabels()
		if deployedLabels[applabels.ManagedBy] == "odo" && deployedLabels[componentlabels.ComponentLabel] == labels[componentlabels.ComponentLabel] {
			deployed[kind+"/"+name] = DeployedInfo{
				Kind:           kind,
				Name:           name,
				IsLinkResource: utils.IsLinkResource(kind),
			}
		}
	}

	return deployed, nil
}

// PushKubernetesResource pushes a Kubernetes resource (u) to the cluster using client
// adding labels to the resource
func (client *Client) PushKubernetesResource(u unstructured.Unstructured, labels map[string]string) (bool, error) {
	if utils.IsLinkResource(u.GetKind()) {
		// it's a service binding related resource
		return false, nil
	}

	isOp, err := client.IsOperatorBackedService(u)
	if err != nil {
		return false, err
	}

	// add labels to the CRD before creation
	existingLabels := u.GetLabels()
	if isOp {
		u.SetLabels(mergeLabels(existingLabels, labels))
	} else {
		// Kubernetes built-in resource; only set managed-by label to it
		u.SetLabels(mergeLabels(existingLabels, map[string]string{"app.kubernetes.io/managed-by": "odo"}))
	}

	e := client.CreateOperatorService(u)
	if e != nil {
		if strings.Contains(e.Error(), "already exists") {
			// this could be the case when "odo push" was executed after making change to code but there was no change to the service itself
			// TODO: better way to handle this might be introduced by https://github.com/openshift/odo/issues/4553
			return isOp, nil // this ensures that services slice is not updated
		} else {
			return isOp, e
		}
	}
	return isOp, nil
}

func mergeLabels(labels ...map[string]string) map[string]string {
	merged := map[string]string{}
	for _, l := range labels {
		for k, v := range l {
			merged[k] = v
		}
	}
	return merged
}

// DeleteOperatorService deletes an Operator backed service
// TODO: make it unlink the service from component as a part of
// https://github.com/openshift/odo/issues/3563
func (client *Client) DeleteOperatorService(serviceName string) error {
	kind, name, err := utils.SplitServiceKindName(serviceName)
	if err != nil {
		return errors.Wrapf(err, "Refer %q to see list of running services", serviceName)
	}

	csv, err := client.GetCSVWithCR(kind)
	if err != nil {
		return err
	}

	if csv == nil {
		return fmt.Errorf("unable to find any Operator providing the service %q", kind)
	}

	crs := client.GetCustomResourcesFromCSV(csv)
	var cr *olm.CRDDescription
	for _, c := range *crs {
		customResource := c
		if customResource.Kind == kind {
			cr = &customResource
			break
		}
	}

	group, version, resource := GetGVRFromCR(cr)
	return client.DeleteDynamicResource(name, group, version, resource)
}

// PushKubernetesResources updates service(s) from Kubernetes Inlined component in a devfile by creating new ones or removing old ones
func (client *Client) PushKubernetesResources(k8sComponents []devfile.Component, labels map[string]string, context string) error {
	// check csv support before proceeding
	csvSupported, err := client.IsCSVSupported()
	if err != nil {
		return err
	}

	var deployed map[string]DeployedInfo

	if csvSupported {
		deployed, err = client.ListDeployedServices(labels)
		if err != nil {
			return err
		}

		for key, deployedResource := range deployed {
			if deployedResource.IsLinkResource {
				delete(deployed, key)
			}
		}
	}

	madeChange := false

	// create an object on the kubernetes cluster for all the Kubernetes Inlined components
	for _, c := range k8sComponents {
		u, er := utils.GetK8sComponentAsUnstructured(c.Kubernetes, context, devfilefs.DefaultFs{})
		if er != nil {
			return er
		}

		isOperatorBackedService, er := client.PushKubernetesResource(u, labels)
		if er != nil {
			return er
		}
		if csvSupported {
			delete(deployed, u.GetKind()+"/"+u.GetName())
		}
		if isOperatorBackedService {
			log.Successf("Created service %q on the cluster; refer %q to know how to link it to the component", strings.Join([]string{u.GetKind(), u.GetName()}, "/"), "odo link -h")
		}
		madeChange = true
	}

	if csvSupported {
		for key, val := range deployed {
			if utils.IsLinkResource(val.Kind) {
				continue
			}
			err = client.DeleteOperatorService(key)
			if err != nil {
				return err

			}

			log.Successf("Deleted service %q from the cluster", key)
			madeChange = true
		}
	}

	if !madeChange {
		log.Success("Services are in sync with the cluster, no changes are required")
	}

	return nil
}

// OperatorSvcExists checks whether an Operator backed service with given name
// exists or not. It takes 'serviceName' of the format
// '<service-kind>/<service-name>'. For example: EtcdCluster/example.
// It doesn't bother about application since
// https://github.com/openshift/odo/issues/2801 is blocked
func (client *Client) OperatorSvcExists(serviceName string) (bool, error) {
	kind, name, err := utils.SplitServiceKindName(serviceName)
	if err != nil {
		return false, errors.Wrapf(err, "Refer %q to see list of running services", serviceName)
	}

	// Get the CSV (Operator) that provides the CR
	csv, err := client.GetCSVWithCR(kind)
	if err != nil {
		return false, err
	}

	// Get the specific CR that matches "kind"
	crs := client.GetCustomResourcesFromCSV(csv)

	var cr *olm.CRDDescription
	for _, custRes := range *crs {
		c := custRes
		if c.Kind == kind {
			cr = &c
			break
		}
	}

	// Get instances of the specific CR
	crInstances, err := client.GetCRInstances(cr)
	if err != nil {
		return false, err
	}

	for _, s := range crInstances.Items {
		if s.GetKind() == kind && s.GetName() == name {
			return true, nil
		}
	}

	return false, nil
}

// UpdateServicesWithOwnerReferences adds an owner reference to an inlined Kubernetes resource (except service binding objects)
// if not already present in the list of owner references
func (client *Client) UpdateServicesWithOwnerReferences(k8sComponents []devfile.Component, ownerReference metav1.OwnerReference, context string) error {
	for _, c := range k8sComponents {
		u, err := utils.GetK8sComponentAsUnstructured(c.Kubernetes, context, devfilefs.DefaultFs{})
		if err != nil {
			return err
		}

		if utils.IsLinkResource(u.GetKind()) {
			// ignore service binding resources
			continue
		}

		restMapping, err := client.GetRestMappingFromUnstructured(u)
		if err != nil {
			return err
		}

		d, err := client.GetDynamicResource(restMapping.Resource.Group, restMapping.Resource.Version, restMapping.Resource.Resource, u.GetName())
		if err != nil {
			return err
		}

		found := false
		for _, ownerRef := range d.GetOwnerReferences() {
			if ownerRef.UID == ownerReference.UID {
				found = true
				break
			}
		}
		if found {
			continue
		}
		d.SetOwnerReferences(append(d.GetOwnerReferences(), ownerReference))

		err = client.UpdateDynamicResource(restMapping.Resource.Group, restMapping.Resource.Version, restMapping.Resource.Resource, u.GetName(), d)
		if err != nil {
			return err
		}
	}
	return nil
}

// ValidateResourceExist validates if a resource definition for the Kubernetes inlined component is installed on the cluster
func (client *Client) ValidateResourceExist(k8sComponent devfile.Component, context string) (kindErr string, err error) {
	u, err := utils.GetK8sComponentAsUnstructured(k8sComponent.Kubernetes, context, devfilefs.DefaultFs{})
	if err != nil {
		return "", err
	}
	_, err = client.GetRestMappingFromUnstructured(u)
	if err != nil && !utils.IsLinkResource(u.GetKind()) {
		// getting a RestMapping would fail if there are no matches for the Kind field on the cluster;
		// but if it's a "ServiceBinding" resource, we don't add it to unsupported list because odo can create links
		// without having SBO installed
		return u.GetKind(), errors.New("resource not supported")
	}
	return "", nil
}

// ValidateResourcesExist validates if the Kubernetes inlined components are installed on the cluster
func (client *Client) ValidateResourcesExist(k8sComponents []devfile.Component, context string) error {
	if len(k8sComponents) == 0 {
		return nil
	}
	var unsupportedResources []string
	for _, c := range k8sComponents {
		kindErr, err := client.ValidateResourceExist(c, context)
		if err != nil {
			if kindErr != "" {
				unsupportedResources = append(unsupportedResources, kindErr)
			} else {
				return err
			}
		}
	}

	if len(unsupportedResources) > 0 {
		// tell the user about all the unsupported resources in one message
		return fmt.Errorf("following resource(s) in the devfile are not supported by your cluster; please install corresponding Operator(s) before doing \"odo push\": %s", strings.Join(unsupportedResources, ", "))
	}
	return nil
}

// ListDevfileServices returns the names of the services defined in a Devfile
func (client *Client) ListDevfileServices(devfileObj parser.DevfileObj, componentContext string) (map[string]unstructured.Unstructured, error) {
	return listDevfileServices(client, devfileObj, componentContext, devfilefs.DefaultFs{})
}

func listDevfileServices(client ClientInterface, devfileObj parser.DevfileObj, componentContext string, fs devfilefs.Filesystem) (map[string]unstructured.Unstructured, error) {
	if devfileObj.Data == nil {
		return nil, nil
	}
	components, err := devfileObj.Data.GetComponents(common.DevfileOptions{
		ComponentOptions: common.ComponentOptions{ComponentType: devfile.KubernetesComponentType},
	})
	if err != nil {
		return nil, err
	}

	csvSupported, err := client.IsCSVSupported()
	if err != nil {
		return nil, err
	}
	var operatorGVRList []meta.RESTMapping
	if csvSupported {
		operatorGVRList, err = client.GetOperatorGVRList()
		if err != nil {
			return nil, err
		}
	}

	services := map[string]unstructured.Unstructured{}
	for _, c := range components {
		u, err := utils.GetK8sComponentAsUnstructured(c.Kubernetes, componentContext, fs)
		if err != nil {
			return nil, err
		}
		restMapping, err := client.GetRestMappingFromUnstructured(u)
		if err != nil {
			// getting a RestMapping would fail if there are no matches for the Kind field on the cluster
			// this could be a case when an Operator backed service was added to devfile while working on a cluster
			// that had the Operator installed but "odo service list" is run when that Operator is either no longer
			// available or on a different cluster
			services[strings.Join([]string{u.GetKind(), c.Name}, "/")] = u
			continue
		}
		var match bool
		for _, i := range operatorGVRList {
			if i.Resource == restMapping.Resource {
				// if it's an Operator backed service, it will match; if it's Pod, Deployment, etc. it won't
				match = true
				break
			}
		}
		if match {
			services[strings.Join([]string{u.GetKind(), c.Name}, "/")] = u
		}
	}
	// final list of services includes Operator backed services both supported and unsupported by the underlying k8s cluster
	// but it doesn't include things like Pod, Deployment, etc.
	return services, nil
}
