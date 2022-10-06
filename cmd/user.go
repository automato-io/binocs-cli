package cmd

import (
	"crypto/sha1"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"os"
	"regexp"
	"runtime"
	"time"

	"github.com/AlecAivazis/survey/v2"
	"github.com/ProtonMail/gopenpgp/v2/helper"
	util "github.com/automato-io/binocs-cli/util"
	"github.com/automato-io/tablewriter"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// User comes from the API as a JSON
type User struct {
	Name          string `json:"name,omitempty"`
	Email         string `json:"email,omitempty"`
	Timezone      string `json:"timezone,omitempty"`
	CreditBalance int    `json:"credit_balance,omitempty"`
	Created       string `json:"created,omitempty"`
}

type ConnectToken struct {
	ClientKey string `json:"c"`
	Time      int64  `json:"t"`
	Location  string `json:"l"`
	OS        string `json:"o"`
	Hostname  string `json:"h"`
}

// `user update` flags
var (
	userFlagName     string
	userFlagTimezone string
)

const (
	validUserNamePattern = `^[\p{L}\p{N}_\s\/\-\.]{0,100}$`
	storageDir           = ".binocs"
	configFile           = "config.json"
	connectTokenSalt     = "GVuUEdQ"
)

func init() {
	rootCmd.AddCommand(userCmd)
	userCmd.AddCommand(userUpdateCmd)

	userUpdateCmd.Flags().StringVarP(&userFlagName, "name", "n", "", "Your name")
	userUpdateCmd.Flags().StringVarP(&userFlagTimezone, "timezone", "t", "", "Your timezone")

	rootCmd.AddCommand(loginCmd)
	rootCmd.AddCommand(logoutCmd)
}

var userCmd = &cobra.Command{
	Use:   "user",
	Short: "Display information about current Binocs user",
	Long: `
Display information about current Binocs user
`,
	Aliases:           []string{"account"},
	DisableAutoGenTag: true,
	Run: func(cmd *cobra.Command, args []string) {
		util.VerifyAuthenticated()

		spin.Start()
		defer spin.Stop()
		spin.Suffix = colorFaint.Sprint(" loading user...")
		respJSON, err := fetchUser()
		if err != nil {
			spin.Stop()
			fmt.Println(err)
			os.Exit(1)
		}

		timezone, err := time.LoadLocation(respJSON.Timezone)
		if err != nil {
			spin.Stop()
			fmt.Println("Unknown timezone " + respJSON.Timezone)
			os.Exit(1)
		}

		tableCheckCellContent := colorBold.Sprint(`Name: `) + respJSON.Name + "\n" +
			colorBold.Sprint(`E-mail: `) + respJSON.Email + "\n" +
			colorBold.Sprint(`Timezone: `) + timezone.String() + "\n" +
			colorBold.Sprint(`Credit balance: `) + fmt.Sprint(respJSON.CreditBalance)

		var tableData [][]string
		tableData = append(tableData, []string{tableCheckCellContent})

		columnDefinitions := []tableColumnDefinition{
			{
				Header:    "USER",
				Priority:  1,
				Alignment: tablewriter.ALIGN_LEFT,
			},
		}

		table := composeTable(tableData, columnDefinitions)
		spin.Stop()
		if respJSON.CreditBalance == 0 {
			printZeroCreditsWarning()
		}
		table.Render()
	},
}

var userUpdateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update current Binocs user",
	Long: `
Update current Binocs user.

This command is interactive and asks user for parameters that were not provided as flags.
`,
	DisableAutoGenTag: true,
	Run: func(cmd *cobra.Command, args []string) {
		util.VerifyAuthenticated()

		var err error
		var match bool
		var tpl string
		var (
			flagName     string
			flagTimezone string
		)
		flagName = userFlagName
		flagTimezone = userFlagTimezone

		spin.Start()
		defer spin.Stop()
		spin.Suffix = colorFaint.Sprint(" loading user...")
		respJSON, err := fetchUser()
		if err != nil {
			spin.Stop()
			fmt.Println(err)
			os.Exit(1)
		}
		spin.Stop()

		match, err = regexp.MatchString(validUserNamePattern, flagName)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		} else if !match || flagName == "" {
			validate := func(val interface{}) error {
				match, err = regexp.MatchString(validUserNamePattern, val.(string))
				if err != nil {
					return err
				} else if !match {
					return errors.New("invalid name format")
				}
				return nil
			}
			prompt := &survey.Input{
				Message: "Enter your name:",
				Default: respJSON.Name,
			}
			err = survey.AskOne(prompt, &flagName, survey.WithValidator(validate))
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
		}

		_, err = time.LoadLocation(flagTimezone)
		if err != nil {
			fmt.Println("Invalid timezone " + flagTimezone)
			os.Exit(1)
		} else if flagTimezone == "" {
			prompt := &survey.Select{
				Message: "Enter your timezone:",
				Options: util.ListTimezones(),
				Default: respJSON.Timezone,
			}
			err = survey.AskOne(prompt, &flagTimezone)
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

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Login to you Binocs account",
	Long: `
Login to you Binocs account.
`,
	Aliases:           []string{"auth"},
	DisableAutoGenTag: true,
	Run: func(cmd *cobra.Command, args []string) {
		clientKey := viper.Get("client_key")
		if clientKey == nil {
			fmt.Println("Cannot read Client Key")
			os.Exit(1)
		}
		connectToken := generateConnectToken(clientKey.(string))
		tpl := `Please visit the following URL in your browser.
It will link your Binocs user account with this Binocs CLI installation.
		
https://binocs.sh/connect/` + connectToken + `
`
		fmt.Println(tpl)
	},
}

var logoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "Logout",
	Long: `
Log out of Binocs on this machine.
`,
	Aliases:           []string{},
	DisableAutoGenTag: true,
	Run: func(cmd *cobra.Command, args []string) {
		util.VerifyAuthenticated()

		var err error
		viper.Set("client_key", generateClientKey())
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
		tpl := `You were logged out of Binocs.`
		fmt.Println(tpl)
	},
}

func fetchUser() (User, error) {
	var respJSON User
	respData, err := util.BinocsAPI("/user", http.MethodGet, []byte{})
	if err != nil {
		return respJSON, err
	}
	err = json.Unmarshal(respData, &respJSON)
	return respJSON, err
}

func generateClientKey() string {
	rand.Seed(time.Now().UnixNano())
	h := sha1.New()
	io.WriteString(h, fmt.Sprintf("%d (.)(.) %v", rand.Int(), time.Now().UnixNano()))
	return fmt.Sprintf("%x", h.Sum(nil))
}

func generateConnectToken(clientKey string) string {
	token := ConnectToken{
		ClientKey: clientKey,
		Time:      time.Now().Unix(),
		Location:  "", // @todo determine user's location
		OS:        getUserOS(),
		Hostname:  getUserHostname(),
	}
	tokenJson, err := json.Marshal(token)
	if err != nil {
		fmt.Println("Cannot marshal Connect Token")
		os.Exit(1)
	}
	tokenPGP, err := helper.EncryptMessageWithPassword([]byte(connectTokenSalt), string(tokenJson))
	if err != nil {
		fmt.Println("Cannot generate Connect Token (E3)")
		os.Exit(1)
	}
	tokenBase64 := base64.StdEncoding.EncodeToString([]byte(tokenPGP))
	return tokenBase64
}

func getUserOS() string {
	userOS := ""
	if runtime.GOOS == "darwin" {
		userOS = "MacOS"
	} else if runtime.GOOS == "windows" {
		userOS = "Windows"
	} else if runtime.GOOS == "linux" {
		userOS = "Linux"
	}
	return userOS
}

func getUserHostname() string {
	hostname, _ := os.Hostname()
	return hostname
}
