package cmd

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	util "github.com/automato-io/binocs-cli/util"
	"github.com/manifoldco/promptui"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
)

// Check comes from the API as a JSON, or from user input as `check add` flags
type Check struct {
	ID                         int      `json:"id,omitempty"`
	Ident                      string   `json:"ident,omitempty"`
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

// Identity method returns "Name (URL)" or "URL"
func (c Check) Identity() string {
	if len(c.Name) > 0 {
		return c.Name + " (" + c.URL + ")"
	}
	return c.Name
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

// ResponseCodesResponse comes from the API as a JSON
type ResponseCodesResponse struct {
	Xx2  int    `json:"2xx"`
	Xx3  int    `json:"3xx"`
	Xx4  int    `json:"4xx"`
	Xx5  int    `json:"5xx"`
	From string `json:"from"`
	To   string `json:"to"`
}

// ResponseTimeHeatmapResponse comes from the API as a JSON
type ResponseTimeHeatmapResponse struct {
	Rt0  int    `json:"rt0"`
	Rt1  int    `json:"rt1"`
	Rt2  int    `json:"rt2"`
	Rt3  int    `json:"rt3"`
	Rt4  int    `json:"rt4"`
	Rt5  int    `json:"rt5"`
	Rt6  int    `json:"rt6"`
	Rt7  int    `json:"rt7"`
	From string `json:"from"`
	To   string `json:"to"`
}

// RegionsResponse comes from the API as a JSON
type RegionsResponse struct {
	Regions []string `json:"regions"`
}

// `check` flags
var (
	checkFlagPeriod string
	checkFlagRegion string
	checkFlagStatus string
)

// `check ls` flags
var (
	checkListFlagPeriod string
	checkListFlagRegion string
	checkListFlagStatus string
)

// `check inspect` flags
var (
	checkInspectFlagPeriod string
	checkInspectFlagRegion string
)

// `check add` flags
var (
	checkAddFlagName                       string
	checkAddFlagURL                        string
	checkAddFlagMethod                     string
	checkAddFlagInterval                   int
	checkAddFlagTarget                     float64
	checkAddFlagRegions                    []string
	checkAddFlagUpCodes                    string
	checkAddFlagUpConfirmationsThreshold   int
	checkAddFlagDownConfirmationsThreshold int
)

// `check update` flags
var (
	checkUpdateFlagName                       string
	checkUpdateFlagURL                        string
	checkUpdateFlagMethod                     string
	checkUpdateFlagInterval                   int
	checkUpdateFlagTarget                     float64
	checkUpdateFlagRegions                    []string
	checkUpdateFlagUpCodes                    string
	checkUpdateFlagUpConfirmationsThreshold   int
	checkUpdateFlagDownConfirmationsThreshold int
)

const (
	supportedIntervalMinimum               = 5
	supportedIntervalMaximum               = 900
	supportedTargetMinimum                 = 0.01
	supportedTargetMaximum                 = 10.0
	validNamePattern                       = `^[a-zA-Z0-9_\ \/\-\.]{0,25}$`
	validMethodPattern                     = `^(GET|HEAD|POST|PUT|DELETE)$` // hardcoded; reflects supportedHTTPMethods
	validUpCodePattern                     = `^([,]?([1-5]{1}[0-9]{2}-[1-5]{1}[0-9]{2}|([1-5]{1}(([0-9]{2}|[0-9]{1}x)|xx))))+$`
	validURLPattern                        = `^https?:\/\/(www\.)?[-a-zA-Z0-9@:%._\+~#=]{2,256}\.[a-z]{2,4}\b([-a-zA-Z0-9@:%_\+.~#?&//=]*)$`
	validRegionPattern                     = `^[a-z0-9\-]{8,30}$`
	validPeriodPattern                     = `^hour|day|week|month$`
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

var aggregateMetricsDataPoints = map[string]int{
	periodHour:  60,
	periodDay:   96,
	periodWeek:  84,
	periodMonth: 120,
}

var supportedRegions = []string{}

func init() {
	rootCmd.AddCommand(checkCmd)
	checkCmd.AddCommand(checkAddCmd)
	checkCmd.AddCommand(checkInspectCmd)
	checkCmd.AddCommand(checkListCmd)
	checkCmd.AddCommand(checkUpdateCmd)
	checkCmd.AddCommand(checkDeleteCmd)

	checkCmd.Flags().StringVarP(&checkFlagPeriod, "period", "p", "day", "Display values and charts for specified period")
	checkCmd.Flags().StringVarP(&checkFlagRegion, "region", "r", "", "Display values and charts from the specified region only")
	checkCmd.Flags().StringVarP(&checkFlagStatus, "status", "s", "", "List only \"up\" or \"down\" checks, default \"all\"")

	checkAddCmd.Flags().StringVarP(&checkAddFlagName, "name", "n", "", "Check name")
	checkAddCmd.Flags().StringVarP(&checkAddFlagURL, "url", "u", "", "URL to check")
	checkAddCmd.Flags().StringVarP(&checkAddFlagMethod, "method", "m", "GET", "HTTP method (GET, HEAD, POST, PUT, DELETE)")
	checkAddCmd.Flags().IntVarP(&checkAddFlagInterval, "interval", "i", 60, "How often binocs checks the URL, in seconds")
	checkAddCmd.Flags().Float64VarP(&checkAddFlagTarget, "target", "t", 1.20, "Response time that accomodates Apdex=1.0, in seconds with up to 3 decimal places")
	// @todo fix help text - how to list all regions
	checkAddCmd.Flags().StringSliceVarP(&checkAddFlagRegions, "regions", "r", []string{"all"}, "From where in the world we check the provided URL. Choose `all` or any combination of `us-east-1`, `eu-central-1`, ...")
	checkAddCmd.Flags().StringVarP(&checkAddFlagUpCodes, "up_codes", "", "200-302", "What are the good (\"UP\") HTTP response codes, e.g. `2xx` or `200-302`, or `200,301`")
	checkAddCmd.Flags().IntVarP(&checkAddFlagUpConfirmationsThreshold, "up_confirmations_threshold", "", 2, "How many subsequent Up responses before triggering notifications")
	checkAddCmd.Flags().IntVarP(&checkAddFlagDownConfirmationsThreshold, "down_confirmations_threshold", "", 2, "How many subsequent Down responses before triggering notifications")
	checkAddCmd.Flags().SortFlags = false

	checkInspectCmd.Flags().StringVarP(&checkInspectFlagPeriod, "period", "p", "day", "Display values and charts for specified period")
	checkInspectCmd.Flags().StringVarP(&checkInspectFlagRegion, "region", "r", "", "Display values and charts from the specified region only")

	checkListCmd.Flags().StringVarP(&checkListFlagPeriod, "period", "p", "day", "Display MRT, UPTIME, APDEX values and APDEX chart for specified period")
	checkListCmd.Flags().StringVarP(&checkListFlagRegion, "region", "r", "", "Display MRT, UPTIME, APDEX values and APDEX chart from the specified region only")
	checkListCmd.Flags().StringVarP(&checkListFlagStatus, "status", "s", "", "List only \"up\" or \"down\" checks, default \"all\"")

	checkUpdateCmd.Flags().StringVarP(&checkUpdateFlagName, "name", "n", "", "Check name")
	checkUpdateCmd.Flags().StringVarP(&checkUpdateFlagURL, "url", "u", "", "URL to check")
	checkUpdateCmd.Flags().StringVarP(&checkUpdateFlagMethod, "method", "m", "", "HTTP method (GET, HEAD, POST, PUT, DELETE)")
	checkUpdateCmd.Flags().IntVarP(&checkUpdateFlagInterval, "interval", "i", 0, "How often we check the URL, in seconds")
	checkUpdateCmd.Flags().Float64VarP(&checkUpdateFlagTarget, "target", "t", 0, "Response time that accomodates Apdex=1.0, in seconds with up to 3 decimal places")
	// @todo fix help text - how to list all regions
	checkUpdateCmd.Flags().StringSliceVarP(&checkUpdateFlagRegions, "regions", "r", []string{}, "From where in the world we check the provided URL. Choose `all` or any combination of `us-east-1`, `eu-central-1`, ...")
	checkUpdateCmd.Flags().StringVarP(&checkUpdateFlagUpCodes, "up_codes", "", "", "What are the good (\"UP\") HTTP response codes, e.g. `2xx` or `200-302`, or `200,301`")
	checkUpdateCmd.Flags().IntVarP(&checkUpdateFlagUpConfirmationsThreshold, "up_confirmations_threshold", "", 0, "How many subsequent Up responses before triggering notifications")
	checkUpdateCmd.Flags().IntVarP(&checkUpdateFlagDownConfirmationsThreshold, "down_confirmations_threshold", "", 0, "How many subsequent Down responses before triggering notifications")
	checkUpdateCmd.Flags().SortFlags = false
}

var checkCmd = &cobra.Command{
	Use:   "check",
	Short: "Manage your checks",
	Long: `
Manage your checks. A command (one of "add", "delete", "inspect", "list" or "update") is optional.

If neither command nor argument are provided, assume "binocs checks list".
	
If an argument is provided without any command, assume "binocs checks inspect <arg>".
`,
	Aliases: []string{"checks"},
	Example: "",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			checkListFlagPeriod = checkFlagPeriod
			checkListFlagRegion = checkFlagRegion
			checkListFlagStatus = checkFlagStatus
			cmd.Run(checkListCmd, args)
		} else if len(args) == 1 && true { // @todo true ~ check id validity regex
			checkInspectFlagPeriod = checkFlagPeriod
			checkInspectFlagRegion = checkFlagRegion
			cmd.Run(checkInspectCmd, args)
		} else {
			fmt.Println("Unsupported command/arguments combination, please see --help")
			os.Exit(1)
		}
	},
}

var checkAddCmd = &cobra.Command{
	Use:  "add",
	Args: cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		checkAddOrUpdate("add", "")
	},
}

var checkInspectCmd = &cobra.Command{
	Use:     "inspect",
	Aliases: []string{"view", "show", "info"},
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return fmt.Errorf("please provide the identifier of the check you wish to inspect")
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		urlValues := url.Values{
			"period": []string{"day"},
		}
		periodTableTitle := "1 DAY"

		match, err := regexp.MatchString(validPeriodPattern, checkInspectFlagPeriod)
		if err == nil && match == true {
			urlValues.Set("period", checkInspectFlagPeriod)
			switch checkInspectFlagPeriod {
			case "hour":
				periodTableTitle = "1 HOUR"
			case "day":
				periodTableTitle = "1 DAY"
			case "week":
				periodTableTitle = "1 WEEK"
			case "month":
				periodTableTitle = "1 MONTH"
			}
		}

		spin.Start()
		spin.Suffix = " loading check " + args[0]
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

		spin.Suffix = " loading check metrics for " + args[0]
		metrics, err := fetchMetrics(respJSON.Ident, urlValues)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		// Table "main"

		tableMainCheckCellContent := `Name: ` + respJSON.Name + `
URL: ` + respJSON.URL + `
Method: ` + respJSON.Method + `
HTTP Status Code: ` + respJSON.LastStatusCode + `
` + statusName[respJSON.LastStatus] + " for " + outputDurationWithDays(respJSON.LastStatusDuration)

		tableMainMetricsCellContent := `Uptime: ` + formatUptime(metrics.Uptime) + `
Apdex: ` + formatApdex(metrics.Apdex) + `
Mean Response Time: ` + formatMRT(metrics.MRT)

		tableMainSettingsCellContent := `Checking interval: ` + strconv.Itoa(respJSON.Interval) + ` s 
Target response time: ` + fmt.Sprintf("%.3f", respJSON.Target) + ` s
UP HTTP Codes: ` + respJSON.UpCodes + `
Confirmations thresholds: UP: ` + strconv.Itoa(respJSON.UpConfirmationsThreshold) + `, DOWN: ` + strconv.Itoa(respJSON.DownConfirmationsThreshold) + ` 
Binocs locations: ` + strings.Join(respJSON.Regions, ", ")

		tableMain := tablewriter.NewWriter(os.Stdout)
		tableMain.SetHeader([]string{"CHECK", "METRICS (" + periodTableTitle + ")", "SETTINGS"})
		tableMain.SetAutoWrapText(false)
		tableMain.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
		tableMain.SetColumnAlignment([]int{tablewriter.ALIGN_DEFAULT, tablewriter.ALIGN_DEFAULT, tablewriter.ALIGN_DEFAULT})
		tableMain.Append([]string{tableMainCheckCellContent, tableMainMetricsCellContent, tableMainSettingsCellContent})

		// Combined table

		tableCharts := tablewriter.NewWriter(os.Stdout)
		tableCharts.SetAutoWrapText(false)
		tableCharts.SetRowLine(true)
		tableCharts.SetColumnAlignment([]int{tablewriter.ALIGN_LEFT})

		// Sub-table "http response codes"

		responseCodesData, err := util.BinocsAPI("/checks/"+respJSON.Ident+"/response-codes?"+urlValues.Encode(), http.MethodGet, []byte{})
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		responseCodes := make([]ResponseCodesResponse, 0)
		decoder := json.NewDecoder(bytes.NewBuffer(responseCodesData))
		err = decoder.Decode(&responseCodes)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		responseCodesChart := drawResponseCodesChart(responseCodes, aggregateMetricsDataPoints[urlValues.Get("period")], "            ")
		responseCodesChartTitle := drawChartTitle("HTTP RESPONSE CODES", responseCodesChart, periodTableTitle)
		tableCharts.Append([]string{responseCodesChartTitle})
		tableCharts.Append([]string{responseCodesChart})

		// Sub-table "apdex trend"

		apdexData, err := util.BinocsAPI("/checks/"+respJSON.Ident+"/apdex?"+urlValues.Encode(), http.MethodGet, []byte{})
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		apdex := make([]ApdexResponse, 0)
		decoder = json.NewDecoder(bytes.NewBuffer(apdexData))
		err = decoder.Decode(&apdex)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		apdexChart := drawApdexChart(apdex, aggregateMetricsDataPoints[urlValues.Get("period")], "      ")
		apdexChartTitle := drawChartTitle("APDEX TREND", apdexChart, periodTableTitle)
		tableCharts.Append([]string{apdexChartTitle})
		tableCharts.Append([]string{apdexChart})

		// Sub-table "response times heatmap"

		responseTimeHeatmapData, err := util.BinocsAPI("/checks/"+respJSON.Ident+"/response-time-heatmap?"+urlValues.Encode(), http.MethodGet, []byte{})
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		responseTimeHeatmap := make([]ResponseTimeHeatmapResponse, 0)
		decoder = json.NewDecoder(bytes.NewBuffer(responseTimeHeatmapData))
		err = decoder.Decode(&responseTimeHeatmap)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		responseTimeHeatmapChart := drawResponseTimeHeatmapChart(responseTimeHeatmap, aggregateMetricsDataPoints[urlValues.Get("period")], "")
		responseTimeHeatmapChartTitle := drawChartTitle("RESPONSE TIME HEATMAP", responseTimeHeatmapChart, periodTableTitle)
		tableCharts.Append([]string{responseTimeHeatmapChartTitle})
		tableCharts.Append([]string{responseTimeHeatmapChart})

		// Timeline

		timeline := drawTimeline(urlValues.Get("period"), aggregateMetricsDataPoints[urlValues.Get("period")], "                ")
		tableCharts.Append([]string{timeline})

		spin.Stop()
		tableMain.Render()
		fmt.Println()
		tableCharts.Render()
	},
}

// @todo allow specifying -interval 24h|3d default 24h for mrt, uptime, apdex and apdex chart
var checkListCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Args:    cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		urlValues1 := url.Values{}
		urlValues2 := url.Values{
			"period": []string{"day"},
		}
		apdexPeriodTableTitle := "1 DAY"

		match, err := regexp.MatchString(validPeriodPattern, checkListFlagPeriod)
		if err == nil && match == true {
			urlValues2.Set("period", checkListFlagPeriod)
			switch checkListFlagPeriod {
			case "hour":
				apdexPeriodTableTitle = "1 HOUR"
			case "day":
				apdexPeriodTableTitle = "1 DAY"
			case "week":
				apdexPeriodTableTitle = "1 WEEK"
			case "month":
				apdexPeriodTableTitle = "1 MONTH"
			}
		}

		// @todo check against currently supported GET /regions
		match, err = regexp.MatchString(validRegionPattern, checkListFlagRegion)
		if len(checkListFlagRegion) > 0 && match == false {
			fmt.Println("Invalid region provided")
			os.Exit(1)
		} else if err == nil && match == true {
			urlValues2.Set("region", checkListFlagRegion)
		}

		checkListFlagStatus = strings.ToUpper(checkListFlagStatus)
		if checkListFlagStatus == statusNameUp || checkListFlagStatus == statusNameDown {
			urlValues1.Set("status", checkListFlagStatus)
		}

		spin.Start()
		spin.Suffix = " loading checks..."

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
		for _, v := range respJSON {
			spin.Suffix = " loading metrics for " + v.Identity()
			metrics, err := fetchMetrics(v.Ident, urlValues2)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}

			apdexData, err := util.BinocsAPI("/checks/"+v.Ident+"/apdex?"+urlValues2.Encode(), http.MethodGet, []byte{})
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

			apdexChart := drawCompactApdexChart(apdex)

			tableValueMRT := formatMRT(metrics.MRT)
			tableValueUptime := formatUptime(metrics.Uptime)
			tableValueApdex := formatApdex(metrics.Apdex)
			if metrics.Apdex == "" {
				apdexChart = ""
			}
			tableRow := []string{
				v.Ident, v.Name, v.URL, v.Method, statusName[v.LastStatus] + " " + outputDurationWithDays(v.LastStatusDuration), v.LastStatusCode, strconv.Itoa(v.Interval) + " s", fmt.Sprintf("%.3f s", v.Target), tableValueMRT, tableValueUptime, tableValueApdex, apdexChart,
			}
			tableData = append(tableData, tableRow)
		}
		table := tablewriter.NewWriter(os.Stdout)
		table.SetHeader([]string{"ID", "NAME", "URL", "METHOD", "STATUS", "HTTP CODE", "INTERVAL", "TARGET", "MRT", "UPTIME", "APDEX", "APDEX " + apdexPeriodTableTitle})
		table.SetColumnAlignment([]int{tablewriter.ALIGN_DEFAULT, tablewriter.ALIGN_DEFAULT, tablewriter.ALIGN_DEFAULT, tablewriter.ALIGN_DEFAULT, tablewriter.ALIGN_DEFAULT, tablewriter.ALIGN_DEFAULT,
			tablewriter.ALIGN_RIGHT, tablewriter.ALIGN_RIGHT, tablewriter.ALIGN_RIGHT, tablewriter.ALIGN_RIGHT, tablewriter.ALIGN_RIGHT, tablewriter.ALIGN_RIGHT,
		})
		for _, v := range tableData {
			table.Append(v)
		}
		spin.Stop()
		table.Render()
	},
}

var checkUpdateCmd = &cobra.Command{
	Use: "update",
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return fmt.Errorf("please provide the identifier of the check you wish to update")
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		checkAddOrUpdate("update", args[0])
	},
}

var checkDeleteCmd = &cobra.Command{
	Use:     "delete",
	Aliases: []string{"del", "rm"},
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return fmt.Errorf("please provide the identifier of the check you wish to delete")
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		respData, err := util.BinocsAPI("/checks/"+args[0], http.MethodGet, []byte{})
		if err != nil {
			// @todo verbose error
			fmt.Println(err)
			os.Exit(1)
		}
		var respJSON Check
		err = json.Unmarshal(respData, &respJSON)
		if err != nil {
			// @todo verbose error
			fmt.Println(err)
			os.Exit(1)
		}

		prompt := promptui.Prompt{
			Label:     "Delete " + respJSON.Name + " (" + respJSON.URL + ")?",
			IsConfirm: true,
		}
		_, err = prompt.Run()
		if err != nil {
			fmt.Println("Aborting")
			os.Exit(0)
		}
		_, err = util.BinocsAPI("/checks/"+args[0], http.MethodDelete, []byte{})
		if err != nil {
			// @todo verbose error
			fmt.Println(err)
			os.Exit(1)
		}
		tpl := `Check successfully deleted
`
		fmt.Print(tpl)
	},
}

func fetchMetrics(ident string, urlValues url.Values) (MetricsResponse, error) {
	var metrics MetricsResponse
	metricsData, err := util.BinocsAPI("/checks/"+ident+"/metrics?"+urlValues.Encode(), http.MethodGet, []byte{})
	if err != nil {
		return metrics, err
	}
	err = json.Unmarshal(metricsData, &metrics)
	if err != nil {
		return metrics, err
	}
	if metrics.Uptime == "100.00" {
		metrics.Uptime = "100"
	}
	return metrics, nil
}

func formatMRT(mrt string) string {
	if mrt == "" {
		return "n/a"
	}
	return mrt + " s"
}

func formatUptime(uptime string) string {
	if uptime == "" {
		return "n/a"
	}
	return fmt.Sprintf("%v %%", uptime)
}

func formatApdex(apdex string) string {
	if apdex == "" {
		return "n/a"
	}
	return apdex
}

func loadSupportedRegions() {
	respData, err := util.BinocsAPI("/regions", http.MethodGet, []byte{})
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	// @toto verbose print respData
	regionsResponse := RegionsResponse{}
	err = json.Unmarshal(respData, &regionsResponse)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	supportedRegions = regionsResponse.Regions
}

func isSupportedRegion(region string) bool {
	for _, r := range supportedRegions {
		if r == region {
			return true
		}
	}
	return false
}

func outputDurationWithDays(d string) string {
	parsed, err := time.ParseDuration(d)
	if err != nil {
		return d
	}
	if parsed.Hours() > 48 {
		days := math.Floor(parsed.Hours() / 24)
		hours := math.Floor(parsed.Hours() - days*24)
		re1 := regexp.MustCompile(`([0-9]+)h`)
		rest := re1.ReplaceAllString(d, fmt.Sprintf("%.0f", hours)+"h")
		re2 := regexp.MustCompile(`([0-9]+)s`)
		rest = re2.ReplaceAllString(rest, "")
		return fmt.Sprintf("%.0fd %s", days, rest)
	}
	return d
}

func drawCompactApdexChart(apdex []ApdexResponse) string {
	var chart string
	var alphabet = map[string]string{
		"00": " ",
		"01": "⢀",
		"02": "⢠",
		"03": "⢰",
		"04": "⢸",
		"10": "⡀",
		"11": "⣀",
		"12": "⣠",
		"13": "⣰",
		"14": "⣸",
		"20": "⡄",
		"21": "⣄",
		"22": "⣤",
		"23": "⣴",
		"24": "⣼",
		"30": "⡆",
		"31": "⣆",
		"32": "⣦",
		"33": "⣶",
		"34": "⣾",
		"40": "⡇",
		"41": "⣇",
		"42": "⣧",
		"43": "⣷",
		"44": "⣿",
	}
	var reverseApdex []ApdexResponse
	for _, v := range apdex {
		reverseApdex = append([]ApdexResponse{v}, reverseApdex...)
	}
	var assignChar = func(left, right float64) string {
		const steps = 5
		var leftDots, rightDots string
		for j := 1; j < 1+steps; j++ {
			if left <= float64(j)/steps {
				leftDots = strconv.Itoa(j - 1)
				break
			}
		}
		for k := 1; k < 1+steps; k++ {
			if right <= float64(k)/steps {
				rightDots = strconv.Itoa(k - 1)
				break
			}
		}
		return alphabet[rightDots+leftDots]
	}
	for i, v := range reverseApdex {
		if i%2 == 1 { // even
			left, _ := strconv.ParseFloat(reverseApdex[i-1].Apdex, 32)
			right, _ := strconv.ParseFloat(v.Apdex, 32)
			chart = chart + assignChar(left, right)
		} else if len(reverseApdex) == i+1 { // last
			left, _ := strconv.ParseFloat(v.Apdex, 32)
			chart = chart + assignChar(left, 0.0)
		}
	}
	chart = reverse(chart)
	return chart
}

func getApdexChartRowRange(i, numRows int) string {
	var up, down float64
	up = (float64(i) + 1.0) / float64(numRows)
	down = float64(i) / float64(numRows)
	return fmt.Sprintf("%.1f - %.1f", down, up)
}

func drawApdexChart(apdex []ApdexResponse, dataPoints int, leftMargin string) string {
	const numRows = 5
	var rows [numRows]string
	var chart string
	for _, v := range apdex {
		var vf, _ = strconv.ParseFloat(v.Apdex, 32)
		for i := 0; i < numRows; i++ {
			if vf > (float64(i)+1.0)/float64(numRows) {
				rows[i] = rows[i] + "░"
			} else if vf <= (float64(i)+1.0)/float64(numRows) && vf > float64(i)/float64(numRows) {
				rows[i] = rows[i] + "●"
			} else {
				rows[i] = rows[i] + " "
			}
		}
	}
	if len(apdex) < dataPoints {
		for i := range rows {
			rows[i] = strings.Repeat(" ", dataPoints-len(apdex)) + rows[i]
		}
	}
	for i := 0; i < numRows; i++ {
		chart = leftMargin + getApdexChartRowRange(i, numRows) + " " + rows[i] + "\n" + chart
	}
	chart = strings.TrimSuffix(chart, "\n")
	return chart
}

func drawResponseCodesChart(responseCodes []ResponseCodesResponse, dataPoints int, leftMargin string) string {
	var rows [4]string
	var chart string
	for _, v := range responseCodes {
		if v.Xx2 > 0 {
			rows[0] = rows[0] + "●"
		} else {
			rows[0] = rows[0] + " "
		}
		if v.Xx3 > 0 {
			rows[1] = rows[1] + "●"
		} else {
			rows[1] = rows[1] + " "
		}
		if v.Xx4 > 0 {
			rows[2] = rows[2] + "●"
		} else {
			rows[2] = rows[2] + " "
		}
		if v.Xx5 > 0 {
			rows[3] = rows[3] + "●"
		} else {
			rows[3] = rows[3] + " "
		}
	}
	if len(responseCodes) < dataPoints {
		for i := range rows {
			rows[i] = strings.Repeat(" ", dataPoints-len(responseCodes)) + rows[i]
		}
	}
	for i := 0; i < 4; i++ {
		chart = chart + leftMargin + strconv.Itoa(i+2) + "xx" + " " + rows[i] + "\n"
	}
	chart = strings.TrimSuffix(chart, "\n")
	return chart
}

func drawResponseTimeHeatmapChart(responseTimeHeatmap []ResponseTimeHeatmapResponse, dataPoints int, leftMargin string) string {
	var rowTitles = [8]string{
		"   error / 8+ s",
		"  4.00 - 8.00 s",
		"  2.00 - 4.00 s",
		"  1.00 - 2.00 s",
		" 0.500 - 1.00 s",
		"0.250 - 0.500 s",
		"0.125 - 0.250 s",
		"0.000 - 0.125 s",
	}
	var heatmapMaximum int
	for _, v := range responseTimeHeatmap {
		if v.Rt0 > heatmapMaximum {
			heatmapMaximum = v.Rt0
		}
		if v.Rt1 > heatmapMaximum {
			heatmapMaximum = v.Rt1
		}
		if v.Rt2 > heatmapMaximum {
			heatmapMaximum = v.Rt2
		}
		if v.Rt3 > heatmapMaximum {
			heatmapMaximum = v.Rt3
		}
		if v.Rt4 > heatmapMaximum {
			heatmapMaximum = v.Rt4
		}
		if v.Rt5 > heatmapMaximum {
			heatmapMaximum = v.Rt5
		}
		if v.Rt6 > heatmapMaximum {
			heatmapMaximum = v.Rt6
		}
		if v.Rt7 > heatmapMaximum {
			heatmapMaximum = v.Rt7
		}
	}
	var palette = [5]string{" ", "░", "▒", "▓", "█"}
	var paletteStep = float32(len(palette) - 1)
	var thresholds = [4]float32{
		1.0,
		float32(heatmapMaximum) / paletteStep,
		2.0 * float32(heatmapMaximum) / paletteStep,
		3.0 * float32(heatmapMaximum) / paletteStep,
	}
	var rows [8]string
	var chart string
	drawHeatmapPixel := func(row string, rt int) string {
		vfrt := float32(rt)
		if vfrt >= thresholds[3] {
			return row + palette[4]
		} else if vfrt >= thresholds[2] {
			return row + palette[3]
		} else if vfrt >= thresholds[1] {
			return row + palette[2]
		} else if vfrt >= thresholds[0] {
			return row + palette[1]
		} else {
			return row + palette[0]
		}
	}
	for _, v := range responseTimeHeatmap {
		rows[0] = drawHeatmapPixel(rows[0], v.Rt7)
		rows[1] = drawHeatmapPixel(rows[1], v.Rt6)
		rows[2] = drawHeatmapPixel(rows[2], v.Rt5)
		rows[3] = drawHeatmapPixel(rows[3], v.Rt4)
		rows[4] = drawHeatmapPixel(rows[4], v.Rt3)
		rows[5] = drawHeatmapPixel(rows[5], v.Rt2)
		rows[6] = drawHeatmapPixel(rows[6], v.Rt1)
		rows[7] = drawHeatmapPixel(rows[7], v.Rt0)
	}
	if len(responseTimeHeatmap) < dataPoints {
		for i := range rows {
			rows[i] = strings.Repeat(" ", dataPoints-len(responseTimeHeatmap)) + rows[i]
		}
	}
	for i := 0; i < len(rows); i++ {
		chart = chart + leftMargin + rowTitles[i] + " " + rows[i] + "\n"
	}
	chart = strings.TrimSuffix(chart, "\n")
	return chart
}

func drawChartTitle(title string, chart string, periodTitle string) string {
	var chartRunes = []rune(chart)
	var chartLineLen = 0
	var newline = '\n'
	for i, r := range chartRunes {
		if r == newline {
			chartLineLen = i
			break
		}
	}
	spacerLen := chartLineLen - len(title) - len(periodTitle)
	if len(title)+len(periodTitle)+1 < chartLineLen {
		title = title + strings.Repeat(" ", spacerLen) + periodTitle
	}
	return title
}

func drawTimeline(period string, dataPoints int, leftMargin string) string {
	var timeline [2]string
	var now = time.Now()
	switch period {
	case periodHour:
		for i := 0; i < 15; i++ {
			if i == 0 {
				timeline[0] = fmt.Sprintf("%02v", now.Minute())
			} else {
				now = now.Add(time.Duration(-4) * time.Minute)
				timeline[0] = fmt.Sprintf("%02v", now.Minute()) + `  ` + timeline[0]
			}
		}
	case periodDay:
		for i := 0; i < 16; i++ {
			if i == 0 {
				now = now.Truncate(time.Duration(15) * time.Minute)
				timeline[0] = fmt.Sprintf("%02v:%02v", now.Hour(), now.Minute())
			} else {
				now = now.Add(time.Duration(-90) * time.Minute)
				timeline[0] = fmt.Sprintf("%02v:%02v", now.Hour(), now.Minute()) + ` ` + timeline[0]
			}
		}
	case periodWeek:
		for i := 0; i < 14; i++ {
			if i == 0 {
				now = now.Truncate(time.Duration(2) * time.Hour)
				timeline[0] = fmt.Sprintf("%02v:%02v", now.Hour(), now.Minute())
			} else {
				now = now.Add(time.Duration(-12) * time.Hour)
				timeline[0] = fmt.Sprintf("%02v:%02v", now.Hour(), now.Minute()) + ` ` + timeline[0]
			}
			// second line
			if now.Hour() < 12 {
				var gap = 9
				if len(timeline[0]) < 12 {
					gap = gap - (12 - len(timeline[0]))
				}
				timeline[1] = fmt.Sprintf("%s", now.Format("Mon")) + strings.Repeat(" ", gap) + timeline[1]
			}
		}
	case periodMonth:
		for i := 0; i < 30; i++ {
			if i == 0 {
				timeline[0] = fmt.Sprintf("%02v.", now.Day())
			} else {
				now = now.Add(time.Duration(-24) * time.Hour)
				timeline[0] = fmt.Sprintf("%02v.", now.Day()) + ` ` + timeline[0]
			}
			// second line
			if now.Day() == 1 {
				var gap = len(timeline[0]) - len(now.Format("Jan"))
				timeline[1] = fmt.Sprintf("%s", now.Format("Jan")) + strings.Repeat(" ", gap) + timeline[1]
			}
		}
	}
	if len(timeline[0]) < dataPoints {
		timeline[0] = strings.Repeat(" ", dataPoints-len(timeline[0])) + timeline[0]
		if len(timeline[1]) > 0 {
			timeline[1] = strings.Repeat(" ", len(timeline[0])-len(timeline[1])) + timeline[1]
		}
	}
	if len(timeline[1]) > 0 {
		return leftMargin + timeline[0] + "\n" + leftMargin + timeline[1]
	}
	return leftMargin + timeline[0]
}

// mode = add|update
func checkAddOrUpdate(mode string, checkIdent string) {
	if mode != "add" && mode != "update" {
		fmt.Println("Unknown mode: " + mode)
		os.Exit(1)
	}

	var err error
	var match bool
	var tpl string

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
	)

	switch mode {
	case "add":
		flagName = checkAddFlagName
		flagURL = checkAddFlagURL
		flagMethod = checkAddFlagMethod
		flagInterval = checkAddFlagInterval
		flagTarget = checkAddFlagTarget
		flagRegions = checkAddFlagRegions
		flagUpCodes = checkAddFlagUpCodes
		flagUpConfirmationsThreshold = checkAddFlagUpConfirmationsThreshold
		flagDownConfirmationsThreshold = checkAddFlagDownConfirmationsThreshold
	case "update":
		flagName = checkUpdateFlagName
		flagURL = checkUpdateFlagURL
		flagMethod = checkUpdateFlagMethod
		flagInterval = checkUpdateFlagInterval
		flagTarget = checkUpdateFlagTarget
		flagRegions = checkUpdateFlagRegions
		flagUpCodes = checkUpdateFlagUpCodes
		flagUpConfirmationsThreshold = checkUpdateFlagUpConfirmationsThreshold
		flagDownConfirmationsThreshold = checkUpdateFlagDownConfirmationsThreshold
	}

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
				Label:    "Check name (optional)",
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
		// check if URL is url, empty not allowed
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
		// check if Method is one from a set, empty not allowed
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
				Label:    "Interval in seconds (default: 60 s, must be a value between " + strconv.Itoa(supportedIntervalMinimum) + " and " + strconv.Itoa(supportedIntervalMaximum) + ")",
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
					return errors.New("Target Response Time must be a value between " + fmt.Sprintf("%.3f", supportedTargetMinimum) + " and " + fmt.Sprintf("%.3f", supportedTargetMaximum))
				}
				return nil
			}
			prompt := promptui.Prompt{
				Label:    "Target Response Time in seconds (default: 1.20 s, must be a value between " + fmt.Sprintf("%.3f", supportedTargetMinimum) + " and " + fmt.Sprintf("%.3f", supportedTargetMaximum) + ")",
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

	// note: we cannot use the Prompt library here because of its lack of multi-select support,
	// we don't prompt, we choose the default "all" regions unless user specifies regions via the -r option; https://github.com/manifoldco/promptui/issues/72
	if mode == "update" && len(flagRegions) == 0 {
		// pass
	} else {
		loadSupportedRegions()
		if len(flagRegions) == 0 {
			flagRegions = supportedRegions
		} else {
			for _, r := range flagRegions {
				if r == "all" {
					flagRegions = supportedRegions
					break
				}
				if isSupportedRegion(r) == false {
					fmt.Println("unsupported region: " + r)
					os.Exit(1)
				}
			}
		}
	}

	if mode == "update" && flagUpCodes == "" {
		// pass
	} else {
		// check UpCodes matches format, empty not allowed
		match, err = regexp.MatchString(validUpCodePattern, flagUpCodes)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		} else if match == false {
			validate := func(input string) error {
				match, err = regexp.MatchString(validUpCodePattern, input)
				if err != nil {
					return errors.New("Invalid input")
				} else if match == false {
					return errors.New("Invalid input value")
				}
				return nil
			}
			prompt := promptui.Prompt{
				Label:    "What are the good (\"UP\") HTTP response codes, e.g. \"2xx\" or \"200-302\", or \"200,301\" (default: \"200-302\" s)",
				Validate: validate,
			}
			flagURL, err = prompt.Run()
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
		}
	}

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
				Label:    "Up Confirmations Threshold (default: 2, must be a value between " + strconv.Itoa(supportedConfirmationsThresholdMinimum) + " and " + strconv.Itoa(supportedConfirmationsThresholdMaximum) + ")",
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
	// @hack check flags
	postData, err := json.Marshal(check)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	if bytes.Equal(postData, []byte("{}")) {
		fmt.Printf("provide at least one parameter that you want to update\n")
		os.Exit(1)
	}
	var reqURL, reqMethod string
	if mode == "add" {
		reqURL = "/checks"
		reqMethod = http.MethodPost
	}
	if mode == "update" {
		reqURL = "/checks/" + checkIdent
		reqMethod = http.MethodPut
	}
	// @todo verbose print postData
	respData, err := util.BinocsAPI(reqURL, reqMethod, postData)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	// @toto verbose print respData
	err = json.Unmarshal(respData, &check)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	if check.ID > 0 {
		var checkDescription string
		if len(check.Name) > 0 {
			checkDescription = check.Name + " (" + check.URL + ")"
		} else {
			checkDescription = check.URL
		}
		if mode == "add" {
			tpl = "[" + check.Ident + "] " + checkDescription + ` added successfully
`
		}
		if mode == "update" {
			tpl = "[" + check.Ident + "] " + checkDescription + ` updated successfully
`
		}
	} else {
		if mode == "add" {
			fmt.Println("Error adding check")
			os.Exit(1)
		}
		if mode == "update" {
			fmt.Println("Error updating check")
			os.Exit(1)
		}
	}
	fmt.Print(tpl)
}

func reverse(s string) string {
	n := 0
	rune := make([]rune, len(s))
	for _, r := range s {
		rune[n] = r
		n++
	}
	rune = rune[0:n]
	for i := 0; i < n/2; i++ {
		rune[i], rune[n-1-i] = rune[n-1-i], rune[i]
	}
	return string(rune)
}
