package command

import (
	"fmt"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(versionCmd)
}

const versionTemplate = "Mail-in-a-Box API command-line interface v%s"

var versionCmd = &cobra.Command{

	Use:   "version",
	Short: "Print the version number of go-miab",
	Long:  `All software has versions. This is go-miab's`,
	Run:   printVersion,
}

func printVersion(cmd *cobra.Command, args []string) {
	fmt.Print(version(rootCmd.Version))
}

func version(version string) string {
	return fmt.Sprintf(versionTemplate, version)
}
