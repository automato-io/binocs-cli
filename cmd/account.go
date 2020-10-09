package cmd

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strings"

	util "github.com/automato-io/binocs-cli/util"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	validAccessKeyIDPattern     = `^[A-Z0-9]{10}$`
	validSecretAccessKeyPattern = `^[a-z0-9]{16}$`
	storageDir                  = ".binocs"
	configFile                  = "config.json"
)

func init() {
	rootCmd.AddCommand(accountCmd)
	rootCmd.AddCommand(loginCmd)
	rootCmd.AddCommand(logoutCmd)
}

var accountCmd = &cobra.Command{
	Use:   "account",
	Short: "Manage your binocs account",
	Long:  `...`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("manage account stuff")
	},
}

var loginCmd = &cobra.Command{
	Use:     "login",
	Short:   "Login to binocs",
	Long:    `Use your Access Key ID and Secret Access Key and login to binocs`,
	Aliases: []string{"auth"},
	Run: func(cmd *cobra.Command, args []string) {
		var err error
		var match bool

		reader := bufio.NewReader(os.Stdin)
		fmt.Print("Enter your Access Key ID: ")
		accessKeyID, _ := reader.ReadString('\n')
		accessKeyID = strings.TrimSpace(accessKeyID)
		fmt.Print("Enter your Secret Access Key: ")
		secretAccessKey, _ := reader.ReadString('\n')
		secretAccessKey = strings.TrimSpace(secretAccessKey)

		match, err = regexp.MatchString(validAccessKeyIDPattern, accessKeyID)
		if err != nil || match == false {
			fmt.Println("Provided Access Key ID is invalid")
			os.Exit(1)
		}

		match, err = regexp.MatchString(validSecretAccessKeyPattern, secretAccessKey)
		if err != nil || match == false {
			fmt.Println("Provided Secret Access Key is invalid")
			os.Exit(1)
		}

		viper.Set("access_key_id", accessKeyID)
		viper.Set("secret_access_key", secretAccessKey)
		err = viper.WriteConfigAs(viper.ConfigFileUsed())
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		util.BinocsAPIGetAccessToken(accessKeyID, secretAccessKey)

		tpl := `Credentials OK
You are authenticated as dev@automato.io
`
		fmt.Print(tpl)
	},
}

var logoutCmd = &cobra.Command{
	Use:     "logout",
	Short:   "Logout from binocs",
	Long:    "Logs you out of the binocs account on this machine",
	Aliases: []string{},
	Run: func(cmd *cobra.Command, args []string) {
		var err error
		viper.Set("access_key_id", "")
		viper.Set("secret_access_key", "")
		err = viper.WriteConfigAs(viper.ConfigFileUsed())
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		err = util.ResetAccessToken()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		tpl := `You were successfully logged out of binocs
`
		fmt.Print(tpl)
	},
}
