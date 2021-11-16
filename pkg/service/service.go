package service

import (
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/openshift/odo/pkg/kclient"
	"github.com/openshift/odo/pkg/service/utils"

	workspaces "github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
	"github.com/devfile/library/pkg/devfile/parser"
	"github.com/devfile/library/pkg/devfile/parser/data/v2/common"
	parsercommon "github.com/devfile/library/pkg/devfile/parser/data/v2/common"
	devfilefs "github.com/devfile/library/pkg/testingutil/filesystem"

	olm "github.com/operator-framework/api/pkg/operators/v1alpha1"
	"github.com/pkg/errors"

	servicebinding "github.com/redhat-developer/service-binding-operator/apis/binding/v1alpha1"
)

// LinkLabel is the name of the name of the link in the devfile
const LinkLabel = "app.kubernetes.io/link-name"

// ServiceLabel is the name of the service in the service binding object
const ServiceLabel = "app.kubernetes.io/service-name"

// ServiceKind is the kind of the service in the service binding object
const ServiceKind = "app.kubernetes.io/service-kind"

const UriFolder = "kubernetes"

const filePrefix = "odo-service-"

func GetGVRFromOperator(csv olm.ClusterServiceVersion, cr string) (string, string, string, error) {
	var group, version, resource string

	for _, customresource := range csv.Spec.CustomResourceDefinitions.Owned {
		custRes := customresource
		if custRes.Kind == cr {
			group, version, resource = kclient.GetGVRFromCR(&custRes)
			return group, version, resource, nil
		}
	}
	return "", "", "", fmt.Errorf("couldn't parse group, version, resource from Operator %q", csv.Name)
}

func GetGVKFromCR(cr *olm.CRDDescription) (group, version, kind string, err error) {
	return getGVKFromCR(cr)
}

// getGVKFromCR parses and returns the values for group, version and resource
// for a given Custom Resource (CR).
func getGVKFromCR(cr *olm.CRDDescription) (group, version, kind string, err error) {
	kind = cr.Kind
	version = cr.Version

	gr := strings.SplitN(cr.Name, ".", 2)
	if len(gr) != 2 {
		err = fmt.Errorf("couldn't split Custom Resource's name into two: %s", cr.Name)
		return
	}
	group = gr[1]

	return
}

// GetAlmExample fetches the ALM example from an Operator's definition. This
// example contains the example yaml to be used to spin up a service for a
// given CR in an Operator
func GetAlmExample(csv olm.ClusterServiceVersion, cr, serviceType string) (almExample map[string]interface{}, err error) {
	var almExamples []map[string]interface{}

	val, ok := csv.Annotations["alm-examples"]
	if ok {
		err = json.Unmarshal([]byte(val), &almExamples)
		if err != nil {
			return nil, errors.Wrap(err, "unable to unmarshal alm-examples")
		}
	} else {
		// There's no alm examples in the CSV's definition
		return nil,
			fmt.Errorf("could not find alm-examples in %q Operator's definition", cr)
	}

	almExample, err = getAlmExample(almExamples, cr, serviceType)
	if err != nil {
		return nil, err
	}

	return almExample, nil
}

// getAlmExample returns the alm-example for exact service of an Operator
func getAlmExample(almExamples []map[string]interface{}, crd, operator string) (map[string]interface{}, error) {
	for _, example := range almExamples {
		if example["kind"].(string) == crd {
			// Remove metadata.namespace from example
			if metadata, ok := example["metadata"].(map[string]interface{}); ok {
				delete(metadata, "namespace")
			}
			return example, nil
		}
	}
	return nil, errors.Errorf("could not find example yaml definition for %q service in %q Operator's definition.", crd, operator)
}

// IsOperatorServiceNameValid checks if the provided name follows
// <service-type>/<service-name> format. For example: "EtcdCluster/example" is
// a valid service name but "EtcdCluster/", "EtcdCluster", "example" aren't.
func IsOperatorServiceNameValid(name string) (string, string, error) {
	checkName := strings.SplitN(name, "/", 2)

	if len(checkName) != 2 || checkName[0] == "" || checkName[1] == "" {
		return "", "", fmt.Errorf("invalid service name. Must adhere to <service-type>/<service-name> formatting. For example: %q. Execute %q for list of services", "EtcdCluster/example", "odo service list")
	}
	return checkName[0], checkName[1], nil
}

// IsDefined checks if a service with the given name is defined in a DevFile
func IsDefined(name string, devfileObj parser.DevfileObj) (bool, error) {
	components, err := devfileObj.Data.GetComponents(common.DevfileOptions{})
	if err != nil {
		return false, err
	}
	for _, c := range components {
		if c.Name == name {
			return true, nil
		}
	}
	return false, nil
}

// ListDevfileLinks returns the names of the links defined in a Devfile
func ListDevfileLinks(devfileObj parser.DevfileObj, context string) ([]string, error) {
	return listDevfileLinks(devfileObj, context, devfilefs.DefaultFs{})
}

func listDevfileLinks(devfileObj parser.DevfileObj, context string, fs devfilefs.Filesystem) ([]string, error) {
	if devfileObj.Data == nil {
		return nil, nil
	}
	components, err := devfileObj.Data.GetComponents(common.DevfileOptions{
		ComponentOptions: parsercommon.ComponentOptions{ComponentType: workspaces.KubernetesComponentType},
	})
	if err != nil {
		return nil, err
	}
	var services []string
	for _, c := range components {
		u, err := utils.GetK8sComponentAsUnstructured(c.Kubernetes, context, fs)
		if err != nil {
			return nil, err
		}
		if !utils.IsLinkResource(u.GetKind()) {
			continue
		}
		var sbr servicebinding.ServiceBinding
		js, err := u.MarshalJSON()
		if err != nil {
			return nil, err
		}
		err = json.Unmarshal(js, &sbr)
		if err != nil {
			return nil, err
		}
		sbrServices := sbr.Spec.Services
		if len(sbrServices) != 1 {
			return nil, errors.New("ServiceBinding should have only one service")
		}
		service := sbrServices[0]
		if service.Kind == "Service" {
			services = append(services, service.Name)
		} else {
			services = append(services, service.Kind+"/"+service.Name)
		}
	}
	return services, nil
}

// FindDevfileServiceBinding returns the name of the ServiceBinding defined in a Devfile matching kind and name
func FindDevfileServiceBinding(devfileObj parser.DevfileObj, kind string, name, context string) (string, bool, error) {
	return findDevfileServiceBinding(devfileObj, kind, name, context, devfilefs.DefaultFs{})
}

func findDevfileServiceBinding(devfileObj parser.DevfileObj, kind string, name, context string, fs devfilefs.Filesystem) (string, bool, error) {
	if devfileObj.Data == nil {
		return "", false, nil
	}
	components, err := devfileObj.Data.GetComponents(common.DevfileOptions{
		ComponentOptions: parsercommon.ComponentOptions{ComponentType: workspaces.KubernetesComponentType},
	})
	if err != nil {
		return "", false, err
	}

	for _, c := range components {
		u, err := utils.GetK8sComponentAsUnstructured(c.Kubernetes, context, fs)
		if err != nil {
			return "", false, err
		}
		if utils.IsLinkResource(u.GetKind()) {
			var sbr servicebinding.ServiceBinding
			js, err := u.MarshalJSON()
			if err != nil {
				return "", false, err
			}
			err = json.Unmarshal(js, &sbr)
			if err != nil {
				return "", false, err
			}
			services := sbr.Spec.Services
			if len(services) != 1 {
				continue
			}
			service := services[0]
			if service.Kind == kind && service.Name == name {
				return u.GetName(), true, nil
			}
		}
	}
	return "", false, nil
}

// AddKubernetesComponentToDevfile adds service definition to devfile as an inlined Kubernetes component
func AddKubernetesComponentToDevfile(crd, name string, devfileObj parser.DevfileObj) error {
	err := devfileObj.Data.AddComponents([]workspaces.Component{{
		Name: name,
		ComponentUnion: workspaces.ComponentUnion{
			Kubernetes: &workspaces.KubernetesComponent{
				K8sLikeComponent: workspaces.K8sLikeComponent{
					BaseComponent: workspaces.BaseComponent{},
					K8sLikeComponentLocation: workspaces.K8sLikeComponentLocation{
						Inlined: crd,
					},
				},
			},
		},
	}})
	if err != nil {
		return err
	}

	return devfileObj.WriteYamlDevfile()
}

// AddKubernetesComponent adds the crd information to a separate file and adds the uri information to a devfile component
func AddKubernetesComponent(crd, name, componentContext string, devfile parser.DevfileObj) error {
	return addKubernetesComponent(crd, name, componentContext, devfile, devfilefs.DefaultFs{})
}

// AddKubernetesComponent adds the crd information to a separate file and adds the uri information to a devfile component
func addKubernetesComponent(crd, name, componentContext string, devfileObj parser.DevfileObj, fs devfilefs.Filesystem) error {
	filePath := filepath.Join(componentContext, UriFolder, filePrefix+name+".yaml")
	if _, err := fs.Stat(filepath.Join(componentContext, UriFolder)); os.IsNotExist(err) {
		err = fs.MkdirAll(filepath.Join(componentContext, UriFolder), os.ModePerm)
		if err != nil {
			return err
		}
	}

	if _, err := fs.Stat(filePath); !os.IsNotExist(err) {
		return fmt.Errorf("the file %q already exists", filePath)
	}

	err := fs.WriteFile(filePath, []byte(crd), 0755)
	if err != nil {
		return err
	}

	err = devfileObj.Data.AddComponents([]workspaces.Component{{
		Name: name,
		ComponentUnion: workspaces.ComponentUnion{
			Kubernetes: &workspaces.KubernetesComponent{
				K8sLikeComponent: workspaces.K8sLikeComponent{
					BaseComponent: workspaces.BaseComponent{},
					K8sLikeComponentLocation: workspaces.K8sLikeComponentLocation{
						Uri: filepath.Join(UriFolder, filePrefix+name+".yaml"),
					},
				},
			},
		},
	}})
	if err != nil {
		return err
	}

	return devfileObj.WriteYamlDevfile()
}

// DeleteKubernetesComponentFromDevfile deletes an inlined Kubernetes component from devfile, if one exists
func DeleteKubernetesComponentFromDevfile(name string, devfileObj parser.DevfileObj, componentContext string) error {
	return deleteKubernetesComponentFromDevfile(name, devfileObj, componentContext, devfilefs.DefaultFs{})
}

// deleteKubernetesComponentFromDevfile deletes an inlined Kubernetes component from devfile, if one exists
func deleteKubernetesComponentFromDevfile(name string, devfileObj parser.DevfileObj, componentContext string, fs devfilefs.Filesystem) error {
	components, err := devfileObj.Data.GetComponents(common.DevfileOptions{})
	if err != nil {
		return err
	}

	found := false
	for _, c := range components {
		if c.Name == name {
			err = devfileObj.Data.DeleteComponent(c.Name)
			if err != nil {
				return err
			}

			if c.Kubernetes.Uri != "" {
				parsedURL, err := url.Parse(c.Kubernetes.Uri)
				if err != nil {
					return err
				}
				if len(parsedURL.Host) == 0 || len(parsedURL.Scheme) == 0 {
					err := fs.Remove(filepath.Join(componentContext, c.Kubernetes.Uri))
					if err != nil {
						return err
					}
				}
			}
			found = true
			break
		}
	}

	if !found {
		return fmt.Errorf("could not find the service %q in devfile", name)
	}

	return devfileObj.WriteYamlDevfile()
}
