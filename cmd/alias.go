package cmd

import (
	"fmt"
	"github.com/rverst/go-miab/miab"
	"github.com/spf13/cobra"
	"os"
)

func init() {
	rootCmd.AddCommand(aliasCmd)
	aliasCmd.AddCommand(aliasAddCmd, aliasRemoveCmd)

	aliasCmd.Flags().String("domain", "", "domain to filter the list of aliases")
	aliasCmd.Flags().String("format", "plain", "the output format (plain, csv, json, yaml)")
	aliasAddCmd.PersistentFlags().String("address", "", "alias address [mandatory]")
	aliasAddCmd.Flags().String("forward", "", "email address(es) to forward to (comma separated) [mandatory]")
	aliasRemoveCmd.PersistentFlags().String("address", "", "alias address [mandatory]")

	_ = aliasAddCmd.MarkPersistentFlagRequired("address")
	_ = aliasAddCmd.MarkFlagRequired("forwards")

	_ = aliasRemoveCmd.MarkPersistentFlagRequired("address")
}

var aliasCmd = &cobra.Command{
	Use:   "alias",
	Short: "Get existing mail aliases for users.",
	Long:  `Get all mail aliases for the users of the server, use the domain-flag to filter the output.`,
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {

		format := miab.PLAIN
		if f, err := cmd.Flags().GetString("format"); err == nil {
			format = miab.Formats(f)
		}

		aliasDomains, err := miab.GetAliases(&config)
		if err != nil {
			fmt.Printf("Error fetching email aliasDomains: %v\n", err)
			os.Exit(1)
		}

		d, err := cmd.Flags().GetString("domain")
		if err == nil && d != "" {
			for _, aliasDomain := range aliasDomains {
				if aliasDomain.Domain == d {
					aliasDomain.Print(format)
					return
				}
			}
		}

		aliasDomains.Print(format)
	},
}

var aliasAddCmd = &cobra.Command{
	Use:   "add",
	Short: "Add an email alias",
	Long:  `Add an email alias`,
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {

		email, _ := cmd.Flags().GetString("address")
		fwd, _ := cmd.Flags().GetString("forward")

		if err := miab.AddAlias(&config, email, fwd); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	},
}

var aliasRemoveCmd = &cobra.Command{
	Use:   "remove",
	Short: "Remove an email alias",
	Long:  `Remove an email alias`,
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {

		email, _ := cmd.Flags().GetString("address")

		if err := miab.RemoveAlias(&config, email); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	},
}