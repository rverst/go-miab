package command

import (
	"fmt"
	"github.com/mitchellh/go-homedir"
	"github.com/rverst/go-miab/miab"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"os"
	"path"
)

//these variables are set in the build process via `ldflags`
var (
	Version    = "unknown"
	CommitHash = "unknown"
	BuildDate  = "unknown"
)

var cfgFile string
var config miab.Config

var rootCmd = &cobra.Command{
	Use:   "miab",
	Short: "Miab is a cli tool for the Mail-in-a-Box API",
	Long: `A cli tool for the Mail-in-a-Box API
Mail-in-a-Box can be found at https://mailinabox.email
The source is available at https://github.com/rverst/go-miab`,
}

func init() {

	rootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "", "config file (default is $HOME/.config/go-miab/miab.yaml)")
	rootCmd.PersistentFlags().StringP("user", "u", "", "user to authenticate, can be set via environment variable (MIAB_USER) or config file")
	rootCmd.PersistentFlags().StringP("password", "p", "", "password to authenticate, can be set via environment variable (MIAB_PASSWORD) or config file")
	rootCmd.PersistentFlags().StringP("endpoint", "e", "", "api endpoint, can be set via environment variable (MIAB_ENDPOINT) or config file")

	viper.SetEnvPrefix("miab")

	_ = viper.BindPFlag("user", rootCmd.PersistentFlags().Lookup("user"))
	_ = viper.BindPFlag("password", rootCmd.PersistentFlags().Lookup("password"))
	_ = viper.BindPFlag("endpoint", rootCmd.PersistentFlags().Lookup("endpoint"))
}

func initConfig(cmd *cobra.Command, args []string) {

	viper.AutomaticEnv()
	cfg, err := miab.NewConfig(viper.GetString("user"), viper.GetString("password"), viper.GetString("endpoint"))
	if err != nil && (viper.GetString("user") == "" || viper.GetString("password") == "" || viper.GetString("endpoint") == "") {
		// not all parameters might have been provided, let's try the config file
		if cfgFile != "" {
			viper.SetConfigFile(cfgFile)
		} else {

			home, err := homedir.Dir()
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}

			viper.AddConfigPath(path.Join(home, ".config", "go-miab"))
			viper.SetConfigName("miab")
		}

		if err := viper.ReadInConfig(); err != nil {
			fmt.Println("Can't read config:", err)
			os.Exit(1)
		}
		cfg, err = miab.NewConfig(viper.GetString("user"), viper.GetString("password"), viper.GetString("endpoint"))
		if err != nil {
			fmt.Println("Config is invalid:", err)
			os.Exit(1)
		}
	} else if err != nil {
		fmt.Println("Config is invalid:", err)
		os.Exit(1)
	}
	config = *cfg
}

// Execute is the main entrance point for the cli parser an should be called from `func main()`
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
