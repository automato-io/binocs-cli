package cmd

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"

	util "github.com/automato-io/binocs-cli/util"
	"github.com/manifoldco/promptui"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
)

// Check comes from the API as a JSON, or from user input as `check add` flags
type Check struct {
	ID                         int      `json:"id,omitempty"`
	Name                       string   `json:"name,omitempty"`
	URL                        string   `json:"url,omitempty"`
	Method                     string   `json:"method,omitempty"`
	Interval                   int      `json:"interval,omitempty"`
	Target                     float64  `json:"target,omitempty"`
	Regions                    []string `json:"regions,omitempty"`
	UpCodes                    string   `json:"up_codes,omitempty"`
	UpConfirmationsThreshold   int      `json:"up_confirmations_threshold,omitempty"`
	UpConfirmations            int      `json:"up_confirmations,omitempty"`
	DownConfirmationsThreshold int      `json:"down_confirmations_threshold,omitempty"`
	DownConfirmations          int      `json:"down_confirmations,omitempty"`
	LastStatus                 int      `json:"last_status,omitempty"`
	LastStatusCode             string   `json:"last_status_code,omitempty"`
	LastStatusDuration         string   `json:"last_status_duration,omitempty"`
	Created                    string   `json:"created,omitempty"`
	Updated                    string   `json:"updated,omitempty"`
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

// `check ls` flags
var (
	flagRegion string
	flagStatus string
)

// `check add` flags
var (
	flagName                       string
	flagURL                        string
	flagMethod                     string
	flagInterval                   int
	flagTarget                     float64
	flagRegions                    []string
	flagUpCodes                    string
	flagUpConfirmationsThreshold   int
	flagDownConfirmationsThreshold int
	flagChannels                   []string
)

const (
	supportedIntervalMinimum               = 5
	supportedIntervalMaximum               = 900
	supportedTargetMinimum                 = 0.01
	supportedTargetMaximum                 = 10.0
	validNamePattern                       = `^[a-zA-Z0-9_\ \-\.]{0,25}$`
	validMethodPattern                     = `^(GET|HEAD|POST|PUT|DELETE)$` // hardcoded; reflects supportedHTTPMethods
	validUpCodePattern                     = `^([1-5]{1}[0-9]{2}-[1-5]{1}[0-9]{2}|([1-5]{1}(([0-9]{2}|[0-9]{1}x)|xx)))$`
	validURLPattern                        = `^https?:\/\/(www\.)?[-a-zA-Z0-9@:%._\+~#=]{2,256}\.[a-z]{2,4}\b([-a-zA-Z0-9@:%_\+.~#?&//=]*)$`
	validRegionPattern                     = `^[a-z0-9\-]{8,30}$`
	supportedConfirmationsThresholdMinimum = 1
	supportedConfirmationsThresholdMaximum = 10
)

var supportedHTTPMethods = map[string]bool{
	http.MethodGet:     true,
	http.MethodHead:    true,
	http.MethodPost:    true,
	http.MethodPut:     true,
	http.MethodPatch:   false,
	http.MethodDelete:  true,
	http.MethodConnect: false,
	http.MethodOptions: false,
	http.MethodTrace:   false,
}

func init() {
	rootCmd.AddCommand(checkCmd)
	checkCmd.AddCommand(checkAddCmd)
	checkCmd.AddCommand(checkInspectCmd)
	checkCmd.AddCommand(checkUpdateCmd)
	checkCmd.AddCommand(checkDeleteCmd)

	checkCmd.Flags().StringVarP(&flagRegion, "region", "r", "", "Display MRT, UPTIME and APDEX from the specified region only")
	checkCmd.Flags().StringVarP(&flagStatus, "status", "s", "", "List only \"up\" or \"down\" checks, default \"all\"")

	checkAddCmd.Flags().StringVarP(&flagName, "name", "n", "", "Check alias")
	checkAddCmd.Flags().StringVarP(&flagURL, "URL", "u", "", "URL to check")
	checkAddCmd.Flags().StringVarP(&flagMethod, "method", "m", "", "HTTP method (GET, POST, ...)")
	checkAddCmd.Flags().IntVarP(&flagInterval, "interval", "i", 30, "How often we check the URL, in seconds")
	checkAddCmd.Flags().Float64VarP(&flagTarget, "target", "t", 0.7, "Response time in miliseconds for Apdex = 1.0")
	checkAddCmd.Flags().StringSliceVarP(&flagRegions, "regions", "r", []string{"all"}, "From where we check the URL, choose `all` or any combination of `us-east-1`, `eu-central-1`, ...")
	checkAddCmd.Flags().StringVarP(&flagUpCodes, "up_codes", "", "200-302", "What are the Up HTTP response codes, e.g. `2xx` or `200-302`")
	checkAddCmd.Flags().IntVarP(&flagUpConfirmationsThreshold, "up_confirmations_threshold", "", 2, "How many subsequent Up responses before triggering notifications")
	checkAddCmd.Flags().IntVarP(&flagDownConfirmationsThreshold, "down_confirmations_threshold", "", 2, "How many subsequent Down responses before triggering notifications")
	checkAddCmd.Flags().StringSliceVarP(&flagChannels, "channels", "", []string{"email", "slack"}, "Where you want to receive notifications for this check, `email`, `slack` or both?")
	checkAddCmd.Flags().SortFlags = false

	checkUpdateCmd.Flags().StringVarP(&flagName, "name", "n", "", "Check alias")
	checkUpdateCmd.Flags().StringVarP(&flagURL, "URL", "u", "", "URL to check")
	checkUpdateCmd.Flags().StringVarP(&flagMethod, "method", "m", "", "HTTP method (GET, POST, ...)")
	checkUpdateCmd.Flags().IntVarP(&flagInterval, "interval", "i", 0, "How often we check the URL, in seconds")
	checkUpdateCmd.Flags().Float64VarP(&flagTarget, "target", "t", 0, "Response time in miliseconds for Apdex = 1.0")
	checkUpdateCmd.Flags().StringSliceVarP(&flagRegions, "regions", "r", []string{}, "From where we check the URL, choose `all` or any combination of `us-east-1`, `eu-central-1`, ...")
	checkUpdateCmd.Flags().StringVarP(&flagUpCodes, "up_codes", "", "", "What are the Up HTTP response codes, e.g. `2xx` or `200-302`")
	checkUpdateCmd.Flags().IntVarP(&flagUpConfirmationsThreshold, "up_confirmations_threshold", "", 0, "How many subsequent Up responses before triggering notifications")
	checkUpdateCmd.Flags().IntVarP(&flagDownConfirmationsThreshold, "down_confirmations_threshold", "", 0, "How many subsequent Down responses before triggering notifications")
	checkUpdateCmd.Flags().StringSliceVarP(&flagChannels, "channels", "", []string{}, "Where you want to receive notifications for this check, `email`, `slack` or both?")
	checkUpdateCmd.Flags().SortFlags = false
}

// @todo allow specifying -interval 24h|3d default 24h for mrt, uptime, apdex and apdex chart
var checkCmd = &cobra.Command{
	Use:     "check",
	Short:   "Manage your checks - checks are the endpoints monitored by binocs",
	Long:    `...`,
	Aliases: []string{"checks"},
	Example: "",
	Run: func(cmd *cobra.Command, args []string) {
		urlValues1 := url.Values{}
		urlValues2 := url.Values{
			"period": []string{"day"},
		}

		// @todo check against currently supported GET /regions
		match, err := regexp.MatchString(validRegionPattern, flagRegion)
		if len(flagRegion) > 0 && match == false {
			fmt.Println("Invalid region provided")
			os.Exit(1)
		} else if err == nil && match == true {
			urlValues2.Set("region", flagRegion)
		}

		flagStatus = strings.ToUpper(flagStatus)
		if flagStatus == statusNameUp || flagStatus == statusNameDown {
			urlValues1.Set("status", flagStatus)
		}

		respData, err := util.BinocsAPI("/checks?"+urlValues1.Encode(), http.MethodGet, []byte{})
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
			metricsData, err := util.BinocsAPI("/checks/"+strconv.Itoa(v.ID)+"/metrics?"+urlValues2.Encode(), http.MethodGet, []byte{})
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

			apdexData, err := util.BinocsAPI("/checks/"+strconv.Itoa(v.ID)+"/apdex?"+urlValues2.Encode(), http.MethodGet, []byte{})
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

			tableValueMRT := metrics.MRT + " s"
			if metrics.MRT == "" {
				tableValueMRT = "n/a"
			}
			tableValueUptime := fmt.Sprintf("%v %%", metrics.Uptime)
			if metrics.Uptime == "" {
				tableValueUptime = "n/a"
			}
			tableValueApdex := metrics.Apdex
			if metrics.Apdex == "" {
				tableValueApdex = "n/a"
				apdexChart = "n/a"
			}
			tableRow := []string{
				strconv.Itoa(i + 1), v.Name, v.URL, statusName[v.LastStatus] + " " + v.LastStatusDuration, v.LastStatusCode, strconv.Itoa(v.Interval) + " s", fmt.Sprintf("%.3f s", v.Target), tableValueMRT, tableValueUptime, tableValueApdex, apdexChart,
			}
			tableData = append(tableData, tableRow)
		}
		table := tablewriter.NewWriter(os.Stdout)
		table.SetHeader([]string{"#", "NAME", "URL", "STATUS", "HTTP CODE", "INTERVAL", "TARGET", "MRT", "UPTIME", "APDEX", "APDEX 24 h"})
		table.SetColumnAlignment([]int{tablewriter.ALIGN_DEFAULT, tablewriter.ALIGN_DEFAULT, tablewriter.ALIGN_DEFAULT, tablewriter.ALIGN_DEFAULT, tablewriter.ALIGN_DEFAULT,
			tablewriter.ALIGN_RIGHT, tablewriter.ALIGN_RIGHT, tablewriter.ALIGN_RIGHT, tablewriter.ALIGN_RIGHT, tablewriter.ALIGN_RIGHT, tablewriter.ALIGN_RIGHT,
		})
		for _, v := range tableData {
			table.Append(v)
		}
		table.Render()
	},
}

// mode = add|update
func checkAddOrUpdate(mode string, checkID int) {
	if mode != "add" && mode != "update" {
		fmt.Println("Unknown mode: " + mode)
		os.Exit(1)
	}
	var err error
	var match bool
	var tpl string

	if mode == "update" && flagName == "" {
		// pass
	} else {
		// check if Name is alphanum, space & normal chars, empty OK
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
				Label:    "Check alias (optional)",
				Validate: validate,
			}
			flagName, err = prompt.Run()
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
		}
	}

	if mode == "update" && flagURL == "" {
		// pass
	} else {
		// check if URL is url, empty not OK
		match, err = regexp.MatchString(validURLPattern, flagURL)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		} else if match == false {
			validate := func(input string) error {
				match, err = regexp.MatchString(validURLPattern, input)
				if err != nil {
					return errors.New("Invalid input")
				} else if match == false {
					return errors.New("Invalid input value")
				}
				return nil
			}
			prompt := promptui.Prompt{
				Label:    "URL to check",
				Validate: validate,
			}
			flagURL, err = prompt.Run()
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
		}
	}

	if mode == "update" && flagMethod == "" {
		// pass
	} else {
		// check if Method is one from a set, empty not OK
		match, err = regexp.MatchString(validMethodPattern, flagMethod)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		} else if match == false {
			prompt := promptui.Select{
				Label: "HTTP method",
				Items: []string{"GET", "HEAD", "POST", "PUT", "DELETE"},
			}
			_, flagMethod, err = prompt.Run()
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
		}
	}

	if mode == "update" && flagInterval == 0 {
		// pass
	} else {
		// check if Interval is in supported range
		if flagInterval < supportedIntervalMinimum || flagInterval > supportedIntervalMaximum {
			validate := func(input string) error {
				var inputInt, _ = strconv.Atoi(input)
				if inputInt < supportedIntervalMinimum || inputInt > supportedIntervalMaximum {
					return errors.New("Interval must be a value between " + strconv.Itoa(supportedIntervalMinimum) + " and " + strconv.Itoa(supportedIntervalMaximum))
				}
				return nil
			}
			prompt := promptui.Prompt{
				Label:    "Check interval in seconds",
				Validate: validate,
			}
			flagIntervalString, err := prompt.Run()
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
			flagInterval, _ = strconv.Atoi(flagIntervalString)
		}
	}

	if mode == "update" && flagTarget == 0 {
		// pass
	} else {
		// check if Target is in supported range
		if flagTarget < supportedTargetMinimum || flagTarget > supportedTargetMaximum {
			validate := func(input string) error {
				var inputFloat, _ = strconv.ParseFloat(input, 64)
				if inputFloat < supportedTargetMinimum || inputFloat > supportedTargetMaximum {
					return errors.New("Target must be a value between " + fmt.Sprintf("%.3f", supportedTargetMinimum) + " and " + fmt.Sprintf("%.3f", supportedTargetMaximum))
				}
				return nil
			}
			prompt := promptui.Prompt{
				Label:    "Response time in seconds",
				Validate: validate,
			}
			flagTargetString, err := prompt.Run()
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
			flagTarget, _ = strconv.ParseFloat(flagTargetString, 64)
		}
	}

	// @todo check if Regions are one or more from a list of values, empty not allowed

	// @todo check if UpCodes matches format, empty not allowed

	if mode == "update" && flagUpConfirmationsThreshold == 0 {
		// pass
	} else {
		// check UpConfirmationsThreshold is in supported range
		if flagUpConfirmationsThreshold < supportedConfirmationsThresholdMinimum || flagUpConfirmationsThreshold > supportedConfirmationsThresholdMaximum {
			validate := func(input string) error {
				var inputInt, _ = strconv.Atoi(input)
				if inputInt < supportedConfirmationsThresholdMinimum || inputInt > supportedConfirmationsThresholdMaximum {
					return errors.New("Up Confirmations Threshold must be a value between " + strconv.Itoa(supportedConfirmationsThresholdMinimum) + " and " + strconv.Itoa(supportedConfirmationsThresholdMaximum))
				}
				return nil
			}
			prompt := promptui.Prompt{
				Label:    "How many subsequent Up responses before triggering notifications",
				Validate: validate,
			}
			flagUpConfirmationsThresholdString, err := prompt.Run()
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
			flagUpConfirmationsThreshold, _ = strconv.Atoi(flagUpConfirmationsThresholdString)
		}
	}

	if mode == "update" && flagDownConfirmationsThreshold == 0 {
		// pass
	} else {
		// check DownConfirmationsThreshold is in supported range
		if flagDownConfirmationsThreshold < supportedConfirmationsThresholdMinimum || flagDownConfirmationsThreshold > supportedConfirmationsThresholdMaximum {
			validate := func(input string) error {
				var inputInt, _ = strconv.Atoi(input)
				if inputInt < supportedConfirmationsThresholdMinimum || inputInt > supportedConfirmationsThresholdMaximum {
					return errors.New("Down Confirmations Threshold must be a value between " + strconv.Itoa(supportedConfirmationsThresholdMinimum) + " and " + strconv.Itoa(supportedConfirmationsThresholdMaximum))
				}
				return nil
			}
			prompt := promptui.Prompt{
				Label:    "How many subsequent Down responses before triggering notifications",
				Validate: validate,
			}
			flagDownConfirmationsThresholdString, err := prompt.Run()
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
			flagDownConfirmationsThreshold, _ = strconv.Atoi(flagDownConfirmationsThresholdString)
		}
	}

	// @todo check if Channels are one or more from a list of values, empty allowed

	// fmt.Println(flagURL+" ("+flagName+") "+flagMethod, flagInterval, flagTarget, flagUpConfirmationsThreshold, flagDownConfirmationsThreshold)

	// all clear, we can call the API and confirm adding new check!
	check := Check{
		Name:                       flagName,
		URL:                        flagURL,
		Method:                     flagMethod,
		Interval:                   flagInterval,
		Target:                     flagTarget,
		Regions:                    flagRegions,
		UpCodes:                    flagUpCodes,
		UpConfirmationsThreshold:   flagUpConfirmationsThreshold,
		DownConfirmationsThreshold: flagDownConfirmationsThreshold,
	}
	postData, err := json.Marshal(check)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	var reqURL, reqMethod string
	if mode == "add" {
		reqURL = "/checks"
		reqMethod = http.MethodPost
	}
	if mode == "update" {
		reqURL = "/checks/" + strconv.Itoa(checkID)
		reqMethod = http.MethodPut
	}
	respData, err := util.BinocsAPI(reqURL, reqMethod, postData)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	err = json.Unmarshal(respData, &check)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	if check.ID > 0 {
		var checkIdent string
		if len(check.Name) > 0 {
			checkIdent = check.Name + " (" + check.URL + ")"
		} else {
			checkIdent = check.URL
		}
		tpl = checkIdent + ` updated successfully
`
	} else {
		fmt.Println("Error updating check")
		os.Exit(1)
	}
	fmt.Print(tpl)
}

var checkAddCmd = &cobra.Command{
	Use: "add",

	Run: func(cmd *cobra.Command, args []string) {
		checkAddOrUpdate("add", 0)
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
		if len(args) == 0 {
			fmt.Println("RTFM")
			os.Exit(1)
		} else if checkID, err := strconv.Atoi(args[0]); err == nil {
			checkAddOrUpdate("update", checkID)
		} else {
			fmt.Println("only reference by id allowed atm")
			os.Exit(1)
		}

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
