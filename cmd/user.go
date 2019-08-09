package cmd

import (
	"fmt"
	"github.com/rverst/go-miab/miab"
	"github.com/spf13/cobra"
	"os"
)

func init() {
	rootCmd.AddCommand(userCmd)
	userCmd.AddCommand(userAddCmd, userRemoveCmd)
	userAddCmd.AddCommand(userAddPrivilege)
	userRemoveCmd.AddCommand(userRemovePrivilege)

	userCmd.Flags().String("domain", "", "domain to filter the list of email users")
	userCmd.Flags().String("format", "plain", "the output format (plain, csv, json, yaml)")
	userAddCmd.PersistentFlags().String("email", "", "email address of the user [mandatory]")
	userRemoveCmd.PersistentFlags().String("email", "", "email address of the user [mandatory]")
	userAddCmd.Flags().String("pass", "", "password for the new user [mandatory]")

	_ = userAddCmd.MarkPersistentFlagRequired("email")
	_ = userAddCmd.MarkFlagRequired("pass")

	_ = userRemoveCmd.MarkPersistentFlagRequired("email")
}

var userCmd = &cobra.Command{
	Use:   "user",
	Short: "Get existing mail users.",
	Long:  `Get all mail user of the server, use the domain-flag to filter the output.`,
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {

		format := miab.PLAIN
		if f, err := cmd.Flags().GetString("format"); err == nil {
			format = miab.Formats(f)
		}

		users, err := miab.GetUsers(&config)
		if err != nil {
			fmt.Printf("Error fetching email users: %v\n", err)
			os.Exit(1)
		}

		d, err := cmd.Flags().GetString("domain")
		if err == nil && d != "" {
			for _, u := range users {
				if u.Domain == d {
					u.Print(format)
					return
				}
			}
		}

		users.Print(format)
	},
}

var userAddCmd = &cobra.Command{
	Use:   "add",
	Short: "Add an mail user",
	Long:  `Add an mail user`,
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {

		email, _ := cmd.Flags().GetString("email")
		pass, _ := cmd.Flags().GetString("pass")

		if err := miab.AddUser(&config, email, pass); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	},
}

var userRemoveCmd = &cobra.Command{
	Use:   "remove",
	Short: "Remove an mail user",
	Long:  `Remove an mail user`,
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {

		email, _ := cmd.Flags().GetString("email")

		if err := miab.RemoveUser(&config, email); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	},
}

var userAddPrivilege = &cobra.Command{
	Use:   "privilege",
	Short: "Add the admin privilege to an mail user",
	Long:  `Add the admin privilege to an mail user`,
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {

		email, _ := cmd.Flags().GetString("email")

		if err := miab.AddPrivileges(&config, email); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	},
}

var userRemovePrivilege = &cobra.Command{
	Use:   "privilege",
	Short: "Remove the admin privilege to an mail user",
	Long:  `Remove the admin privilege to an mail user`,
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {

		email, _ := cmd.Flags().GetString("email")

		if err := miab.RemovePrivileges(&config, email); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	},
}