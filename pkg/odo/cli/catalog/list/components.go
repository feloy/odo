package list

import (
	"fmt"
	"io"
	"os"
	"text/tabwriter"

	"github.com/redhat-developer/odo/pkg/catalog"
	"github.com/redhat-developer/odo/pkg/log"
	"github.com/redhat-developer/odo/pkg/machineoutput"
	"github.com/redhat-developer/odo/pkg/odo/cmdline"
	"github.com/redhat-developer/odo/pkg/odo/genericclioptions"
	"github.com/redhat-developer/odo/pkg/util"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const componentsRecommendedCommandName = "components"

var componentsExample = `  # Get the supported components
  %[1]s`

// ListComponentsOptions encapsulates the options for the odo catalog list components command
type ListComponentsOptions struct {
	// No context needed

	// Clients
	catalogClient catalog.Client

	// list of known devfiles
	catalogDevfileList catalog.DevfileComponentTypeList
}

// NewListComponentsOptions creates a new ListComponentsOptions instance
func NewListComponentsOptions(catalogClient catalog.Client) *ListComponentsOptions {
	return &ListComponentsOptions{
		catalogClient: catalogClient,
	}
}

// Complete completes ListComponentsOptions after they've been created
func (o *ListComponentsOptions) Complete(cmdline cmdline.Cmdline, args []string) (err error) {
	o.catalogDevfileList, err = o.catalogClient.ListDevfileComponents("")
	if err != nil {
		return err
	}

	if o.catalogDevfileList.DevfileRegistries == nil {
		log.Warning("Please run 'odo registry add <registry name> <registry URL>' to add registry for listing devfile components\n")
	}

	return nil
}

// Validate validates the ListComponentsOptions based on completed values
func (o *ListComponentsOptions) Validate() error {
	if len(o.catalogDevfileList.Items) == 0 {
		return fmt.Errorf("no deployable components found")
	}
	return nil
}

type catalogList struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Items             []catalog.DevfileComponentType `json:"items,omitempty"`
}

// Run contains the logic for the command associated with ListComponentsOptions
func (o *ListComponentsOptions) Run() (err error) {
	if log.IsJSON() {
		combinedList := catalogList{
			TypeMeta: metav1.TypeMeta{
				Kind:       "List",
				APIVersion: "odo.dev/v1alpha1",
			},
			Items: o.catalogDevfileList.Items,
		}
		machineoutput.OutputSuccess(combinedList)
	} else {
		w := tabwriter.NewWriter(os.Stdout, 5, 2, 3, ' ', tabwriter.TabIndent)
		if len(o.catalogDevfileList.Items) != 0 {
			fmt.Fprintln(w, "Odo Devfile Components:")
			fmt.Fprintln(w, "NAME", "\t", "DESCRIPTION", "\t", "REGISTRY")

			o.printDevfileCatalogList(w, o.catalogDevfileList.Items, "")
		}
		w.Flush()
	}
	return
}

// NewCmdCatalogListComponents implements the odo catalog list components command
func NewCmdCatalogListComponents(name, fullName string) *cobra.Command {
	o := NewListComponentsOptions(catalog.NewCatalogClient())

	var componentListCmd = &cobra.Command{
		Use:         name,
		Short:       "List all components",
		Long:        "List all available component types from OpenShift's Image Builder",
		Example:     fmt.Sprintf(componentsExample, fullName),
		Annotations: map[string]string{"machineoutput": "json"},
		Run: func(cmd *cobra.Command, args []string) {
			genericclioptions.GenericRun(o, cmd, args)
		},
	}

	return componentListCmd
}

func (o *ListComponentsOptions) printDevfileCatalogList(w io.Writer, catalogDevfileList []catalog.DevfileComponentType, supported string) {
	for _, devfileComponent := range catalogDevfileList {
		if supported != "" {
			fmt.Fprintln(w, devfileComponent.Name, "\t", util.TruncateString(devfileComponent.Description, 60, "..."), "\t", devfileComponent.Registry.Name, "\t", supported)
		} else {
			fmt.Fprintln(w, devfileComponent.Name, "\t", util.TruncateString(devfileComponent.Description, 60, "..."), "\t", devfileComponent.Registry.Name)
		}
	}
}
