package cmd

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"regexp"
	"strings"

	util "github.com/automato-io/binocs-cli/util"
	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
)

const (
	validAccessKeyIDPattern     = `^[A-Z0-9]{10}$`
	validSecretAccessKeyPattern = `^[a-z0-9]{16}$`
	storageDir                  = ".binocs"
	configFile                  = "config.json"
	jwtFile                     = "auth.json"
)

// AuthResponse comes from the API
type AuthResponse struct {
	AccessToken string `json:"access_token"`
}

func init() {
	rootCmd.AddCommand(loginCmd)
}

var loginCmd = &cobra.Command{
	Use:     "login",
	Short:   "Login to binocs",
	Long:    `Use your Access Key ID and Secret Access Key and login to binocs`,
	Aliases: []string{"auth"},
	Run: func(cmd *cobra.Command, args []string) {
		var err error
		var match bool

		home, err := homedir.Dir()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

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

		resetConfigStorage(home)
		storeConfig(home, accessKeyID, secretAccessKey)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		postData := []byte("{\"access_key_id\": \"" + accessKeyID + "\", \"secret_access_key\": \"" + secretAccessKey + "\"}")
		respData, err := util.BinocsAPI2("/authenticate", http.MethodPost, postData)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		var respJSON AuthResponse
		err = json.Unmarshal(respData, &respJSON)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		err = storeAccessToken(home, &respJSON)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	},
}

func resetConfigStorage(home string) {
	var err error
	err = os.Mkdir(home+"/"+storageDir, 0755)
	if err != nil && !os.IsExist(err) {
		fmt.Println(err)
		os.Exit(1)
	}
	err = os.Remove(home + "/" + storageDir + "/" + configFile)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func storeConfig(home, accessKeyID, secretAccessKey string) error {
	configContent := []byte("{\"access_key_id\": \"" + accessKeyID + "\", \"secret_access_key\": \"" + secretAccessKey + "\"}")
	return ioutil.WriteFile(home+"/"+storageDir+"/"+configFile, configContent, 0600)
}

func storeAccessToken(home string, d *AuthResponse) error {
	authContent := []byte("{\"access_token\": \"" + d.AccessToken + "\"}")
	return ioutil.WriteFile(home+"/"+storageDir+"/"+jwtFile, authContent, 0600)
}
