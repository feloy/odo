package devfile

import (
	odoutil "github.com/redhat-developer/odo/pkg/odo/util"
	"github.com/spf13/cobra"
)

const RecommendedCommandName = "devfile"

func NewCmdDevfile(name, fullName string) *cobra.Command {
	infoCmd := NewCmdInfo(InfoRecommendedCommandName, odoutil.GetFullName(fullName, InfoRecommendedCommandName))
	pullCmd := NewCmdPull(PullRecommendedCommandName, odoutil.GetFullName(fullName, PullRecommendedCommandName))

	devfileCmd := &cobra.Command{
		Use:  name,
		Long: "devfile related commands",
		Args: cobra.MaximumNArgs(0),
		Run:  func(cmd *cobra.Command, args []string) {},
	}
	devfileCmd.AddCommand(infoCmd, pullCmd)
	return devfileCmd
}
