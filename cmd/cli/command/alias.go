package command

import (
	"fmt"
	"github.com/rverst/go-miab/miab"
	"github.com/spf13/cobra"
	"os"
)

func init() {
	rootCmd.AddCommand(aliasGetCmd)
	aliasGetCmd.AddCommand(aliasAddCmd, aliasDeleteCmd)

	aliasGetCmd.Flags().String("domain", "", "domain to filter the list of aliases")
	aliasGetCmd.Flags().String("format", "plain", "the output format (plain, csv, json, yaml)")
	aliasAddCmd.PersistentFlags().String("address", "", "alias address [mandatory]")
	aliasAddCmd.Flags().String("forward", "", "email address(es) to forward to (comma separated) [mandatory]")
	aliasDeleteCmd.PersistentFlags().String("address", "", "alias address [mandatory]")

	_ = aliasAddCmd.MarkPersistentFlagRequired("address")
	_ = aliasAddCmd.MarkFlagRequired("forwards")

	_ = aliasDeleteCmd.MarkPersistentFlagRequired("address")
}

var aliasGetCmd = &cobra.Command{
	Use:   "alias",
	Short: "Get existing mail aliases for users.",
	Long:  `Get all mail aliases for the users of the server, use the domain-flag to filter the output.`,
	Args:  cobra.NoArgs,
	Run:   getAlias,
}

var aliasAddCmd = &cobra.Command{
	Use:   "add",
	Short: "Add an email alias",
	Long:  `Add an email alias`,
	Args:  cobra.NoArgs,
	Run:   addAlias,
}

var aliasDeleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete an email alias",
	Long:  `Delete an email alias`,
	Args:  cobra.NoArgs,
	Run:   delAlias,
}

func getAlias(cmd *cobra.Command, args []string) {
	format := miab.PLAIN
	if f, err := cmd.Flags().GetString("format"); err == nil {
		format = miab.Format(f)
	}
	domain, _ := cmd.Flags().GetString("domain")

	aliasDomains, err := miab.GetAliases(&config)
	if err != nil {
		fmt.Printf("Error fetching email aliasDomains: %v\n", err)
		os.Exit(1)
	}

	if domain != "" {
		for _, aliasDomain := range aliasDomains {
			if aliasDomain.Domain == domain {
				fmt.Println(aliasDomain.ToString(format))
			}
		}
	}
	fmt.Println(aliasDomains.ToString(format))
}

func addAlias(cmd *cobra.Command, args []string) {
	email, _ := cmd.Flags().GetString("address")
	fwd, _ := cmd.Flags().GetString("forward")

	if err := miab.AddAlias(&config, email, fwd); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func delAlias(cmd *cobra.Command, args []string) {
	email, _ := cmd.Flags().GetString("address")

	if err := miab.DeleteAlias(&config, email); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
