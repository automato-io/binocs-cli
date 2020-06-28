package cmd

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"

	util "github.com/automato-io/binocs-cli/util"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
)

// Check comes from the API as a JSON, or from user input as `check add` flags
type Check struct {
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
	LastStatusDuration         string   `json:"last_status_duration"`
	Created                    string   `json:"created"`
	Updated                    string   `json:"updated"`
}

// MetricsResponse comes from the API as a JSON
type MetricsResponse struct {
	Apdex  string `json:"apdex"`
	MRT    string `json:"mrt"`
	Uptime string `json:"uptime"`
}

// ApdexResponse comes from the API as a JSON
type ApdexResponse struct {
	Apdex string `json:"apdex"`
	From  string `json:"from"`
	To    string `json:"to"`
}

// `check add` flags and defaults
var (
	flagName                       string
	flagURL                        string
	flagMethod                     string
	flagInterval                   int
	flagTarget                     float64
	flagRegions                    []string
	flagUpCodes                    []string
	flagUpConfirmationsThreshold   int
	flagDownConfirmationsThreshold int
	flagChannels                   []string
)

const (
	flagMandatory = true
	flagOptional  = false
)

const (
	supportedIntervalMinimum               = 5
	supportedIntervalMaximum               = 900
	supportedTargetMinimum                 = 0.01
	supportedTargetMaximum                 = 10.0
	validNamePattern                       = `^[a-zA-Z0-9_\ \-\.]{0,25}$`
	validUpCodePattern                     = `^([1-5]{1}[0-9]{2}-[1-5]{1}[0-9]{2}|([1-5]{1}(([0-9]{2}|[0-9]{1}x)|xx)))$`
	validURLPattern                        = `^https?:\/\/(www\.)?[-a-zA-Z0-9@:%._\+~#=]{2,256}\.[a-z]{2,4}\b([-a-zA-Z0-9@:%_\+.~#?&//=]*)$`
	supportedConfirmationsThresholdMinimum = 1
	supportedConfirmationsThresholdMaximum = 10
)

func init() {
	rootCmd.AddCommand(checkCmd)
	checkCmd.AddCommand(checkAddCmd)
	checkCmd.AddCommand(checkInspectCmd)
	checkCmd.AddCommand(checkUpdateCmd)
	checkCmd.AddCommand(checkDeleteCmd)

	checkAddCmd.Flags().StringVarP(&flagName, "name", "n", "", "Check alias")
	checkAddCmd.Flags().StringVarP(&flagURL, "URL", "u", "", "URL to check")
	checkAddCmd.Flags().StringVarP(&flagMethod, "method", "m", "", "HTTP method (GET, POST, ...)")
	checkAddCmd.Flags().IntVarP(&flagInterval, "interval", "i", 30, "How often we check the URL, in seconds")
	checkAddCmd.Flags().Float64VarP(&flagTarget, "target", "t", 0.7, "Response time in miliseconds for Apdex = 1.0")
	checkAddCmd.Flags().StringSliceVarP(&flagRegions, "regions", "r", []string{"all"}, "From where we check the URL, choose `all` or any combination of `us-east-1`, `eu-central-1`, ...")
	checkAddCmd.Flags().StringSliceVarP(&flagUpCodes, "up_codes", "", []string{"200-302"}, "What are the Up HTTP response codes, e.g. `2xx` or `200-302`")
	checkAddCmd.Flags().IntVarP(&flagUpConfirmationsThreshold, "up_confirmations_threshold", "", 2, "How many subsequent Up responses before triggering notifications")
	checkAddCmd.Flags().IntVarP(&flagDownConfirmationsThreshold, "down_confirmations_threshold", "", 2, "How many subsequent Down responses before triggering notifications")
	checkAddCmd.Flags().StringSliceVarP(&flagChannels, "channels", "", []string{"email", "slack"}, "Where you want to receive notifications for this check, `email`, `slack` or both?")
	checkAddCmd.Flags().SortFlags = false
}

var checkCmd = &cobra.Command{
	Use:     "check",
	Short:   "Manage your checks - checks are the endpoints monitored by binocs",
	Long:    `...`,
	Aliases: []string{"checks"},
	Example: "",
	Run: func(cmd *cobra.Command, args []string) {
		// @todo filter by status, pseudo-fulltext
		// @todo set -interval 24h|3d default 24h
		respData, err := util.BinocsAPI("/checks", http.MethodGet, []byte{})
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		respJSON := make([]Check, 0)
		decoder := json.NewDecoder(bytes.NewBuffer(respData))
		err = decoder.Decode(&respJSON)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		var tableData [][]string
		for i, v := range respJSON {
			// @todo accept period, region params
			metricsData, err := util.BinocsAPI("/checks/"+strconv.Itoa(v.ID)+"/metrics?period=day", http.MethodGet, []byte{})
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
			var metrics MetricsResponse
			err = json.Unmarshal(metricsData, &metrics)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
			if metrics.Uptime == "100.00" {
				metrics.Uptime = "100"
			}

			apdexData, err := util.BinocsAPI("/checks/"+strconv.Itoa(v.ID)+"/apdex?period=day", http.MethodGet, []byte{})
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
			apdex := make([]ApdexResponse, 0)
			decoder := json.NewDecoder(bytes.NewBuffer(apdexData))
			err = decoder.Decode(&apdex)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}

			apdexChart := drawCompactApdexChart(apdex, 3)

			tableRow := []string{
				strconv.Itoa(i + 1), v.Name, v.URL, statusName[v.LastStatus] + " " + v.LastStatusDuration, v.LastStatusCode, strconv.Itoa(v.Interval) + " s", fmt.Sprintf("%.3f s", v.Target), metrics.MRT + " s", fmt.Sprintf("%v %%", metrics.Uptime), metrics.Apdex, apdexChart,
			}
			tableData = append(tableData, tableRow)
		}
		table := tablewriter.NewWriter(os.Stdout)
		table.SetHeader([]string{"#", "NAME", "URL", "STATUS", "HTTP CODE", "INTERVAL", "TARGET", "MRT", "UPTIME", "APDEX", "APDEX 24 h"})

		for _, v := range tableData {
			table.Append(v)
		}
		table.Render()
	},
}

func askInput(paramName string, times int, required bool, prompt string, validPattern string) (string, error) {
	var err error
	var match bool
	if times < 1 {
		return "", fmt.Errorf("First `times` parameter must be > 1")
	}
	fmt.Print(prompt)
	reader := bufio.NewReader(os.Stdin)
	for i := 0; i < times; i = i + 1 {
		val, _ := reader.ReadString('\n')
		val = strings.TrimSpace(val)
		match, err = regexp.MatchString(validPattern, val)
		if err != nil {
			return "", err
		} else if match == false {
			if i == times-1 {
				fmt.Printf("Invalid input for %s\n", paramName)
			} else {
				fmt.Printf("Invalid input for %s\nTry again: ", paramName)
			}
			continue
		} else {
			return val, nil
		}
	}
	if required {
		return "", fmt.Errorf("Failed to set %s", paramName)
	} else {
		return "", nil
	}
}

var checkAddCmd = &cobra.Command{
	Use: "add",

	Run: func(cmd *cobra.Command, args []string) {
		var err error
		var match bool

		// check if Name is alphanum, space & normal chars, empty OK
		if flagName != "" {
			match, err = regexp.MatchString(validNamePattern, flagName)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			} else if match == false {
				flagName, err := askInput("name", 1, false, "Check alias (optional): ", validNamePattern)
				if err != nil {
					fmt.Println(err)
					os.Exit(1)
				} else if len(flagName) > 0 {
					// verbose ack
					fmt.Println("name okay: " + flagName)
				}
			} else {
				// verbose ack
				fmt.Println("name okay: " + flagName)
			}
		} else {
			flagName, err := askInput("name", 1, false, "Check alias (optional): ", validNamePattern)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			} else {
				fmt.Println("name okay: " + flagName)
			}
		}

		tpl := `zzz
`
		fmt.Print(tpl)
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
		var respJSON Check
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

func drawCompactApdexChart(apdex []ApdexResponse, compress int) string {
	var compressed []float64
	var compressStack []float64
	for i, v := range apdex {
		vf, _ := strconv.ParseFloat(v.Apdex, 32)
		compressStack = append(compressStack, vf)
		i = i + 1
		if i%compress == 0 || i == len(apdex) {
			var sum float64
			for _, f := range compressStack {
				sum = sum + f
			}
			compressed = append(compressed, sum/float64(len(compressStack)))
			compressStack = []float64{}
		}
	}

	var chart string
	for _, v := range compressed {
		var dot string
		if v < 0.125 {
			dot = "▁"
		} else if v < 0.250 {
			dot = "▂"
		} else if v < 0.375 {
			dot = "▃ "
		} else if v < 0.500 {
			dot = "▄ "
		} else if v < 0.625 {
			dot = "▅"
		} else if v < 0.750 {
			dot = "▆"
		} else if v < 0.875 {
			dot = "▇"
		} else {
			dot = "█"
		}
		chart = chart + dot
	}
	return chart
}
