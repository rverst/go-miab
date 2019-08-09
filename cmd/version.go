package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(versionCmd)
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of go-miab",
	Long:  `All software has versions. This is go-miab's`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Mail-in-a-Box API cli v0.1")
	},
}