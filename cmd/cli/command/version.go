package command

import (
	"fmt"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(versionCmd)

	versionCmd.Flags().BoolP("extended", "x", false, "")
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of miab",
	Long:  `All software should have a (semantic) version, this prints miab's`,
	Args:  cobra.NoArgs,
	Run:   printVersion,
}

func printVersion(cmd *cobra.Command, args []string) {

	if b, _ := cmd.Flags().GetBool("extended"); !b {
		fmt.Println(Version)
		return
	}

	fmt.Printf("Mail-in-a-Box API command-line interface - version: %s\n", Version)
	fmt.Printf("Build date: %s\n", BuildDate)
	fmt.Printf("Commit hash: %s\n", CommitHash)
}
