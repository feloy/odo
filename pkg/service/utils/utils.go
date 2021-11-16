package utils

import (
	"fmt"
	"strings"

	"github.com/openshift/odo/pkg/devfile"

	workspaces "github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
	devfilefs "github.com/devfile/library/pkg/testingutil/filesystem"

	"github.com/ghodss/yaml"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

// SplitServiceKindName splits the service name provided for deletion by the
// user. It has to be of the format <service-kind>/<service-name>. Example: EtcdCluster/myetcd
func SplitServiceKindName(serviceName string) (string, string, error) {
	sn := strings.SplitN(serviceName, "/", 2)
	if len(sn) != 2 || sn[0] == "" || sn[1] == "" {
		return "", "", fmt.Errorf("couldn't split %q into exactly two", serviceName)
	}

	kind := sn[0]
	name := sn[1]

	return kind, name, nil
}

func IsLinkResource(kind string) bool {
	return kind == "ServiceBinding"
}

func GetK8sComponentAsUnstructured(component *workspaces.KubernetesComponent, context string, fs devfilefs.Filesystem) (unstructured.Unstructured, error) {
	strCRD := component.Inlined
	var err error
	if component.Uri != "" {
		strCRD, err = devfile.GetDataFromURI(component.Uri, context, fs)
		if err != nil {
			return unstructured.Unstructured{}, err
		}
	}

	// convert the YAML definition into map[string]interface{} since it's needed to create dynamic resource
	u := unstructured.Unstructured{}
	if err = yaml.Unmarshal([]byte(strCRD), &u.Object); err != nil {
		return unstructured.Unstructured{}, err
	}
	return u, nil
}
