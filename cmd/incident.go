package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"

	util "github.com/automato-io/binocs-cli/util"
	"github.com/fatih/color"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
)

// Incident comes from the API as a JSON
type Incident struct {
	ID            int       `json:"id"`
	Ident         string    `json:"ident"`
	CheckID       int       `json:"check_id"`
	CheckIdent    string    `json:"check_ident"`
	CheckName     string    `json:"check_name"`
	CheckProtocol string    `json:"check_protocol"`
	CheckResource string    `json:"check_resource"`
	IncidentNote  string    `json:"incident_note"`
	IncidentState string    `json:"incident_state"`
	Opened        string    `json:"opened"`
	Closed        string    `json:"closed"`
	Duration      string    `json:"duration"`
	ResponseCodes []string  `json:"response_codes"`
	Requests      []Request `json:"requests"`
}

// Request struct
type Request struct {
	Region             string  `json:"region"`
	Status             int     `json:"status"`
	RequestProtocol    string  `json:"request_protocol"`
	RequestResource    string  `json:"request_resource"`
	RequestMethod      string  `json:"request_method"`
	ResponseStatusCode string  `json:"response_status"`
	Timings            Timings `json:"timings"`
	Timestamp          string  `json:"timestamp"`
}

// Timings struct
type Timings struct {
	DSNLookup  string `json:"dns_lookup"`
	Connection string `json:"connection"`
	TLS        string `json:"tls"`
	Wait       string `json:"wait"`
	Transfer   string `json:"transfer"`
}

// `incident ls` flags
var (
	incidentListFlagCheck    string
	incidentListFlagOpen     bool
	incidentListFlagResolved bool
)

// `incident update` flags
var (
	incidentUpdateFlagNote string
)

const (
	validCheckIdentPattern = `^[a-f0-9]{7}$`
)

func init() {
	rootCmd.AddCommand(incidentsCmd)

	rootCmd.AddCommand(incidentCmd)

	incidentCmd.AddCommand(incidentInspectCmd)
	incidentCmd.AddCommand(incidentListCmd)
	incidentCmd.AddCommand(incidentUpdateCmd)

	incidentsCmd.Flags().StringVarP(&incidentListFlagCheck, "check", "c", "", "list only incidents of this check")
	incidentsCmd.Flags().BoolVar(&incidentListFlagOpen, "open", false, "list only open incidents")
	incidentsCmd.Flags().BoolVar(&incidentListFlagResolved, "resolved", false, "list only resolved incidents")
	incidentListCmd.Flags().StringVarP(&incidentListFlagCheck, "check", "c", "", "list only incidents of this check")
	incidentListCmd.Flags().BoolVar(&incidentListFlagOpen, "open", false, "list only open incidents")
	incidentListCmd.Flags().BoolVar(&incidentListFlagResolved, "resolved", false, "list only resolved incidents")

	incidentUpdateCmd.Flags().StringVarP(&incidentUpdateFlagNote, "note", "n", "", "incident note")
}

var incidentCmd = &cobra.Command{
	Use:               "incident",
	Short:             "Manage incidents",
	DisableAutoGenTag: true,
}

var incidentsCmd = &cobra.Command{
	Use:               "incidents",
	Args:              cobra.NoArgs,
	Short:             incidentListCmd.Short,
	Long:              incidentListCmd.Long,
	DisableAutoGenTag: true,
	Run: func(cmd *cobra.Command, args []string) {
		incidentListCmd.Run(cmd, args)
	},
}

var incidentInspectCmd = &cobra.Command{
	Use:   "inspect",
	Short: "View incident details",
	Long: `
View incident details, notes and associated requests.
`,
	Aliases:           []string{"view", "show", "info"},
	Args:              cobra.ExactArgs(1),
	DisableAutoGenTag: true,
	Run: func(cmd *cobra.Command, args []string) {
		util.VerifyAuthenticated()

		spin.Start()
		defer spin.Stop()
		spin.Suffix = colorFaint.Sprint(" loading incident...")
		user, err := fetchUser()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		respData, err := util.BinocsAPI("/incidents/"+args[0], http.MethodGet, []byte{})
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		var respJSON Incident
		err = json.Unmarshal(respData, &respJSON)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		var checkName string
		if respJSON.CheckName == "" {
			checkName = "-"
		} else {
			checkName = respJSON.CheckName
		}

		// Table "main"

		var stateSnippet string
		switch respJSON.IncidentState {
		case incidentStateOpen:
			stateSnippet = color.YellowString(strings.ToUpper(respJSON.IncidentState))
		case incidentStateResolved:
			stateSnippet = color.GreenString(strings.ToUpper(respJSON.IncidentState))
		}

		var openedSnippet = respJSON.Opened

		var closedSnippet = respJSON.Closed
		if closedSnippet == "" {
			closedSnippet = "-"
		}

		tableMainIncidentCellContent := colorBold.Sprint(`ID: `) + respJSON.Ident + "\n" +
			colorBold.Sprint(`Status: `) + stateSnippet + "\n" +
			colorBold.Sprint(`Opened: `) + openedSnippet + "\n" +
			colorBold.Sprint(`Closed: `) + closedSnippet + "\n" +
			colorBold.Sprint(`Duration: `) + util.OutputDurationWithDays(respJSON.Duration)

		tableMainCheckCellContent := colorBold.Sprint(`ID: `) + respJSON.CheckIdent + "\n" +
			colorBold.Sprint("Name: ") + checkName + "\n" +
			colorBold.Sprint("URL: ") + respJSON.CheckResource

		tableMainNotesCellContent := respJSON.IncidentNote
		if tableMainNotesCellContent == "" {
			tableMainNotesCellContent = colorFaint.Sprint("-")
		}

		tableMainColumnDefinitions := []tableColumnDefinition{
			{
				Header:    "CHECK",
				Priority:  2,
				Alignment: tablewriter.ALIGN_LEFT,
			},
			{
				Header:    "INCIDENT",
				Priority:  1,
				Alignment: tablewriter.ALIGN_LEFT,
			},
			{
				Header:    "NOTES",
				Priority:  3,
				Alignment: tablewriter.ALIGN_LEFT,
			},
		}

		var tableMainData [][]string
		tableMainData = append(tableMainData, []string{tableMainCheckCellContent, tableMainIncidentCellContent, tableMainNotesCellContent})
		tableMain := composeTable(tableMainData, tableMainColumnDefinitions)

		// Table "requests"

		tableRequestsColumnDefinitions := []tableColumnDefinition{
			{
				Header:    "CHECKED AT",
				Priority:  1,
				Alignment: tablewriter.ALIGN_LEFT,
			},
			{
				Header:    "CHECKED FROM",
				Priority:  2,
				Alignment: tablewriter.ALIGN_LEFT,
			},
			{
				Header:    "RESPONSE",
				Priority:  1,
				Alignment: tablewriter.ALIGN_LEFT,
			},
			{
				Header:    "RESPONSE TIME",
				Priority:  2,
				Alignment: tablewriter.ALIGN_RIGHT,
			},
			{
				Header:    "DNS LOOKUP",
				Priority:  3,
				Alignment: tablewriter.ALIGN_RIGHT,
			},
			{
				Header:    "CONNECTION",
				Priority:  3,
				Alignment: tablewriter.ALIGN_RIGHT,
			},
			{
				Header:    "TLS",
				Priority:  3,
				Alignment: tablewriter.ALIGN_RIGHT,
			},
			{
				Header:    "WAITING",
				Priority:  3,
				Alignment: tablewriter.ALIGN_RIGHT,
			},
			{
				Header:    "TRANSFER",
				Priority:  4,
				Alignment: tablewriter.ALIGN_RIGHT,
			},
		}

		var tableRequests *tablewriter.Table
		var tableRequestsData [][]string
		if len(respJSON.Requests) > 0 {

			var placeholder = "·"
			var fieldLengthCheckedFrom int
			for _, request := range respJSON.Requests {
				if fieldLengthCheckedFrom < len(request.Region) {
					fieldLengthCheckedFrom = len(request.Region)
				}
			}
			for _, request := range respJSON.Requests {
				if strings.Contains(request.Timestamp, "0001") {
					sameSameSpace := len(request.Timestamp) - len(request.RequestResource+" requests") - 2
					sameSamePlaceholders := [2]int{0, 0}
					if sameSameSpace > 0 {
						if sameSameSpace%2 == 1 {
							sameSamePlaceholders[0] = sameSameSpace/2 + 1
							sameSamePlaceholders[1] = sameSameSpace / 2
						} else {
							sameSamePlaceholders[0] = sameSameSpace / 2
							sameSamePlaceholders[1] = sameSameSpace / 2
						}
					}
					sameSame := fmt.Sprintf("%s %s requests %s", strings.Repeat(placeholder, sameSamePlaceholders[0]), request.RequestResource, strings.Repeat(placeholder, sameSamePlaceholders[1]))
					tableRequestsData = append(tableRequestsData, []string{sameSame, strings.Repeat(placeholder, fieldLengthCheckedFrom), request.ResponseStatusCode, strings.Repeat(placeholder, 7), colorFaint.Sprint(strings.Repeat(placeholder, 7)),
						colorFaint.Sprint(strings.Repeat(placeholder, 7)), colorFaint.Sprint(strings.Repeat(placeholder, 7)), colorFaint.Sprint(strings.Repeat(placeholder, 7)), colorFaint.Sprint(strings.Repeat(placeholder, 7))})
				} else {
					var responseTime, timingsDNSLookup, timingsConnection, timingsTLS, timingsWait, timingsTransfer string
					var timingsDNSLookupFloat, timingsConnectionFloat, timingsTLSFloat, timingsWaitFloat, timingsTransferFloat float64
					if request.Timings.DSNLookup != "nil" {
						timingsDNSLookup = fmt.Sprintf("%s s", request.Timings.DSNLookup)
						timingsConnection = fmt.Sprintf("%s s", request.Timings.Connection)
						timingsTLS = fmt.Sprintf("%s s", request.Timings.TLS)
						timingsWait = fmt.Sprintf("%s s", request.Timings.Wait)
						timingsTransfer = fmt.Sprintf("%s s", request.Timings.Transfer)
						timingsDNSLookupFloat, _ = strconv.ParseFloat(request.Timings.DSNLookup, 32)
						timingsConnectionFloat, _ = strconv.ParseFloat(request.Timings.Connection, 32)
						timingsTLSFloat, _ = strconv.ParseFloat(request.Timings.TLS, 32)
						timingsWaitFloat, _ = strconv.ParseFloat(request.Timings.Wait, 32)
						timingsTransferFloat, _ = strconv.ParseFloat(request.Timings.Transfer, 32)
						responseTime = fmt.Sprintf("%.3f s", timingsDNSLookupFloat+timingsConnectionFloat+timingsTLSFloat+timingsWaitFloat+timingsTransferFloat)
					} else {
						responseTime = "n/a"
						timingsDNSLookup = "n/a"
						timingsConnection = "n/a"
						timingsTLS = "n/a"
						timingsWait = "n/a"
						timingsTransfer = "n/a"
					}
					tableRequestsData = append(tableRequestsData, []string{request.Timestamp, request.Region, request.ResponseStatusCode, responseTime, colorFaint.Sprint(timingsDNSLookup),
						colorFaint.Sprint(timingsConnection), colorFaint.Sprint(timingsTLS), colorFaint.Sprint(timingsWait), colorFaint.Sprint(timingsTransfer)})
				}
			}
			tableRequests = composeTable(tableRequestsData, tableRequestsColumnDefinitions)
		}

		spin.Stop()
		if user.CreditBalance == 0 {
			printZeroCreditsWarning()
		}
		tableMain.Render()
		if len(respJSON.Requests) > 0 {
			tableRequests.Render()
		}
	},
}

var incidentListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all past and current incidents",
	Long: `
List all past and current incidents.
`,
	Aliases:           []string{"ls"},
	Args:              cobra.NoArgs,
	DisableAutoGenTag: true,
	Run: func(cmd *cobra.Command, args []string) {
		util.VerifyAuthenticated()

		spin.Start()
		defer spin.Stop()
		spin.Suffix = colorFaint.Sprint(" loading incidents...")
		user, err := fetchUser()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		urlValues := url.Values{
			"period": []string{"all"},
		}
		match, err := regexp.MatchString(validCheckIdentPattern, incidentListFlagCheck)
		if err == nil && match {
			urlValues.Set("check", incidentListFlagCheck)
		}
		if incidentListFlagOpen && incidentListFlagResolved {
			spin.Stop()
			fmt.Println("Cannot use --open and --resolved flags together")
			os.Exit(1)
		}
		if incidentListFlagOpen {
			urlValues.Set("state", "open")
		}
		if incidentListFlagResolved {
			urlValues.Set("state", "resolved")
		}

		incidents, err := fetchIncidents(urlValues)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		var tableData [][]string
		for _, v := range incidents {
			var identSnippet, checkNameSnippet, stateSnippet, closedSnippet, responseCodesSnippet string
			identSnippet = colorBold.Sprint(v.Ident)
			if v.CheckName == "" {
				checkNameSnippet = colorFaint.Sprint("-")
			} else {
				checkNameSnippet = colorBold.Sprint(v.CheckName)
			}
			switch v.IncidentState {
			case incidentStateOpen:
				stateSnippet = color.YellowString(strings.ToUpper(v.IncidentState))
			case incidentStateResolved:
				stateSnippet = color.GreenString(strings.ToUpper(v.IncidentState))
			}
			if v.Closed == "" {
				closedSnippet = colorFaint.Sprint("-")
			} else {
				closedSnippet = v.Closed
			}
			responseCodesSnippet = colorFaint.Sprint(strings.Join(v.ResponseCodes, "\n"))

			tableRow := []string{
				identSnippet, v.CheckIdent, checkNameSnippet, util.Ellipsis(v.CheckResource, 50), stateSnippet, v.Opened, closedSnippet, util.OutputDurationWithDays(v.Duration), responseCodesSnippet,
			}
			tableData = append(tableData, tableRow)
		}

		columnDefinitions := []tableColumnDefinition{
			{
				Header:    "INCIDENT ID",
				Priority:  1,
				Alignment: tablewriter.ALIGN_LEFT,
			},
			{
				Header:    "CHECK ID",
				Priority:  1,
				Alignment: tablewriter.ALIGN_LEFT,
			},
			{
				Header:    "CHECK NAME",
				Priority:  2,
				Alignment: tablewriter.ALIGN_LEFT,
			},
			{
				Header:    "URL/HOST",
				Priority:  3,
				Alignment: tablewriter.ALIGN_LEFT,
			},
			{
				Header:    "STATE",
				Priority:  1,
				Alignment: tablewriter.ALIGN_LEFT,
			},
			{
				Header:    "OPENED",
				Priority:  2,
				Alignment: tablewriter.ALIGN_LEFT,
			},
			{
				Header:    "CLOSED",
				Priority:  2,
				Alignment: tablewriter.ALIGN_LEFT,
			},
			{
				Header:    "DURATION",
				Priority:  3,
				Alignment: tablewriter.ALIGN_LEFT,
			},
			{
				Header:    "RESPONSES",
				Priority:  3,
				Alignment: tablewriter.ALIGN_LEFT,
			},
		}

		table := composeTable(tableData, columnDefinitions)

		spin.Stop()
		if user.CreditBalance == 0 {
			printZeroCreditsWarning()
		}
		table.Render()
	},
}

var incidentUpdateCmd = &cobra.Command{
	Use:   "update",
	Short: "Provide incident with a note",
	Long: `
Provide incident with a note. This note would be visible on incident page.

This command is interactive and asks user for parameters that were not provided as flags.
`,
	Args:              cobra.ExactArgs(1),
	DisableAutoGenTag: true,
	Run: func(cmd *cobra.Command, args []string) {
		util.VerifyAuthenticated()

		// @todo implement
	},
}

func fetchIncidents(urlValues url.Values) ([]Incident, error) {
	var incidents []Incident
	respData, err := util.BinocsAPI("/incidents?"+urlValues.Encode(), http.MethodGet, []byte{})
	if err != nil {
		return incidents, err
	}
	incidents = make([]Incident, 0)
	decoder := json.NewDecoder(bytes.NewBuffer(respData))
	err = decoder.Decode(&incidents)
	if err != nil {
		return incidents, err
	}
	return incidents, nil
}
