// envvar package helps converting name/value pairs to devfile EnvVars
package envvar

import (
	"errors"
	"strings"

	devfilev1 "github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
)

// EnvVar represents an environment variable
type EnvVar struct {
	Name  string `yaml:"Name"`
	Value string `yaml:"Value"`
}

// List represents a list of environment variables
type List []EnvVar

// ToDevfileEnv converts the EnvVarList to the list of Envs supported by devfile
func (evl List) ToDevfileEnv() []devfilev1.EnvVar {
	var envList []devfilev1.EnvVar
	for _, ev := range evl {
		envList = append(envList, devfilev1.EnvVar{
			Name:  ev.Name,
			Value: ev.Value,
		})
	}
	return envList
}

// newFromString takes a string of format "name=value" and returns an Env
// variable struct
func newFromString(envStr string) (EnvVar, error) {
	envList := strings.SplitN(envStr, "=", 2)
	// if there is not = in the string
	if len(envList) < 2 {
		return EnvVar{}, errors.New("invalid environment variable format")
	}

	return EnvVar{
		Name:  strings.TrimSpace(envList[0]),
		Value: strings.TrimSpace(envList[1]),
	}, nil
}

// NewListFromSlice takes multiple env variables with format
// "name=value" and returns an EnvVarList
func NewListFromSlice(envList []string) (List, error) {
	var envVarList List
	for _, envStr := range envList {
		envVar, err := newFromString(envStr)
		if err != nil {
			return nil, err
		}
		envVarList = append(envVarList, envVar)
	}

	return envVarList, nil

}

// NewListFromDevfileEnv creates an EnvVarList from the list of envs present in a devfile.
func NewListFromDevfileEnv(envList []devfilev1.EnvVar) List {
	var envVarList List
	for _, env := range envList {
		envVarList = append(envVarList, EnvVar{
			Name:  env.Name,
			Value: env.Value,
		})
	}
	return envVarList
}
