package cmd

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"regexp"
	"strings"

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
		configContent := []byte("{\"access_key_id\": \"" + accessKeyID + "\", \"secret_access_key\": \"" + secretAccessKey + "\"}")
		err = ioutil.WriteFile(home+"/"+storageDir+"/"+configFile, configContent, 0600)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		// @todo do auth via api and obtain jwt
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
