package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"

	util "github.com/automato-io/binocs-cli/util"
	"github.com/spf13/cobra"
)

// CheckResponse comes from the API as a JSON
type CheckResponse struct {
	ID                         int      `json:"id"`
	Name                       string   `json:"name"`
	URL                        string   `json:"url"`
	Method                     string   `json:"method"`
	Interval                   int      `json:"interval"`
	Target                     float64  `json:"target"`
	Regions                    []string `json:"regions"`
	UpCodes                    string   `json:"up_codes"`
	UpConfirmationsThreshold   int      `json:"up_confirmations_threshold"`
	UpConfirmations            int      `json:"up_confirmations"`
	DownConfirmationsThreshold int      `json:"down_confirmations_threshold"`
	DownConfirmations          int      `json:"down_confirmations"`
	LastStatus                 int      `json:"last_status"`
	LastStatusCode             string   `json:"last_status_code"`
	Created                    string   `json:"created"`
	Updated                    string   `json:"updated"`
}

func init() {
	rootCmd.AddCommand(checkCmd)
	checkCmd.AddCommand(checkAddCmd)
	checkCmd.AddCommand(checkInspectCmd)
	checkCmd.AddCommand(checkUpdateCmd)
	checkCmd.AddCommand(checkDeleteCmd)
}

var checkCmd = &cobra.Command{
	Use:     "check",
	Short:   "Manage your checks - checks are the endpoints monitored by binocs",
	Long:    `...`,
	Aliases: []string{"checks"},
	Example: "",
	Run: func(cmd *cobra.Command, args []string) {
		respData, err := util.BinocsAPI("/checks", http.MethodGet, []byte{})
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		respJSON := make([]CheckResponse, 0)
		decoder := json.NewDecoder(bytes.NewBuffer(respData))
		err = decoder.Decode(&respJSON)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		fmt.Print(respJSON)
	},
}

var checkAddCmd = &cobra.Command{
	Use: "add",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("check add")
	},
}

var checkInspectCmd = &cobra.Command{
	Use:     "inspect",
	Aliases: []string{"view", "show"},
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			fmt.Println("RTFM")
			os.Exit(1)
		} else if _, err := strconv.Atoi(args[0]); err != nil {
			fmt.Println("only reference by id allowed atm")
			os.Exit(1)
		}
		respData, err := util.BinocsAPI("/checks/"+args[0], http.MethodGet, []byte{})
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		var respJSON CheckResponse
		err = json.Unmarshal(respData, &respJSON)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		tpl := `Name: ` + respJSON.Name + `
URL: ` + respJSON.URL + `
Method: ` + respJSON.Method + `
Status: ` + respJSON.LastStatusCode + `
Interval: ` + strconv.Itoa(respJSON.Interval) + ` s
Target response time: ` + fmt.Sprintf("%.3f", respJSON.Target) + ` s
Check from: ` + strings.Join(respJSON.Regions, ", ") + `
`
		fmt.Print(tpl)
	},
}

var checkUpdateCmd = &cobra.Command{
	Use: "update",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("check update")
	},
}

var checkDeleteCmd = &cobra.Command{
	Use:     "delete",
	Aliases: []string{"del"},
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("check delete")
	},
}
