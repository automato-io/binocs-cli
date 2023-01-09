package cmd

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/AlecAivazis/survey/v2"
	util "github.com/automato-io/binocs-cli/util"
	"github.com/automato-io/tablewriter"
	"github.com/fatih/color"
	"github.com/muesli/reflow/ansi"
	"github.com/spf13/cobra"
)

// Check comes from the API as a JSON, or from user input as `check add` flags
type Check struct {
	ID                         int      `json:"id,omitempty"`
	Ident                      string   `json:"ident,omitempty"`
	Name                       string   `json:"name"`
	Protocol                   string   `json:"protocol,omitempty"`
	Resource                   string   `json:"resource,omitempty"`
	Method                     string   `json:"method,omitempty"`
	Interval                   int      `json:"interval,omitempty"`
	Target                     float64  `json:"target,omitempty"`
	Regions                    []string `json:"regions,omitempty"`
	UpCodes                    string   `json:"up_codes,omitempty"`
	UpConfirmationsThreshold   int      `json:"up_confirmations_threshold,omitempty"`
	UpConfirmations            int      `json:"up_confirmations,omitempty"`
	DownConfirmationsThreshold int      `json:"down_confirmations_threshold,omitempty"`
	DownConfirmations          int      `json:"down_confirmations,omitempty"`
	LastChecked                string   `json:"last_checked,omitempty"`
	LastStatus                 int      `json:"last_status,omitempty"`
	LastStatusCode             string   `json:"last_status_code,omitempty"`
	LastStatusDuration         string   `json:"last_status_duration,omitempty"`
	Created                    string   `json:"created,omitempty"`
	Updated                    string   `json:"updated,omitempty"`
	Channels                   []string `json:"channels,omitempty"`
}

// Identity method returns formatted Name + Resource
func (c Check) Identity() string {
	if len(c.Name) > 0 {
		return c.Name + " (" + c.Resource + ")"
	}
	return c.Resource
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
	Xx1  int    `json:"1xx"`
	Xx2  int    `json:"2xx"`
	Xx3  int    `json:"3xx"`
	Xx4  int    `json:"4xx"`
	Xx5  int    `json:"5xx"`
	Err  int    `json:"Err"`
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
	checkListFlagWatch  bool
)

// `check inspect` flags
var (
	checkInspectFlagPeriod string
	checkInspectFlagRegion string
	checkInspectFlagWatch  bool
)

// `check add` flags
var (
	checkAddFlagName                       string
	checkAddFlagProtocol                   string
	checkAddFlagResource                   string
	checkAddFlagMethod                     string
	checkAddFlagInterval                   int
	checkAddFlagTarget                     float64
	checkAddFlagRegions                    []string
	checkAddFlagUpCodes                    string
	checkAddFlagUpConfirmationsThreshold   int
	checkAddFlagDownConfirmationsThreshold int
	checkAddFlagAttach                     []string
)

// `check update` flags
var (
	checkUpdateFlagName                       string
	checkUpdateFlagMethod                     string
	checkUpdateFlagInterval                   int
	checkUpdateFlagTarget                     float64
	checkUpdateFlagRegions                    []string
	checkUpdateFlagUpCodes                    string
	checkUpdateFlagUpConfirmationsThreshold   int
	checkUpdateFlagDownConfirmationsThreshold int
	checkUpdateFlagAttach                     []string
)

const (
	supportedIntervalMinimum               = 5
	supportedIntervalMaximum               = 900
	supportedTargetMinimum                 = 0.01
	supportedTargetMaximum                 = 10.0
	validNamePattern                       = `^[\p{L}\p{N}_\s\/\-\.\(\)]{0,25}$`
	validProtocolPattern                   = `^(?i)(` + protocolHTTP + `|` + protocolHTTPS + `|` + protocolTCP + `)$`
	validMethodPattern                     = `^(GET|HEAD|POST|PUT|DELETE)$` // hardcoded; reflects supportedHTTPMethods
	validUpCodePattern                     = `^([1-5]{1}[0-9]{2}-[1-5]{1}[0-9]{2}|([1-5]{1}(([0-9]{2}|[0-9]{1}x)|xx))){1}(,([1-5]{1}[0-9]{2}-[1-5]{1}[0-9]{2}|([1-5]{1}(([0-9]{2}|[0-9]{1}x)|xx))))*$`
	validRegionPattern                     = `^[a-z0-9\-]{8,30}$`
	validPeriodPattern                     = `^hour|day|week|month$`
	validChecksIdentListPattern            = `^(all|([a-f0-9]{7})(,[a-f0-9]{7})*)$`
	supportedConfirmationsThresholdMinimum = 1
	supportedConfirmationsThresholdMaximum = 10

	maxURLRuneCount          = 2083
	minURLRuneCount          = 3
	validIPPattern           = `(([0-9a-fA-F]{1,4}:){7,7}[0-9a-fA-F]{1,4}|([0-9a-fA-F]{1,4}:){1,7}:|([0-9a-fA-F]{1,4}:){1,6}:[0-9a-fA-F]{1,4}|([0-9a-fA-F]{1,4}:){1,5}(:[0-9a-fA-F]{1,4}){1,2}|([0-9a-fA-F]{1,4}:){1,4}(:[0-9a-fA-F]{1,4}){1,3}|([0-9a-fA-F]{1,4}:){1,3}(:[0-9a-fA-F]{1,4}){1,4}|([0-9a-fA-F]{1,4}:){1,2}(:[0-9a-fA-F]{1,4}){1,5}|[0-9a-fA-F]{1,4}:((:[0-9a-fA-F]{1,4}){1,6})|:((:[0-9a-fA-F]{1,4}){1,7}|:)|fe80:(:[0-9a-fA-F]{0,4}){0,4}%[0-9a-zA-Z]{1,}|::(ffff(:0{1,4}){0,1}:){0,1}((25[0-5]|(2[0-4]|1{0,1}[0-9]){0,1}[0-9])\.){3,3}(25[0-5]|(2[0-4]|1{0,1}[0-9]){0,1}[0-9])|([0-9a-fA-F]{1,4}:){1,4}:((25[0-5]|(2[0-4]|1{0,1}[0-9]){0,1}[0-9])\.){3,3}(25[0-5]|(2[0-4]|1{0,1}[0-9]){0,1}[0-9]))`
	validURLUsernamePattern  = `(\S+(:\S*)?@)`
	validURLPathPattern      = `((\/|\?|#)[^\s]*)`
	validURLPortPattern      = `(:(\d{1,5}))`
	validURLIPPattern        = `([1-9]\d?|1\d\d|2[01]\d|22[0-3]|24\d|25[0-5])(\.(\d{1,2}|1\d\d|2[0-4]\d|25[0-5])){2}(?:\.([0-9]\d?|1\d\d|2[0-4]\d|25[0-5]))`
	validURLSubdomainPattern = `((www\.)|([a-zA-Z0-9]+([-_\.]?[a-zA-Z0-9])*[a-zA-Z0-9]\.[a-zA-Z0-9]+))`
	validDNSNamePattern      = `^([a-zA-Z0-9_]{1}[a-zA-Z0-9_-]{0,62}){1}(\.[a-zA-Z0-9_]{1}[a-zA-Z0-9_-]{0,62})*[\._]?$`

	validHTTPResourcePattern  = `^((http):\/\/)` + validURLUsernamePattern + `?` + `((` + validURLIPPattern + `|(\[` + validIPPattern + `\])|(([a-zA-Z0-9]([a-zA-Z0-9-_]+)?[a-zA-Z0-9]([-\.][a-zA-Z0-9]+)*)|(` + validURLSubdomainPattern + `?))?(([a-zA-Z\x{00a1}-\x{ffff}0-9]+-?-?)*[a-zA-Z\x{00a1}-\x{ffff}0-9]+)(?:\.([a-zA-Z\x{00a1}-\x{ffff}]{1,}))?))\.?` + validURLPortPattern + `?` + validURLPathPattern + `?$`
	validHTTPSResourcePattern = `^((https):\/\/)` + validURLUsernamePattern + `?` + `((` + validURLIPPattern + `|(\[` + validIPPattern + `\])|(([a-zA-Z0-9]([a-zA-Z0-9-_]+)?[a-zA-Z0-9]([-\.][a-zA-Z0-9]+)*)|(` + validURLSubdomainPattern + `?))?(([a-zA-Z\x{00a1}-\x{ffff}0-9]+-?-?)*[a-zA-Z\x{00a1}-\x{ffff}0-9]+)(?:\.([a-zA-Z\x{00a1}-\x{ffff}]{1,}))?))\.?` + validURLPortPattern + `?` + validURLPathPattern + `?$`
)

// var supportedHTTPMethods = map[string]bool{
// 	http.MethodGet:     true,
// 	http.MethodHead:    true,
// 	http.MethodPost:    true,
// 	http.MethodPut:     true,
// 	http.MethodPatch:   false,
// 	http.MethodDelete:  true,
// 	http.MethodConnect: false,
// 	http.MethodOptions: false,
// 	http.MethodTrace:   false,
// }

var aggregateMetricsDataPoints = map[string]int{
	periodHour:  60,
	periodDay:   96,
	periodWeek:  84,
	periodMonth: 120,
}

func init() {
	loadSupportedRegions()

	rootCmd.AddCommand(checksCmd)

	rootCmd.AddCommand(checkCmd)

	checkCmd.AddCommand(checkAddCmd)
	checkCmd.AddCommand(checkInspectCmd)
	checkCmd.AddCommand(checkListCmd)
	checkCmd.AddCommand(checkUpdateCmd)
	checkCmd.AddCommand(checkDeleteCmd)

	checkCmd.Flags().StringVarP(&checkFlagPeriod, "period", "p", "day", "display values and charts for specified period")
	checkCmd.Flags().StringVarP(&checkFlagRegion, "region", "r", "", "display values and charts from the specified region only")
	checkCmd.Flags().StringVarP(&checkFlagStatus, "status", "s", "", "list only \"up\" or \"down\" checks, default \"all\"")

	checkAddCmd.Flags().StringVarP(&checkAddFlagName, "name", "n", "", "check name")
	checkAddCmd.Flags().StringVarP(&checkAddFlagProtocol, "protocol", "p", "", "protocol (HTTP, HTTPS or TCP)")
	checkAddCmd.Flags().StringVarP(&checkAddFlagResource, "resource", "r", "", "resource to check, a URL in case of HTTP(S), or HOSTNAME:PORT in case of TCP")
	checkAddCmd.Flags().StringVarP(&checkAddFlagMethod, "method", "m", "", "HTTP(S) method (GET, HEAD, POST, PUT, DELETE)")
	checkAddCmd.Flags().IntVarP(&checkAddFlagInterval, "interval", "i", 60, "how often Binocs checks given resource, in seconds")
	checkAddCmd.Flags().Float64VarP(&checkAddFlagTarget, "target", "t", 1.20, "response time that accommodates Apdex=1.0, in seconds with up to 3 decimal places")
	checkAddCmd.Flags().StringSliceVar(&checkAddFlagRegions, "region", []string{}, fmt.Sprintf("from where in the world Binocs checks given resource; choose one or more from: %v", strings.Join(getSupportedRegionAliases(), ", ")))
	checkAddCmd.Flags().StringVarP(&checkAddFlagUpCodes, "up_codes", "", "200-302", "what are the good (\"up\") HTTP(S) response codes, e.g. `2xx` or `200-302`, or `200,301`")
	checkAddCmd.Flags().IntVarP(&checkAddFlagUpConfirmationsThreshold, "up_confirmations_threshold", "", 2, "how many subsequent \"up\" responses before triggering notifications")
	checkAddCmd.Flags().IntVarP(&checkAddFlagDownConfirmationsThreshold, "down_confirmations_threshold", "", 2, "how many subsequent \"down\" responses before triggering notifications")
	checkAddCmd.Flags().StringSliceVar(&checkAddFlagAttach, "attach", []string{}, "channels to attach to this check (optional); can be either \"all\", or one or more channel identifiers")
	checkAddCmd.Flags().SortFlags = false

	checkInspectCmd.Flags().StringVarP(&checkInspectFlagPeriod, "period", "p", "day", "display values and charts for specified period")
	checkInspectCmd.Flags().StringVarP(&checkInspectFlagRegion, "region", "r", "", "display values and charts from the specified region only")
	checkInspectCmd.Flags().BoolVar(&checkInspectFlagWatch, "watch", false, "run in cell view and refresh binocs output every 5 seconds")

	checksCmd.Flags().StringVarP(&checkListFlagPeriod, "period", "p", "day", "display MRT, UPTIME, APDEX values and APDEX chart for specified period")
	checksCmd.Flags().StringVarP(&checkListFlagRegion, "region", "r", "", "display MRT, UPTIME, APDEX values and APDEX chart from the specified region only")
	checksCmd.Flags().StringVarP(&checkListFlagStatus, "status", "s", "", "list only \"up\" or \"dow\" checks, default \"all\"")
	checksCmd.Flags().BoolVar(&checkListFlagWatch, "watch", false, "run in cell view and refresh binocs output every 5 seconds")
	checkListCmd.Flags().StringVarP(&checkListFlagPeriod, "period", "p", "day", "display MRT, UPTIME, APDEX values and APDEX chart for specified period")
	checkListCmd.Flags().StringVarP(&checkListFlagRegion, "region", "r", "", "display MRT, UPTIME, APDEX values and APDEX chart from the specified region only")
	checkListCmd.Flags().StringVarP(&checkListFlagStatus, "status", "s", "", "list only \"up\" or \"down\" checks, default \"all\"")
	checkListCmd.Flags().BoolVar(&checkListFlagWatch, "watch", false, "run in cell view and refresh binocs output every 5 seconds")

	checkUpdateCmd.Flags().StringVarP(&checkUpdateFlagName, "name", "n", "", "check name")
	checkUpdateCmd.Flags().StringVarP(&checkUpdateFlagMethod, "method", "m", "", "HTTP(S) method (GET, HEAD, POST, PUT, DELETE)")
	checkUpdateCmd.Flags().IntVarP(&checkUpdateFlagInterval, "interval", "i", 0, "how often Binocs checks given resource, in seconds")
	checkUpdateCmd.Flags().Float64VarP(&checkUpdateFlagTarget, "target", "t", 0, "response time that accommodates Apdex=1.0, in seconds with up to 3 decimal places")
	checkUpdateCmd.Flags().StringSliceVarP(&checkUpdateFlagRegions, "region", "r", []string{}, fmt.Sprintf("from where in the world Binocs checks given resource; choose one or more from: %v", strings.Join(getSupportedRegionAliases(), ", ")))
	checkUpdateCmd.Flags().StringVarP(&checkUpdateFlagUpCodes, "up_codes", "", "", "what are the good (\"up\") HTTP(S) response codes, e.g. `2xx` or `200-302`, or `200,301`")
	checkUpdateCmd.Flags().IntVarP(&checkUpdateFlagUpConfirmationsThreshold, "up_confirmations_threshold", "", 0, "how many subsequent \"up\" responses before triggering notifications")
	checkUpdateCmd.Flags().IntVarP(&checkUpdateFlagDownConfirmationsThreshold, "down_confirmations_threshold", "", 0, "how many subsequent \"down\" responses before triggering notifications")
	checkUpdateCmd.Flags().StringSliceVar(&checkUpdateFlagAttach, "attach", []string{}, "channels to attach to this check (optional); can be either \"all\", or one or more channel identifiers")
	checkUpdateCmd.Flags().SortFlags = false
}

func isURL(str, protocol string) bool {
	if str == "" || utf8.RuneCountInString(str) >= maxURLRuneCount || len(str) <= minURLRuneCount || strings.HasPrefix(str, ".") {
		return false
	}
	u, err := url.Parse(str)
	if err != nil {
		return false
	}
	if strings.HasPrefix(u.Host, ".") {
		return false
	}
	if u.Host == "" && (u.Path != "" && !strings.Contains(u.Path, ".")) {
		return false
	}
	if protocol == protocolHTTP {
		rxURL := regexp.MustCompile(validHTTPResourcePattern)
		return rxURL.MatchString(str)
	}
	if protocol == protocolHTTPS {
		rxURL := regexp.MustCompile(validHTTPSResourcePattern)
		return rxURL.MatchString(str)
	}
	return false
}

func isIP(str string) bool {
	return net.ParseIP(str) != nil
}

func isDNSName(str string) bool {
	if str == "" || len(strings.Replace(str, ".", "", -1)) > 255 {
		return false
	}
	rxDNSName := regexp.MustCompile(validDNSNamePattern)
	return !isIP(str) && rxDNSName.MatchString(str)
}

func isPort(str string) bool {
	if i, err := strconv.Atoi(str); err == nil && i > 0 && i < 65536 {
		return true
	}
	return false
}

func isHost(str string) bool {
	return isIP(str) || isDNSName(str)
}

func isValidHTTPResource(res string) bool {
	if strings.HasPrefix(res, "http://") {
		return isURL(res, protocolHTTP)
	} else {
		return isURL("http://"+res, protocolHTTP)
	}
}

func isValidHTTPSResource(res string) bool {
	if strings.HasPrefix(res, "https://") {
		return isURL(res, protocolHTTPS)
	} else {
		return isURL("https://"+res, protocolHTTPS)
	}
}

func isValidICMPResource(res string) bool {
	if strings.HasPrefix(res, "icmp://") {
		return isHost(res[7:])
	}
	return isHost(res)
}

func isValidTCPResource(res string) bool {
	var rc []string
	if strings.HasPrefix(res, "tcp://") {
		rc = strings.Split(res[6:], ":")
	} else {
		rc = strings.Split(res, ":")
	}
	if len(rc) != 2 {
		return false
	}
	return isHost(rc[0]) && isPort(rc[1])
}

var checkCmd = &cobra.Command{
	Use:   "check",
	Short: "Manage checks",
	Long: `
Manage HTTP and HTTPS checks.

`,
	DisableAutoGenTag: true,
}

var checksCmd = &cobra.Command{
	Use:               "checks",
	Args:              cobra.NoArgs,
	Short:             checkListCmd.Short,
	Long:              checkListCmd.Long,
	DisableAutoGenTag: true,
	Run: func(cmd *cobra.Command, args []string) {
		checkListCmd.Run(cmd, args)
	},
}

var checkAddCmd = &cobra.Command{
	Use:   "add",
	Short: "Add a new endpoint that you want to check",
	Long: `
Add a check and start reporting on it. Check identifier is returned upon successful add operation.

This command is interactive and asks user for parameters that were not provided as flags.
`,
	Aliases:           []string{"create"},
	Args:              cobra.NoArgs,
	DisableAutoGenTag: true,
	Run: func(cmd *cobra.Command, args []string) {
		util.VerifyAuthenticated()
		checkAddOrUpdate("add", "")
	},
}

var checkInspectCmd = &cobra.Command{
	Use:   "inspect",
	Short: "View check status and metrics",
	Long: `
View check status and metrics.
`,
	Aliases:           []string{"view", "show", "info"},
	Args:              cobra.ExactArgs(1),
	DisableAutoGenTag: true,
	Run: func(cmd *cobra.Command, args []string) {
		util.VerifyAuthenticated()

		if checkInspectFlagWatch {
			runAsWatch()
			return
		}

		var decoder *json.Decoder

		urlValues := url.Values{
			"period": []string{"day"},
		}
		periodTableTitle := "1 DAY"

		match, err := regexp.MatchString(validPeriodPattern, checkInspectFlagPeriod)
		if err == nil && match {
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

		match = isValidRegionAlias(checkInspectFlagRegion)
		if len(checkInspectFlagRegion) > 0 && !match {
			handleErr(fmt.Errorf("Invalid region provided. Supported regions: " + strings.Join(getSupportedRegionAliases(), ", ")))
		} else if err == nil && match {
			urlValues.Set("region", getRegionIdByAlias(checkInspectFlagRegion))
		}

		spin.Start()
		defer spin.Stop()
		spin.Suffix = colorFaint.Sprint(" loading metrics...")

		user, err := fetchUser()
		if err != nil {
			handleErr(err)
		}

		respData, err := util.BinocsAPI("/checks/"+args[0], http.MethodGet, []byte{})
		if err != nil {
			handleErr(err)
		}
		var respJSON Check
		err = json.Unmarshal(respData, &respJSON)
		if err != nil {
			handleErr(err)
		}

		metrics, err := fetchMetrics(respJSON.Ident, &urlValues)
		if err != nil {
			handleErr(err)
		}

		// Table "main"

		var resourceTitle, methodLine, responseLine, lastCheckedLine, upHTTPCodesLine, checkName, statusLine string
		if respJSON.Protocol == protocolHTTP || respJSON.Protocol == protocolHTTPS {
			resourceTitle = "URL"
			methodLine = colorBold.Sprint("Method: ") + respJSON.Method + "\n"
			if len(respJSON.LastStatusCode) > 0 {
				responseLine = colorBold.Sprint("Response: ") + respJSON.LastStatusCode + "\n"
			} else {
				responseLine = colorBold.Sprint("Response: ") + "[waiting for data]" + "\n"
			}
			upHTTPCodesLine = colorBold.Sprint("UP HTTP Codes: ") + respJSON.UpCodes + "\n"
		}
		if respJSON.Protocol == protocolICMP || respJSON.Protocol == protocolTCP {
			resourceTitle = "Host"
		}

		if respJSON.Name == "" {
			checkName = "-"
		} else {
			checkName = respJSON.Name
		}

		if respJSON.LastChecked != "nil" {
			lastChecked, err := time.Parse("2006-01-02 15:04:05 -0700", respJSON.LastChecked)
			if err == nil {
				lastCheckedLine = colorBold.Sprint("Last checked: ") + time.Since(lastChecked).Round(1*time.Second).String() + " ago"
			}
		}

		if respJSON.LastStatus == statusUnknown {
			statusLine = ""
		} else {
			statusLine = colorBold.Sprint("Status: ") + formatStatus(&respJSON) + "\n"
		}

		if user.CreditBalance == 0 {
			statusLine = colorBold.Sprint("Status: ") + color.YellowString(statusName[statusUnknown]) + "\n"
			responseLine = colorBold.Sprint("Response: ") + "n/a" + "\n"
		}

		tableMainCheckCellContent := colorBold.Sprint(`ID: `) + respJSON.Ident + "\n" +
			colorBold.Sprint(`Name: `) + checkName + "\n" +
			colorBold.Sprint(resourceTitle+`: `) + respJSON.Resource + "\n" +
			methodLine +
			statusLine +
			responseLine +
			lastCheckedLine

		uptimeValue := formatUptime(metrics.Uptime)
		apdexValue := formatApdex(metrics.Apdex)
		mrtValue := formatMRT(metrics.MRT)
		if uptimeValue == "n/a" && apdexValue == "n/a" && mrtValue == "n/a" {
			uptimeValue = "[waiting for data]"
			apdexValue = "[waiting for data]"
			mrtValue = "[waiting for data]"
		}

		if user.CreditBalance == 0 {
			uptimeValue = "n/a"
			apdexValue = "n/a"
			mrtValue = "n/a"
		}

		tableMainMetricsCellContent := colorBold.Sprint(`Uptime: `) + uptimeValue + "\n" +
			colorBold.Sprint(`Apdex: `) + apdexValue + "\n" +
			colorBold.Sprint(`MRT: `) + mrtValue

		regions := ""
		for i, v := range respJSON.Regions {
			if len(regions) > 0 {
				if i%4 == 0 {
					regions = regions + ",\n" + regionAliases[v]
				} else {
					regions = regions + ", " + regionAliases[v]
				}
			} else {
				regions = regionAliases[v]
			}
		}

		tableMainSettingsCellContent := colorBold.Sprint(`Checking interval: `) + strconv.Itoa(respJSON.Interval) + ` s ` + "\n" +
			upHTTPCodesLine +
			colorBold.Sprint(`Target response time: `) + fmt.Sprintf("%.3f s", respJSON.Target) + "\n" +
			colorBold.Sprint(`Thresholds: `) + `UP - ` + strconv.Itoa(respJSON.UpConfirmationsThreshold) + `, DOWN - ` + strconv.Itoa(respJSON.DownConfirmationsThreshold) + "\n" +
			colorBold.Sprint(`Binocs regions: `) + regions

		tableMainColumnDefinitions := []tableColumnDefinition{
			{
				Header:    "METRICS (" + periodTableTitle + ")",
				Priority:  1,
				Alignment: tablewriter.ALIGN_LEFT,
			},
			{
				Header:    "CHECK",
				Priority:  1,
				Alignment: tablewriter.ALIGN_LEFT,
			},
			{
				Header:    "SETTINGS",
				Priority:  2,
				Alignment: tablewriter.ALIGN_LEFT,
			},
		}

		var tableMainData [][]string
		tableMainData = append(tableMainData, []string{tableMainMetricsCellContent, tableMainCheckCellContent, tableMainSettingsCellContent})
		tableMain := composeTable(tableMainData, tableMainColumnDefinitions)

		// Combined table

		tableChartsColumnDefinitions := []tableColumnDefinition{
			{
				Header:    "",
				Priority:  1,
				Alignment: tablewriter.ALIGN_LEFT,
			},
		}

		var tableChartsData [][]string

		// Sub-table "http response codes"

		if respJSON.Protocol == protocolHTTP || respJSON.Protocol == protocolHTTPS {
			responseCodesData, err := util.BinocsAPI("/checks/"+respJSON.Ident+"/response-codes?"+urlValues.Encode(), http.MethodGet, []byte{})
			if err != nil {
				handleErr(err)
			}
			responseCodes := make([]ResponseCodesResponse, 0)
			decoder = json.NewDecoder(bytes.NewBuffer(responseCodesData))
			err = decoder.Decode(&responseCodes)
			if err != nil {
				handleErr(err)
			}

			responseCodesChart := drawResponseCodesChart(responseCodes, aggregateMetricsDataPoints[urlValues.Get("period")], respJSON.UpCodes, 16)
			responseCodesChartTitle := drawChartTitle("HTTP RESPONSE CODES", responseCodesChart, periodTableTitle)
			tableChartsData = append(tableChartsData, []string{responseCodesChartTitle})
			tableChartsData = append(tableChartsData, []string{responseCodesChart})
		}

		// Sub-table "apdex trend"

		apdexData, err := util.BinocsAPI("/checks/"+respJSON.Ident+"/apdex?"+urlValues.Encode(), http.MethodGet, []byte{})
		if err != nil {
			handleErr(err)
		}
		apdex := make([]ApdexResponse, 0)
		decoder = json.NewDecoder(bytes.NewBuffer(apdexData))
		err = decoder.Decode(&apdex)
		if err != nil {
			handleErr(err)
		}

		apdexChart := drawApdexChart(apdex, aggregateMetricsDataPoints[urlValues.Get("period")], "      ")
		apdexChartTitle := drawChartTitle("APDEX TREND", apdexChart, periodTableTitle)
		tableChartsData = append(tableChartsData, []string{apdexChartTitle})
		tableChartsData = append(tableChartsData, []string{apdexChart})

		// Sub-table "response times heatmap"

		responseTimeHeatmapData, err := util.BinocsAPI("/checks/"+respJSON.Ident+"/response-time-heatmap?"+urlValues.Encode(), http.MethodGet, []byte{})
		if err != nil {
			handleErr(err)
		}
		responseTimeHeatmap := make([]ResponseTimeHeatmapResponse, 0)
		decoder = json.NewDecoder(bytes.NewBuffer(responseTimeHeatmapData))
		err = decoder.Decode(&responseTimeHeatmap)
		if err != nil {
			handleErr(err)
		}

		responseTimeHeatmapChart := drawResponseTimeHeatmapChart(responseTimeHeatmap, aggregateMetricsDataPoints[urlValues.Get("period")], respJSON.Target, "")
		responseTimeHeatmapChartTitle := drawChartTitle("RESPONSE TIME HEATMAP", responseTimeHeatmapChart, periodTableTitle)
		tableChartsData = append(tableChartsData, []string{responseTimeHeatmapChartTitle})
		tableChartsData = append(tableChartsData, []string{responseTimeHeatmapChart})

		// Timeline

		timeline := drawTimeline(&user, urlValues.Get("period"), aggregateMetricsDataPoints[urlValues.Get("period")], "                ")
		tableChartsData = append(tableChartsData, []string{timeline})

		tableCharts := composeTable(tableChartsData, tableChartsColumnDefinitions)
		tableCharts.SetRowLine(true)

		spin.Stop()
		if user.CreditBalance == 0 {
			printZeroCreditsWarning()
		}
		tableMain.Render()
		tableCharts.Render()
	},
}

var checkListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all checks with status and metrics overview",
	Long: `
List all checks with status and metrics overview.
`,
	Aliases:           []string{"ls"},
	Args:              cobra.NoArgs,
	DisableAutoGenTag: true,
	Run: func(cmd *cobra.Command, args []string) {
		util.VerifyAuthenticated()

		if checkListFlagWatch {
			runAsWatch()
			return
		}

		urlValues1 := url.Values{}
		urlValues2 := url.Values{
			"period": []string{"day"},
		}
		apdexPeriodTableTitle := "1 DAY"

		match, err := regexp.MatchString(validPeriodPattern, checkListFlagPeriod)
		if err == nil && match {
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

		match = isValidRegionAlias(checkListFlagRegion)
		if len(checkListFlagRegion) > 0 && !match {
			handleErr(fmt.Errorf("Invalid region provided. Supported regions: " + strings.Join(getSupportedRegionAliases(), ", ")))
		} else if err == nil && match {
			urlValues2.Set("region", getRegionIdByAlias(checkListFlagRegion))
		}

		checkListFlagStatus = strings.ToUpper(checkListFlagStatus)
		if checkListFlagStatus == statusNameUp || checkListFlagStatus == statusNameDown {
			urlValues1.Set("status", checkListFlagStatus)
		}

		spin.Start()
		defer spin.Stop()
		spin.Suffix = colorFaint.Sprint(" loading checks...")

		user, err := fetchUser()
		if err != nil {
			handleErr(err)
		}

		checks, err := fetchChecks(urlValues1)
		if err != nil {
			handleErr(err)
		}
		ch := make(chan []string)
		var tableData [][]string
		var checksLen int
		for _, v := range checks {
			if urlValues2.Has("region") && !util.StringInSlice(urlValues2.Get("region"), v.Regions) {
				continue
			}
			go makeCheckListRow(v, ch, &urlValues2, user.CreditBalance == 0)
			checksLen++
		}
		var i int
		for _, v := range checks {
			if urlValues2.Has("region") && !util.StringInSlice(urlValues2.Get("region"), v.Regions) {
				continue
			}
			i++
			spin.Suffix = colorFaint.Sprintf(" loading metrics... (%d/%d)", i, checksLen)
			tableData = append(tableData, <-ch)
		}
		sort.Slice(tableData, func(i, j int) bool {
			return strings.ToLower(tableData[i][1]) < strings.ToLower(tableData[j][1])
		})

		columnDefinitions := []tableColumnDefinition{
			{
				Header:    "ID",
				Priority:  1,
				Alignment: tablewriter.ALIGN_LEFT,
			},
			{
				Header:    "NAME",
				Priority:  2,
				Alignment: tablewriter.ALIGN_LEFT,
			},
			{
				Header:    "URL/HOST",
				Priority:  3,
				Alignment: tablewriter.ALIGN_LEFT,
			},
			{
				Header:    "METHOD",
				Priority:  3,
				Alignment: tablewriter.ALIGN_LEFT,
			},
			{
				Header:    "STATUS",
				Priority:  2,
				Alignment: tablewriter.ALIGN_LEFT,
			},
			{
				Header:    "CHAN",
				Priority:  4,
				Alignment: tablewriter.ALIGN_LEFT,
			},
			{
				Header:    "HTTP",
				Priority:  3,
				Alignment: tablewriter.ALIGN_RIGHT,
			},
			{
				Header:    "MRT",
				Priority:  2,
				Alignment: tablewriter.ALIGN_RIGHT,
			},
			{
				Header:    "UPTIME",
				Priority:  2,
				Alignment: tablewriter.ALIGN_RIGHT,
			},
			{
				Header:    "APDEX",
				Priority:  2,
				Alignment: tablewriter.ALIGN_RIGHT,
			},
			{
				Header:    "APDEX " + apdexPeriodTableTitle,
				Priority:  4,
				Alignment: tablewriter.ALIGN_RIGHT,
			},
		}

		decorateStatusColumn(tableData)
		table := composeTable(tableData, columnDefinitions)
		spin.Stop()
		if user.CreditBalance == 0 {
			printZeroCreditsWarning()
		}
		table.Render()
	},
}

func makeCheckListRow(check Check, ch chan<- []string, urlValues *url.Values, zeroCredits bool) {
	lastStatusCodeRegex, _ := regexp.Compile(`^[1-5]{1}[0-9]{2}`)
	lastStatusCodeMatch := lastStatusCodeRegex.FindString(check.LastStatusCode)
	if lastStatusCodeMatch == "" {
		lastStatusCodeMatch = "-"
	}
	metrics, err := fetchMetrics(check.Ident, urlValues)
	if err != nil {
		handleErr(err)
	}
	apdexData, err := util.BinocsAPI("/checks/"+check.Ident+"/apdex?"+urlValues.Encode(), http.MethodGet, []byte{})
	if err != nil {
		handleErr(err)
	}
	apdex := make([]ApdexResponse, 0)
	decoder := json.NewDecoder(bytes.NewBuffer(apdexData))
	err = decoder.Decode(&apdex)
	if err != nil {
		handleErr(err)
	}
	apdexChart := drawCompactApdexChart(apdex, metrics.Apdex)
	tableValueMRT := formatMRT(metrics.MRT)
	tableValueUptime := formatUptime(metrics.Uptime)
	tableValueApdex := formatApdex(metrics.Apdex)
	if tableValueMRT == "n/a" && tableValueUptime == "n/a" && tableValueApdex == "n/a" {
		tableValueMRT = "[waiting for data]"
		tableValueUptime = "[waiting for data]"
		tableValueApdex = "[waiting for data]"
	}
	if metrics.Apdex == "" {
		apdexChart = ""
	}
	var method, name string
	if check.Protocol == protocolHTTP || check.Protocol == protocolHTTPS {
		method = check.Method
	} else {
		method = colorFaint.Sprint("-")
	}
	if check.Name == "" {
		name = colorFaint.Sprint("-")
	} else {
		name = colorBold.Sprint(check.Name)
	}
	var identSnippet, statusSnippet, lastStatusCodeSnippet string
	identSnippet = colorBold.Sprint(check.Ident)
	statusSnippet = formatStatus(&check)
	lastStatusCodeSnippet = lastStatusCodeMatch
	if zeroCredits {
		statusSnippet = color.YellowString(statusName[statusUnknown])
		lastStatusCodeSnippet = "n/a"
		tableValueMRT = "n/a"
		tableValueUptime = "n/a"
		tableValueApdex = "n/a"
	}
	tableRow := []string{
		identSnippet, name, util.Ellipsis(check.Resource, 40), colorFaint.Sprint(method), statusSnippet,
		colorFaint.Sprint(strconv.Itoa(len(check.Channels))), lastStatusCodeSnippet, tableValueMRT, tableValueUptime, tableValueApdex, apdexChart,
	}
	ch <- tableRow
}

func decorateStatusColumn(tableData [][]string) {
	var statusColumnIndex = 4
	var delimiter = " for "
	var minStatusColumnSpace = 1
	var maxStatusColumnLen int
	for _, td := range tableData {
		statusColumnLen := ansi.PrintableRuneWidth(td[statusColumnIndex]) - ansi.PrintableRuneWidth(delimiter) + minStatusColumnSpace
		if statusColumnLen > maxStatusColumnLen {
			maxStatusColumnLen = statusColumnLen
		}
	}
	for i, td := range tableData {
		statusColumnComps := strings.Split(td[statusColumnIndex], delimiter)
		if len(statusColumnComps) == 2 {
			statusColumnSpacer := strings.Repeat(" ", maxStatusColumnLen-ansi.PrintableRuneWidth(statusColumnComps[0])-ansi.PrintableRuneWidth(statusColumnComps[1]))
			tableData[i][statusColumnIndex] = statusColumnComps[0] + statusColumnSpacer + statusColumnComps[1]
		}
	}
}

var checkUpdateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update attributes of an existing check",
	Long: `
Update attributes of an existing check.

This command is interactive and asks user for parameters that were not provided as flags.
`,
	Args:              cobra.ExactArgs(1),
	DisableAutoGenTag: true,
	Run: func(cmd *cobra.Command, args []string) {
		util.VerifyAuthenticated()
		checkAddOrUpdate("update", args[0])
	},
}

var checkDeleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete existing check(s) and collected metrics",
	Long: `
Delete existing check(s) and collected metrics.

This command is interactive and asks user for confirmation.
`,
	Aliases:           []string{"del", "rm"},
	Args:              cobra.MatchAll(),
	DisableAutoGenTag: true,
	Run: func(cmd *cobra.Command, args []string) {
		util.VerifyAuthenticated()
		for _, arg := range args {
			respData, err := util.BinocsAPI("/checks/"+arg, http.MethodGet, []byte{})
			if err != nil {
				handleWarn("Error loading check " + arg)
				continue
			}
			var respJSON Check
			err = json.Unmarshal(respData, &respJSON)
			if err != nil {
				handleWarn("Invalid response from binocs.sh")
				continue
			}
			prompt := &survey.Confirm{
				Message: "Delete " + respJSON.Ident + " " + respJSON.Identity() + "?",
			}
			var yes bool
			err = survey.AskOne(prompt, &yes)
			if err != nil {
				continue
			}
			if yes {
				_, err = util.BinocsAPI("/checks/"+arg, http.MethodDelete, []byte{})
				if err != nil {
					handleWarn("Error deleting check " + arg)
					continue
				} else {
					fmt.Println("Check successfully deleted")
				}
			} else {
				fmt.Println("OK, skipping")
			}
		}
	},
}

func fetchChecks(urlValues url.Values) ([]Check, error) {
	var checks []Check
	respData, err := util.BinocsAPI("/checks?"+urlValues.Encode(), http.MethodGet, []byte{})
	if err != nil {
		return checks, err
	}
	checks = make([]Check, 0)
	decoder := json.NewDecoder(bytes.NewBuffer(respData))
	err = decoder.Decode(&checks)
	if err != nil {
		return checks, err
	}
	return checks, nil
}

func fetchMetrics(ident string, urlValues *url.Values) (MetricsResponse, error) {
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

func formatStatus(c *Check) string {
	var snippet string
	switch c.LastStatus {
	case statusDown:
		snippet = color.RedString(statusName[c.LastStatus]) + " for " + util.OutputDurationWithDays(c.LastStatusDuration)
	case statusStepDown:
		snippet = color.YellowString(statusName[c.LastStatus]) + " for " + util.OutputDurationWithDays(c.LastStatusDuration)
	case statusStepUp:
		snippet = color.YellowString(statusName[c.LastStatus]) + " for " + util.OutputDurationWithDays(c.LastStatusDuration)
	case statusUnknown:
		snippet = color.YellowString(statusName[c.LastStatus]) + " for " + util.OutputDurationWithDays(c.LastStatusDuration)
	case statusUp:
		snippet = color.GreenString(statusName[c.LastStatus]) + " for " + util.OutputDurationWithDays(c.LastStatusDuration)
	}
	return snippet
}

func formatMRT(mrt string) string {
	if mrt == "" || mrt == "nil" {
		return colorFaint.Sprint("n/a")
	}
	return mrt + " s"
}

func formatUptime(uptime string) string {
	var empty = "n/a"
	var uptimeFloat, err = strconv.ParseFloat(uptime, 32)
	if uptime == "" || uptime == "nil" {
		return color.HiBlackString(empty)
	}
	if err != nil {
		return color.HiBlackString(empty)
	}
	if uptimeFloat == 100.0 {
		return color.GreenString("%v %%", uptime)
	}
	if uptimeFloat > 99.9 {
		return color.YellowString("%v %%", uptime)
	}
	return color.RedString("%v %%", uptime)
}

func formatApdex(apdex string) string {
	var empty = "n/a"
	var apdexFloat, err = strconv.ParseFloat(apdex, 32)
	if apdex == "" || apdex == "nil" {
		return color.HiBlackString(empty)
	}
	if err != nil {
		return color.HiBlackString(empty)
	}
	if apdexFloat >= 0.8 {
		return color.GreenString("%v", apdex)
	}
	if apdexFloat >= 0.6 {
		return color.YellowString("%v", apdex)
	}
	return color.RedString("%v", apdex)
}

func drawCompactApdexChart(apdexChartData []ApdexResponse, currentApdex string) string {
	var chart []rune
	var chartSnippet string
	var alphabet = map[string]rune{
		"11": '⣀',
		"12": '⣠',
		"13": '⣰',
		"14": '⣸',
		"21": '⣄',
		"22": '⣤',
		"23": '⣴',
		"24": '⣼',
		"31": '⣆',
		"32": '⣦',
		"33": '⣶',
		"34": '⣾',
		"41": '⣇',
		"42": '⣧',
		"43": '⣷',
		"44": '⣿',
	}
	var reverseApdex []ApdexResponse
	for _, v := range apdexChartData {
		reverseApdex = append([]ApdexResponse{v}, reverseApdex...)
	}
	var assignChar = func(left, right float64) rune {
		const steps = 4
		var leftDots, rightDots string
		for j := 1; j < 1+steps; j++ {
			if left <= float64(j)/steps {
				leftDots = strconv.Itoa(j)
				break
			}
		}
		for k := 1; k < 1+steps; k++ {
			if right <= float64(k)/steps {
				rightDots = strconv.Itoa(k)
				break
			}
		}
		return alphabet[rightDots+leftDots]
	}
	for i, v := range reverseApdex {
		if i%2 == 1 { // even
			left, _ := strconv.ParseFloat(reverseApdex[i-1].Apdex, 32)
			right, _ := strconv.ParseFloat(v.Apdex, 32)
			chart = append(chart, assignChar(left, right)) // chart + assignChar(left, right)
		} else if len(reverseApdex) == i+1 { // last
			left, _ := strconv.ParseFloat(v.Apdex, 32)
			chart = append(chart, assignChar(left, 0.0))
		}
	}

	chartSnippet = util.Reverse(string(chart))

	var apdexFloat, err = strconv.ParseFloat(currentApdex, 32)
	if err != nil {
		return color.HiBlackString(chartSnippet)
	}
	if apdexFloat >= 0.9 {
		return color.GreenString("%v", chartSnippet)
	}
	if apdexFloat >= 0.6 {
		return color.YellowString("%v", chartSnippet)
	}
	return color.RedString("%v", chartSnippet)
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
			if v.Apdex == "nil" {
				rows[i] = rows[i] + " "
			} else if vf > (float64(i)+1.0)/float64(numRows) {
				rows[i] = rows[i] + " "
			} else if vf <= (float64(i)+1.0)/float64(numRows) && vf >= float64(i)/float64(numRows) {
				rows[i] = rows[i] + "▩"
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
		rowFraction := (float64(i) + 1.0) / float64(numRows)
		if rowFraction > 0.8 {
			chart = leftMargin + getApdexChartRowRange(i, numRows) + " " + color.GreenString(rows[i]) + "\n" + chart
		} else if rowFraction > 0.6 {
			chart = leftMargin + getApdexChartRowRange(i, numRows) + " " + color.YellowString(rows[i]) + "\n" + chart
		} else {
			chart = leftMargin + getApdexChartRowRange(i, numRows) + " " + color.RedString(rows[i]) + "\n" + chart
		}
	}
	chart = strings.TrimSuffix(chart, "\n")
	return chart
}

func drawResponseCodesChart(responseCodes []ResponseCodesResponse, dataPoints int, upCodes string, yAreaWidth int) string {
	const numRows = 6
	var rows [numRows]string
	var chart string
	for _, v := range responseCodes {
		if v.Xx1 > 0 {
			rows[0] = rows[0] + "▩"
		} else {
			rows[0] = rows[0] + " "
		}
		if v.Xx2 > 0 {
			rows[1] = rows[1] + "▩"
		} else {
			rows[1] = rows[1] + " "
		}
		if v.Xx3 > 0 {
			rows[2] = rows[2] + "▩"
		} else {
			rows[2] = rows[2] + " "
		}
		if v.Xx4 > 0 {
			rows[3] = rows[3] + "▩"
		} else {
			rows[3] = rows[3] + " "
		}
		if v.Xx5 > 0 {
			rows[4] = rows[4] + "▩"
		} else {
			rows[4] = rows[4] + " "
		}
		if v.Err > 0 {
			rows[5] = rows[5] + "▩"
		} else {
			rows[5] = rows[5] + " "
		}
	}
	if len(responseCodes) < dataPoints {
		for i := range rows {
			rows[i] = strings.Repeat(" ", dataPoints-len(responseCodes)) + rows[i]
		}
	}
	for i := 0; i < numRows; i++ {
		if i == numRows-1 { // the err case
			chart = chart + strings.Repeat(" ", yAreaWidth-6) + "error" + " " + color.RedString(rows[i]) + "\n"
		} else {
			// @todo we should distinguish between a green code (in range and up), and a yellow code (in range and down, e.g. 303 in default range setup)
			if ok, _ := util.IsCodeInRange((i+1)*100, upCodes); ok {
				chart = chart + strings.Repeat(" ", yAreaWidth-4) + strconv.Itoa(i+1) + "xx" + " " + color.GreenString(rows[i]) + "\n"
			} else {
				chart = chart + strings.Repeat(" ", yAreaWidth-4) + strconv.Itoa(i+1) + "xx" + " " + color.RedString(rows[i]) + "\n"
			}
		}
	}
	chart = strings.TrimSuffix(chart, "\n")
	return chart
}

func drawResponseTimeHeatmapChart(responseTimeHeatmap []ResponseTimeHeatmapResponse, dataPoints int, targetTime float64, leftMargin string) string {
	var rowThresholds = []float32{
		8.0, 4.0, 2.0, 1.0, 0.5, 0.25, 0.125, 0.0,
	}
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
	var palette = [5]string{" ", "▨", "▨", "▩", "▩"}
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
	if heatmapMaximum > 0 {
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
	}
	if len(responseTimeHeatmap) < dataPoints {
		for i := range rows {
			rows[i] = strings.Repeat(" ", dataPoints-len(responseTimeHeatmap)) + rows[i]
		}
	}
	for i := 0; i < len(rows); i++ {
		if float64(rowThresholds[i]) > targetTime {
			chart = chart + leftMargin + rowTitles[i] + " " + color.RedString(rows[i]) + "\n"
		} else if i > 1 && float64(rowThresholds[i-1]) > targetTime {
			chart = chart + leftMargin + rowTitles[i] + " " + color.YellowString(rows[i]) + "\n"
		} else {
			chart = chart + leftMargin + rowTitles[i] + " " + color.GreenString(rows[i]) + "\n"
		}
	}
	chart = strings.TrimSuffix(chart, "\n")
	return chart
}

func drawChartTitle(title string, chart string, periodTitle string) string {
	chartRows := strings.Split(chart, "\n")
	chartWidth := ansi.PrintableRuneWidth(chartRows[0])
	spacerWidth := chartWidth - len(title) - len(periodTitle)
	if len(title)+len(periodTitle)+1 < chartWidth {
		title = title + strings.Repeat(" ", spacerWidth) + periodTitle
	}
	return colorBold.Sprint(title)
}

func drawTimeline(user *User, period string, dataPoints int, leftMargin string) string {
	var timeline [2]string

	tz, err := time.LoadLocation(user.Timezone)
	if err != nil {
		tz = time.UTC
	}

	var now = time.Now().In(tz)
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
				timeline[1] = now.Format("Mon") + strings.Repeat(" ", gap) + timeline[1]
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
				timeline[1] = now.Format("Jan") + strings.Repeat(" ", gap) + timeline[1]
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

func setProtocolPrefix(res, proto string) string {
	var lcp = strings.ToLower(proto)
	m, _ := regexp.MatchString(`^`+lcp+`:\/\/`, res)
	if !m {
		res = lcp + "://" + res
	}
	return res
}

func checkAddOrUpdate(mode string, checkIdent string) {
	if mode != "add" && mode != "update" {
		handleErr(fmt.Errorf("Unknown mode: " + mode))
	}

	var err error
	var match bool
	var tpl string

	var (
		flagName                       string
		flagProtocol                   string
		flagResource                   string
		flagMethod                     string
		flagInterval                   int
		flagTarget                     float64
		flagRegions                    []string
		flagUpCodes                    string
		flagUpConfirmationsThreshold   int
		flagDownConfirmationsThreshold int
		flagAttach                     []string
	)

	switch mode {
	case "add":
		flagName = checkAddFlagName
		flagProtocol = checkAddFlagProtocol
		flagResource = checkAddFlagResource
		flagMethod = checkAddFlagMethod
		flagInterval = checkAddFlagInterval
		flagTarget = checkAddFlagTarget
		flagRegions = checkAddFlagRegions
		flagUpCodes = checkAddFlagUpCodes
		flagUpConfirmationsThreshold = checkAddFlagUpConfirmationsThreshold
		flagDownConfirmationsThreshold = checkAddFlagDownConfirmationsThreshold
		flagAttach = checkAddFlagAttach
	case "update":
		flagName = checkUpdateFlagName
		flagMethod = checkUpdateFlagMethod
		flagInterval = checkUpdateFlagInterval
		flagTarget = checkUpdateFlagTarget
		flagRegions = checkUpdateFlagRegions
		flagUpCodes = checkUpdateFlagUpCodes
		flagUpConfirmationsThreshold = checkUpdateFlagUpConfirmationsThreshold
		flagDownConfirmationsThreshold = checkUpdateFlagDownConfirmationsThreshold
		flagAttach = checkUpdateFlagAttach
	}

	var currentCheck Check
	if mode == "update" {
		spin.Start()
		defer spin.Stop()
		spin.Suffix = colorFaint.Sprint(" loading check...")
		respData, err := util.BinocsAPI("/checks/"+checkIdent, http.MethodGet, []byte{})
		if err != nil {
			handleErr(err)
		}
		err = json.Unmarshal(respData, &currentCheck)
		if err != nil {
			handleErr(err)
		}
		spin.Stop()
	}

	match, err = regexp.MatchString(validNamePattern, flagName)
	if err != nil {
		handleErr(err)
	} else if !match || flagName == "" {
		validate := func(val interface{}) error {
			match, err = regexp.MatchString(validNamePattern, val.(string))
			if err != nil {
				return err
			} else if !match {
				return errors.New("invalid name format")
			}
			return nil
		}
		prompt := &survey.Input{
			Message: "Check name (optional):",
		}
		if mode == "update" {
			prompt.Default = currentCheck.Name
		}
		err = survey.AskOne(prompt, &flagName, survey.WithValidator(validate))
		if err != nil {
			handleErr(err)
		}
	}

	if mode == "update" {
		// pass; never update check protocol
	} else {
		match, err = regexp.MatchString(validProtocolPattern, flagProtocol)
		if err != nil {
			handleErr(err)
		} else if !match || flagProtocol == "" {
			prompt := &survey.Select{
				Message: "Protocol:",
				Options: []string{protocolHTTP, protocolHTTPS, protocolTCP},
				Default: protocolHTTPS,
			}
			err := survey.AskOne(prompt, &flagProtocol)
			if err != nil {
				handleErr(err)
			}
		}
		flagProtocol = strings.ToUpper(flagProtocol)
	}

	if mode == "update" {
		// pass; never update check resource
	} else {
		var isValidResource bool
		var message string
		switch flagProtocol {
		case protocolHTTP:
			isValidResource = isValidHTTPResource(flagResource)
			message = "URL:"
		case protocolHTTPS:
			isValidResource = isValidHTTPSResource(flagResource)
			message = "URL:"
		case protocolICMP:
			isValidResource = isValidICMPResource(flagResource)
			message = "Hostname:"
		case protocolTCP:
			isValidResource = isValidTCPResource(flagResource)
			message = "Hostname and port:"
		}
		if !isValidResource {
			validate := func(val interface{}) error {
				switch flagProtocol {
				case protocolHTTP:
					if !isValidHTTPResource(val.(string)) {
						return errors.New("invalid HTTP URL")
					}
				case protocolHTTPS:
					if !isValidHTTPSResource(val.(string)) {
						return errors.New("invalid HTTPS URL")
					}
				case protocolICMP:
					if !isValidICMPResource(val.(string)) {
						return errors.New("invalid ICMP host")
					}
				case protocolTCP:
					if !isValidTCPResource(val.(string)) {
						return errors.New("invalid TCP <host>:<port>")
					}
				}
				return nil
			}
			prompt := &survey.Input{
				Message: message,
				// Help:    "",
			}
			err = survey.AskOne(prompt, &flagResource, survey.WithValidator(validate))
			if err != nil {
				handleErr(err)
			}
		}
		flagResource = setProtocolPrefix(flagResource, flagProtocol)
	}

	if flagProtocol == protocolHTTP || flagProtocol == protocolHTTPS || currentCheck.Protocol == protocolHTTP || currentCheck.Protocol == protocolHTTPS {
		match, err = regexp.MatchString(validMethodPattern, flagMethod)
		if err != nil {
			handleErr(err)
		} else if !match || flagMethod == "" {
			prompt := &survey.Select{
				Message: "HTTP method:",
				Options: []string{"GET", "HEAD", "POST", "PUT", "DELETE"},
				Default: "GET",
			}
			if mode == "update" {
				prompt.Default = currentCheck.Method
			}
			err := survey.AskOne(prompt, &flagMethod)
			if err != nil {
				handleErr(err)
			}
		}
	} else {
		flagMethod = ""
	}

	if flagInterval < supportedIntervalMinimum || flagInterval > supportedIntervalMaximum {
		validate := func(val interface{}) error {
			var inputInt, _ = strconv.Atoi(val.(string))
			if inputInt < supportedIntervalMinimum || inputInt > supportedIntervalMaximum {
				return errors.New("Interval must be a value between " + strconv.Itoa(supportedIntervalMinimum) + " and " + strconv.Itoa(supportedIntervalMaximum))
			}
			return nil
		}
		prompt := &survey.Input{
			Message: "Interval in seconds:",
			Help:    "Interval must be a value between " + strconv.Itoa(supportedIntervalMinimum) + " and " + strconv.Itoa(supportedIntervalMaximum),
			Default: "60",
		}
		if mode == "update" {
			prompt.Default = fmt.Sprintf("%d", currentCheck.Interval)
		}
		err := survey.AskOne(prompt, &flagInterval, survey.WithValidator(validate))
		if err != nil {
			handleErr(err)
		}
	}

	if flagTarget < supportedTargetMinimum || flagTarget > supportedTargetMaximum {
		validate := func(val interface{}) error {
			var inputFloat, _ = strconv.ParseFloat(val.(string), 64)
			if inputFloat < supportedTargetMinimum || inputFloat > supportedTargetMaximum {
				return errors.New("Target Response Time must be a value between " + fmt.Sprintf("%.3f", supportedTargetMinimum) + " and " + fmt.Sprintf("%.3f", supportedTargetMaximum))
			}
			return nil
		}
		prompt := &survey.Input{
			Message: "Target Response Time in seconds:",
			Help:    "Target Response Time must be a value between " + fmt.Sprintf("%.3f", supportedTargetMinimum) + " and " + fmt.Sprintf("%.3f", supportedTargetMaximum),
			Default: "1.20",
		}
		if mode == "update" {
			prompt.Default = fmt.Sprintf("%.3f", currentCheck.Target)
		}
		err := survey.AskOne(prompt, &flagTarget, survey.WithValidator(validate))
		if err != nil {
			handleErr(err)
		}
	}

	match = true
	for _, fr := range flagRegions {
		if !isValidRegionAlias(fr) {
			match = false
		}
	}
	if !match || len(flagRegions) == 0 {
		prompt := &survey.MultiSelect{
			Message:  "Regions:",
			Options:  getSupportedRegionAliases(),
			PageSize: len(getSupportedRegionAliases()),
		}
		if mode == "update" {
			prompt.Default = getRegionAliasesByIds(currentCheck.Regions)
		} else {
			prompt.Default = getDefaultRegionAliases()
		}
		err = survey.AskOne(prompt, &flagRegions, survey.WithValidator(survey.MinItems(1)))
		if err != nil {
			handleErr(err)
		}
	}

	if flagProtocol == protocolHTTP || flagProtocol == protocolHTTPS || currentCheck.Protocol == protocolHTTP || currentCheck.Protocol == protocolHTTPS {
		match, err = regexp.MatchString(validUpCodePattern, flagUpCodes)
		if err != nil {
			handleErr(err)
		} else if !match || flagUpCodes == "" {
			validate := func(val interface{}) error {
				match, err = regexp.MatchString(validUpCodePattern, val.(string))
				if err != nil {
					return err
				} else if !match {
					return errors.New("invalid input value")
				}
				return nil
			}
			prompt := &survey.Input{
				Message: "What are the good (\"up\") HTTP(S) response codes, e.g. \"2xx\" or \"200-302\", or \"200,301\":",
				Default: "200-302",
			}
			if mode == "update" {
				prompt.Default = currentCheck.UpCodes
			}
			err := survey.AskOne(prompt, &flagUpCodes, survey.WithValidator(validate))
			if err != nil {
				handleErr(err)
			}
		}
	} else {
		flagUpCodes = ""
	}

	if mode == "update" && flagUpConfirmationsThreshold == 0 {
		// pass
	} else {
		if flagUpConfirmationsThreshold < supportedConfirmationsThresholdMinimum || flagUpConfirmationsThreshold > supportedConfirmationsThresholdMaximum {
			validate := func(val interface{}) error {
				var inputInt, _ = strconv.Atoi(val.(string))
				if inputInt < supportedConfirmationsThresholdMinimum || inputInt > supportedConfirmationsThresholdMaximum {
					return errors.New("Up Confirmations Threshold must be a value between " + strconv.Itoa(supportedConfirmationsThresholdMinimum) + " and " + strconv.Itoa(supportedConfirmationsThresholdMaximum))
				}
				return nil
			}
			prompt := &survey.Input{
				Message: "Up Confirmations Threshold:",
				Help:    "Up Confirmations Threshold must be a value between " + strconv.Itoa(supportedConfirmationsThresholdMinimum) + " and " + strconv.Itoa(supportedConfirmationsThresholdMaximum),
				Default: "2",
			}
			err := survey.AskOne(prompt, &flagUpConfirmationsThreshold, survey.WithValidator(validate))
			if err != nil {
				handleErr(err)
			}
		}
	}

	if mode == "update" && flagDownConfirmationsThreshold == 0 {
		// pass
	} else {
		// check DownConfirmationsThreshold is in supported range
		if flagDownConfirmationsThreshold < supportedConfirmationsThresholdMinimum || flagDownConfirmationsThreshold > supportedConfirmationsThresholdMaximum {
			validate := func(val interface{}) error {
				var inputInt, _ = strconv.Atoi(val.(string))
				if inputInt < supportedConfirmationsThresholdMinimum || inputInt > supportedConfirmationsThresholdMaximum {
					return errors.New("Down Confirmations Threshold must be a value between " + strconv.Itoa(supportedConfirmationsThresholdMinimum) + " and " + strconv.Itoa(supportedConfirmationsThresholdMaximum))
				}
				return nil
			}
			prompt := &survey.Input{
				Message: "Down Confirmations Threshold:",
				Help:    "Down Confirmations Threshold must be a value between " + strconv.Itoa(supportedConfirmationsThresholdMinimum) + " and " + strconv.Itoa(supportedConfirmationsThresholdMaximum),
			}
			err := survey.AskOne(prompt, &flagDownConfirmationsThreshold, survey.WithValidator(validate))
			if err != nil {
				handleErr(err)
			}
		}
	}

	spin.Start()
	defer spin.Stop()
	spin.Suffix = colorFaint.Sprint(" loading channels...")
	channels, err := fetchChannels(url.Values{})
	if err != nil {
		handleErr(err)
	}
	spin.Stop()

	if len(channels) > 0 {
		match, err = regexp.MatchString(validChannelsIdentListPattern, strings.Join(flagAttach, ","))
		if err != nil {
			handleErr(err)
		} else if !match || len(flagAttach) == 0 {
			var options = []string{}
			for _, ch := range channels {
				options = append(options, ch.Ident+" "+ch.Type+" "+ch.Identity())
			}
			var defaultOptions = []string{}
			if mode == "update" {
				for _, cc := range currentCheck.Channels {
					for _, ch := range channels {
						if ch.Ident == cc {
							defaultOption := ch.Ident + " " + ch.Type + " " + ch.Identity()
							defaultOptions = append(defaultOptions, defaultOption)
						}
					}
				}
			}
			prompt := &survey.MultiSelect{
				Message:  "Channels to attach (optional):",
				Options:  options,
				Default:  defaultOptions,
				PageSize: 9,
			}
			err = survey.AskOne(prompt, &flagAttach)
			if err != nil {
				handleErr(err)
			}
		} else if len(flagAttach) == 1 && flagAttach[0] == "all" {
			flagAttach = []string{}
			for _, v := range channels {
				flagAttach = append(flagAttach, v.Ident)
			}
		}
	} else {
		fmt.Println("No notification channels configured. Skipping.")
	}

	check := Check{
		Name:                       flagName,
		Protocol:                   flagProtocol,
		Resource:                   flagResource,
		Method:                     flagMethod,
		Interval:                   flagInterval,
		Target:                     flagTarget,
		Regions:                    getRegionIdsByAliases(flagRegions),
		UpCodes:                    flagUpCodes,
		UpConfirmationsThreshold:   flagUpConfirmationsThreshold,
		DownConfirmationsThreshold: flagDownConfirmationsThreshold,
	}
	postData, err := json.Marshal(check)
	if err != nil {
		handleErr(err)
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
	spin.Start()
	defer spin.Stop()
	spin.Suffix = colorFaint.Sprint(" saving check...")
	respData, err := util.BinocsAPI(reqURL, reqMethod, postData)
	if err != nil {
		handleErr(err)
	}
	err = json.Unmarshal(respData, &check)
	if err != nil {
		handleErr(err)
	}
	if check.ID > 0 {
		var checkDescription string
		if len(check.Name) > 0 {
			checkDescription = check.Name + " (" + check.Resource + ")"
		} else {
			checkDescription = check.Resource
		}
		if mode == "add" {
			tpl = "[" + check.Ident + "] " + checkDescription + ` added successfully`
		}
		if mode == "update" {
			tpl = "[" + check.Ident + "] " + checkDescription + ` updated successfully`
		}
		spin.Suffix = colorFaint.Sprintf(" attaching check to %d channel(s)...", len(flagAttach))
		var detachChannelIdents = []string{}
		for _, ch := range channels {
			for _, cc := range ch.Checks {
				if cc == check.Ident {
					detachChannelIdents = append(detachChannelIdents, ch.Ident)
				}
			}
		}
		for _, ch := range detachChannelIdents {
			deleteData, err := json.Marshal(ChannelAttachment{})
			if err != nil {
				handleErr(err)
			}
			_, err = util.BinocsAPI("/channels/"+ch+"/check/"+check.Ident, http.MethodDelete, deleteData)
			if err != nil {
				handleErr(err)
			}
		}
		for _, fa := range flagAttach {
			attachIdent := strings.Split(fa, " ")[0]
			postData, err := json.Marshal(ChannelAttachment{})
			if err != nil {
				handleErr(err)
			}
			_, err = util.BinocsAPI("/channels/"+attachIdent+"/check/"+check.Ident, http.MethodPost, postData)
			if err != nil {
				handleErr(err)
			}
		}
	} else {
		if mode == "add" {
			handleErr(errors.New("error adding check"))
		}
		if mode == "update" {
			handleErr(errors.New("error updating check"))
		}
	}
	spin.Stop()
	fmt.Println(tpl)
}
