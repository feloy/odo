package devfile

import (
	"fmt"
	"net/url"
	"path/filepath"

	"strings"

	"github.com/devfile/library/pkg/devfile"
	"github.com/devfile/library/pkg/devfile/parser"
	devfilefs "github.com/devfile/library/pkg/testingutil/filesystem"
	"github.com/openshift/odo/pkg/devfile/validate"
	"github.com/openshift/odo/pkg/log"
	"github.com/openshift/odo/pkg/util"
)

func parseDevfile(args parser.ParserArgs) (parser.DevfileObj, error) {
	devObj, varWarnings, err := devfile.ParseDevfileAndValidate(args)
	if err != nil {
		return parser.DevfileObj{}, err
	}

	// odo specific validations
	err = validate.ValidateDevfileData(devObj.Data)
	if err != nil {
		return parser.DevfileObj{}, err
	}

	// display warnings related to variable substitution
	for variable, messages := range varWarnings.Commands {
		log.Warningf(variableWarning("commands", variable, messages))
	}
	for variable, messages := range varWarnings.Components {
		log.Warningf(variableWarning("components", variable, messages))
	}
	for variable, messages := range varWarnings.Projects {
		log.Warningf(variableWarning("projects", variable, messages))
	}
	for variable, messages := range varWarnings.StarterProjects {
		log.Warningf(variableWarning("starterProjects", variable, messages))
	}

	return devObj, nil
}

// ParseAndValidateFromFile reads, parses and validates  devfile from a file
// if there are warning it logs them on stdout
func ParseAndValidateFromFile(devfilePath string) (parser.DevfileObj, error) {
	return parseDevfile(parser.ParserArgs{Path: devfilePath})
}

// ParseAndValidateFromData parses devfile from []byte and does all the validation
// if there are warning it logs them on stdout
func ParseAndValidateFromData(data []byte) (parser.DevfileObj, error) {
	return parseDevfile(parser.ParserArgs{Data: data})

}

// ParseAndValidateFromURL parses devfile from given url and does all the validation
// if there are warning it logs them on stdout
func ParseAndValidateFromURL(url string) (parser.DevfileObj, error) {
	return parseDevfile(parser.ParserArgs{URL: url})
}

func variableWarning(section string, variable string, messages []string) string {
	quotedVars := []string{}
	for _, v := range messages {
		quotedVars = append(quotedVars, fmt.Sprintf("%q", v))
	}
	return fmt.Sprintf("Invalid variable(s) %s in %q section with name %q. ", strings.Join(quotedVars, ","), section, variable)

}

// GetDataFromURI gets the data from the given URI
// if the uri is a local path, we use the componentContext to complete the local path
func GetDataFromURI(uri, componentContext string, fs devfilefs.Filesystem) (string, error) {

	parsedURL, err := url.Parse(uri)
	if err != nil {
		return "", err
	}
	if len(parsedURL.Host) != 0 && len(parsedURL.Scheme) != 0 {
		params := util.HTTPRequestParams{
			URL: uri,
		}
		dataBytes, err := util.DownloadFileInMemoryWithCache(params, 1)
		if err != nil {
			return "", err
		}
		return string(dataBytes), nil
	} else {
		dataBytes, err := fs.ReadFile(filepath.Join(componentContext, uri))
		if err != nil {
			return "", err
		}
		return string(dataBytes), nil
	}
}
