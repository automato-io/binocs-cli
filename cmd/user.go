package cmd

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"

	util "github.com/automato-io/binocs-cli/util"
	"github.com/manifoldco/promptui"
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

// AccessKeyPair struct
type AccessKeyPair struct {
	AccessKeyID     string `json:"access_key_id,omitempty"`
	SecretAccessKey string `json:"secret_access_key,omitempty"`
}

// `user update` flags
var (
	userFlagName     string
	userFlagTimezone string
)

const (
	validAccessKeyIDPattern     = `^[A-Z0-9]{10}$`
	validSecretAccessKeyPattern = `^[a-z0-9]{16}$`
	storageDir                  = ".binocs"
	configFile                  = "config.json"
)

func init() {
	rootCmd.AddCommand(userCmd)
	userCmd.AddCommand(userUpdateCmd)

	userUpdateCmd.Flags().StringVarP(&userFlagName, "name", "n", "", "Your name")
	userUpdateCmd.Flags().StringVarP(&userFlagTimezone, "timezone", "t", "", "Your timezone")

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
	Aliases:           []string{"account"},
	DisableAutoGenTag: true,
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
name, timezone
`,
	DisableAutoGenTag: true,
	Run: func(cmd *cobra.Command, args []string) {
		var err error
		var match bool
		var tpl string
		var (
			flagName     string
			flagTimezone string
		)
		flagName = userFlagName
		flagTimezone = userFlagTimezone

		match, err = regexp.MatchString(validNamePattern, flagName)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		} else if match == false || flagName == "" {
			validate := func(input string) error {
				match, err = regexp.MatchString(validNamePattern, input)
				if err != nil {
					return errors.New("Invalid input")
				} else if match == false {
					return errors.New("Invalid input value")
				}
				return nil
			}
			prompt := promptui.Prompt{
				Label:    "Enter a valid name",
				Validate: validate,
			}
			flagName, err = prompt.Run()
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
		}

		_, err = time.LoadLocation(flagTimezone)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		} else if match == false || flagTimezone == "" {
			validate := func(input string) error {
				_, err := time.LoadLocation(input)
				if err != nil {
					return errors.New("Invalid timezone value")
				}
				return nil
			}
			prompt := promptui.Prompt{
				Label:    "Enter a valid timezone",
				Validate: validate,
			}
			flagTimezone, err = prompt.Run()
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
		}

		user := User{
			Name:     flagName,
			Timezone: flagTimezone,
		}
		postData, err := json.Marshal(user)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		respData, err := util.BinocsAPI("/user", http.MethodPut, postData)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		err = json.Unmarshal(respData, &user)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		tpl = "User " + user.Email + " updated successfully"
		fmt.Println(tpl)
	},
}

var generateKeyCmd = &cobra.Command{
	Use:   "generate-key",
	Short: "Generate new Access ID and Secret Key",
	Long: `
Generate new Access ID and Secret Key
`,
	DisableAutoGenTag: true,
	Run: func(cmd *cobra.Command, args []string) {
		var tpl string
		spin.Start()
		spin.Suffix = " generating new key pair..."
		respData, err := util.BinocsAPI("/user/generate-key", http.MethodPost, []byte{})
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		var respJSON AccessKeyPair
		err = json.Unmarshal(respData, &respJSON)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		tpl = `Please keep your secret safe. We will not be able to show it to you again.
Type ` + "`binocs login`" + ` to authenticate to Binocs using your key pair.

This is your new key pair: 
Access Key ID: ` + respJSON.AccessKeyID + `
Secret Access Key: ` + respJSON.SecretAccessKey + `
`
		spin.Stop()
		fmt.Print(tpl)
	},
}

var invalidateKeyCmd = &cobra.Command{
	Use:   "invalidate-key",
	Short: "Deny future login attempts using this key",
	Long: `
Deny future login attempts using this key
`,
	Args:              cobra.ExactArgs(1),
	DisableAutoGenTag: true,
	Run: func(cmd *cobra.Command, args []string) {
		var tpl string
		spin.Start()
		spin.Suffix = " invalidating key " + args[0]
		accessKey := AccessKeyPair{
			AccessKeyID: args[0],
		}
		postData, err := json.Marshal(accessKey)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		_, err = util.BinocsAPI("/user/invalidate-key", http.MethodPost, postData)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		tpl = `Successfully invalidated Access Key ID ` + args[0] + `
`
		spin.Stop()
		fmt.Print(tpl)
	},
}

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Login to binocs",
	Long: `
Use your Access Key ID and Secret Access Key and login to binocs
`,
	Aliases:           []string{"auth"},
	DisableAutoGenTag: true,
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
	Aliases:           []string{},
	DisableAutoGenTag: true,
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
