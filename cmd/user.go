package cmd

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"regexp"
	"strings"

	util "github.com/automato-io/binocs-cli/util"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// User comes from the API as a JSON
type User struct {
	Name     string `json:"name,omitempty"`
	Email    string `json:"email,omitempty"`
	Timezone string `json:"timezone,omitempty"`
	Created  string `json:"created,omitempty"`
}

const (
	validAccessKeyIDPattern     = `^[A-Z0-9]{10}$`
	validSecretAccessKeyPattern = `^[a-z0-9]{16}$`
	storageDir                  = ".binocs"
	configFile                  = "config.json"
)

func init() {
	rootCmd.AddCommand(userCmd)
	userCmd.AddCommand(userUpdateCmd)
	userCmd.AddCommand(generateKeyCmd)
	userCmd.AddCommand(invalidateKeyCmd)
	rootCmd.AddCommand(loginCmd)
	rootCmd.AddCommand(logoutCmd)
}

var userCmd = &cobra.Command{
	Use:   "user",
	Short: "Display information about your binocs user",
	Long: `
Display information about your binocs user
`,
	Aliases: []string{"account"},
	Run: func(cmd *cobra.Command, args []string) {
		spin.Start()
		spin.Suffix = " loading user..."
		respData, err := util.BinocsAPI("/user", http.MethodGet, []byte{})
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		var respJSON User
		err = json.Unmarshal(respData, &respJSON)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		// Table "main"

		tableMainCheckCellContent := `Name: ` + respJSON.Name + `
E-mail: ` + respJSON.Email + `
Timezone: ` + respJSON.Timezone + ``

		tableMain := tablewriter.NewWriter(os.Stdout)
		tableMain.SetHeader([]string{"BINOCS USER"})
		tableMain.SetAutoWrapText(false)
		tableMain.Append([]string{tableMainCheckCellContent})

		spin.Stop()
		tableMain.Render()
	},
}

var userUpdateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update current binocs user",
	Long: `
Update any of the following parameters of the current binocs user: 
email, password, name, billing-address, timezone
`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("update user")
	},
}

var generateKeyCmd = &cobra.Command{
	Use:   "generate-key",
	Short: "Generate new Access ID and Secret Key",
	Long: `
Generate new Access ID and Secret Key
`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("gen key")
	},
}

var invalidateKeyCmd = &cobra.Command{
	Use:   "invalidate-key",
	Short: "Deny future login attempts using this key",
	Long: `
Deny future login attempts using this key
`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("key rm")
	},
}

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Login to binocs",
	Long: `
Use your Access Key ID and Secret Access Key and login to binocs
`,
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

		// @todo load user
		tpl := `Credentials OK
You are authenticated as dev@automato.io
`
		fmt.Print(tpl)
	},
}

var logoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "Logout",
	Long: `
Log out of the binocs user on this machine
`,
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
