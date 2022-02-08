package init

import (
	"errors"
	"fmt"
	"net/url"
	"path/filepath"
	"strings"

	"github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
	"github.com/devfile/library/pkg/devfile/parser"

	"github.com/redhat-developer/odo/pkg/catalog"
	"github.com/redhat-developer/odo/pkg/init/asker"
	"github.com/redhat-developer/odo/pkg/init/backend"
	"github.com/redhat-developer/odo/pkg/init/registry"
	"github.com/redhat-developer/odo/pkg/log"
	"github.com/redhat-developer/odo/pkg/preference"
	"github.com/redhat-developer/odo/pkg/segment"
	"github.com/redhat-developer/odo/pkg/testingutil/filesystem"
	"github.com/redhat-developer/odo/pkg/util"
)

type InitClient struct {
	backends         []backend.InitBackend
	fsys             filesystem.Filesystem
	preferenceClient preference.Client
	registryClient   registry.Client
}

func NewInitClient(fsys filesystem.Filesystem, preferenceClient preference.Client, registryClient registry.Client) *InitClient {
	backends := []backend.InitBackend{
		backend.NewFlagsBackend(preferenceClient),
		backend.NewInteractiveBackend(asker.NewSurveyAsker(), catalog.NewCatalogClient(fsys, preferenceClient)),
	}
	return &InitClient{
		backends:         backends,
		fsys:             fsys,
		preferenceClient: preferenceClient,
		registryClient:   registryClient,
	}
}

// Validate calls Validate methods of all backends and returns the first error returned
// or nil if all backends returns a nil error
func (o *InitClient) Validate(flags map[string]string) error {
	for _, backend := range o.backends {
		err := backend.Validate(flags)
		if err != nil {
			return err
		}
	}
	return nil
}

// SelectDevfile calls SelectDevfile methods of backends in order
// and returns the result of the first method accepting to reply, based on flags
func (o *InitClient) SelectDevfile(flags map[string]string) (*backend.DevfileLocation, error) {
	for _, backend := range o.backends {
		ok, location, err := backend.SelectDevfile(flags)
		if ok {
			return location, err
		}
	}
	return nil, errors.New("no backend found to select a devfile. This should not happen")
}

func (o *InitClient) DownloadDevfile(devfileLocation *backend.DevfileLocation, destDir string) (string, error) {
	destDevfile := filepath.Join(destDir, "devfile.yaml")
	if devfileLocation.DevfilePath != "" {
		return destDevfile, o.downloadDirect(devfileLocation.DevfilePath, destDevfile)
	} else {
		return destDevfile, o.downloadFromRegistry(devfileLocation.DevfileRegistry, devfileLocation.Devfile, destDir)
	}
}

// downloadDirect downloads a devfile at the provided URL and saves it in dest
func (o *InitClient) downloadDirect(URL string, dest string) error {
	parsedURL, err := url.Parse(URL)
	if err != nil {
		return err
	}
	if strings.HasPrefix(parsedURL.Scheme, "http") {
		downloadSpinner := log.Spinnerf("Downloading devfile from %q", URL)
		defer downloadSpinner.End(false)
		params := util.HTTPRequestParams{
			URL: URL,
		}
		devfileData, err := o.registryClient.DownloadFileInMemory(params)
		if err != nil {
			return err
		}
		err = o.fsys.WriteFile(dest, devfileData, 0644)
		if err != nil {
			return err
		}
		downloadSpinner.End(true)
	} else {
		downloadSpinner := log.Spinnerf("Copying devfile from %q", URL)
		defer downloadSpinner.End(false)
		content, err := o.fsys.ReadFile(URL)
		if err != nil {
			return err
		}
		info, err := o.fsys.Stat(URL)
		if err != nil {
			return err
		}
		err = o.fsys.WriteFile(dest, content, info.Mode().Perm())
		if err != nil {
			return err
		}
		downloadSpinner.End(true)
	}

	return nil
}

// downloadFromRegistry downloads a devfile from the provided registry and saves it in dest
// If registryName is empty, will try to download the devfile from the list of registries in preferences
func (o *InitClient) downloadFromRegistry(registryName string, devfile string, dest string) error {
	var downloadSpinner *log.Status
	var forceRegistry bool
	if registryName == "" {
		downloadSpinner = log.Spinnerf("Downloading devfile %q", devfile)
		forceRegistry = false
	} else {
		downloadSpinner = log.Spinnerf("Downloading devfile %q from registry %q", devfile, registryName)
		forceRegistry = true
	}
	defer downloadSpinner.End(false)

	registries := o.preferenceClient.RegistryList()
	var reg preference.Registry
	for _, reg = range *registries {
		if forceRegistry && reg.Name == registryName {
			err := o.registryClient.PullStackFromRegistry(reg.URL, devfile, dest, segment.GetRegistryOptions())
			if err != nil {
				return err
			}
			downloadSpinner.End(true)
			return nil
		} else if !forceRegistry {
			err := o.registryClient.PullStackFromRegistry(reg.URL, devfile, dest, segment.GetRegistryOptions())
			if err != nil {
				continue
			}
			downloadSpinner.End(true)
			return nil
		}
	}

	return fmt.Errorf("unable to find the registry with name %q", devfile)
}

// SelectStarterProject calls SelectStarterProject methods of backends in order
// and returns the result of the first method accepting to reply, based on flags
func (o *InitClient) SelectStarterProject(devfile parser.DevfileObj, flags map[string]string) (*v1alpha2.StarterProject, error) {
	for _, backend := range o.backends {
		ok, starter, err := backend.SelectStarterProject(devfile, flags)
		if ok {
			return starter, err
		}
	}
	return nil, errors.New("no backend found to select starter project. This should not happen")
}

func (o *InitClient) DownloadStarterProject(starter *v1alpha2.StarterProject, dest string) error {
	downloadSpinner := log.Spinnerf("Downloading starter project %q", starter.Name)
	err := o.registryClient.DownloadStarterProject(starter, "", dest, false)
	if err != nil {
		downloadSpinner.End(false)
		return err
	}
	downloadSpinner.End(true)
	return nil
}

// PersonalizeName calls PersonalizeName methods of backends in order
// and returns the result of the first method accepting to reply, based on flags
func (o *InitClient) PersonalizeName(devfile parser.DevfileObj, flags map[string]string) error {
	for _, backend := range o.backends {
		ok, err := backend.PersonalizeName(devfile, flags)
		if ok {
			return err
		}
	}
	return errors.New("no backend found to personalize name. This should not happen")
}
