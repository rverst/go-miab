package command

import (
	"fmt"
	"github.com/rverst/go-miab/miab"
	"github.com/spf13/cobra"
	"os"
)

func init() {
	rootCmd.AddCommand(userGetCmd)
	userGetCmd.AddCommand(userAddCmd, userRemoveCmd)
	userAddCmd.AddCommand(userAddPrivilege)
	userRemoveCmd.AddCommand(userDeletePrivilege)

	userGetCmd.Flags().String("domain", "", "domain to filter the list of email users")
	userGetCmd.Flags().String("format", "plain", "the output format (plain, csv, json, yaml)")
	userAddCmd.PersistentFlags().String("email", "", "email address of the user [mandatory]")
	userRemoveCmd.PersistentFlags().String("email", "", "email address of the user [mandatory]")
	userAddCmd.Flags().String("pass", "", "password for the new user [mandatory]")

	_ = userAddCmd.MarkPersistentFlagRequired("email")
	_ = userAddCmd.MarkFlagRequired("pass")

	_ = userRemoveCmd.MarkPersistentFlagRequired("email")
}

var userGetCmd = &cobra.Command{
	Use:   "user",
	Short: "Get existing mail users.",
	Long:  `Get all mail user of the server, use the domain-flag to filter the output.`,
	Args:  cobra.NoArgs,
	Run:   getUser,
}

var userAddCmd = &cobra.Command{
	Use:   "add",
	Short: "Add an mail user",
	Long:  `Add an mail user`,
	Args:  cobra.NoArgs,
	Run:   addUser,
}

var userRemoveCmd = &cobra.Command{
	Use:   "del",
	Short: "Delete an mail user",
	Long:  `Delete an mail user`,
	Args:  cobra.NoArgs,
	Run:   delUser,
}

var userAddPrivilege = &cobra.Command{
	Use:   "privilege",
	Short: "Add the admin privilege to an mail user",
	Long:  `Add the admin privilege to an mail user`,
	Args:  cobra.NoArgs,
	Run:   addPrivilege,
}

var userDeletePrivilege = &cobra.Command{
	Use:   "privilege",
	Short: "Delete the admin privilege to an mail user",
	Long:  `Delete the admin privilege to an mail user`,
	Args:  cobra.NoArgs,
	Run:   delPrivilege,
}

func getUser(cmd *cobra.Command, args []string) {
	format := miab.PLAIN
	if f, err := cmd.Flags().GetString("format"); err == nil {
		format = miab.Format(f)
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
				fmt.Println(u.ToString(format))
				return
			}
		}
	}

	fmt.Println(users.ToString(format))
}

func addUser(cmd *cobra.Command, args []string) {
	email, _ := cmd.Flags().GetString("email")
	pass, _ := cmd.Flags().GetString("pass")

	if err := miab.AddUser(&config, email, pass); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func delUser(cmd *cobra.Command, args []string) {
	email, _ := cmd.Flags().GetString("email")

	if err := miab.DeleteUser(&config, email); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func addPrivilege(cmd *cobra.Command, args []string) {
	email, _ := cmd.Flags().GetString("email")

	if err := miab.AddPrivileges(&config, email); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func delPrivilege(cmd *cobra.Command, args []string) {
	email, _ := cmd.Flags().GetString("email")

	if err := miab.RemovePrivileges(&config, email); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
