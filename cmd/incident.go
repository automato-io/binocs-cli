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
	"time"

	util "github.com/automato-io/binocs-cli/util"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
)

// Incident comes from the API as a JSON
type Incident struct {
	ID            int       `json:"id"`
	Ident         string    `json:"ident"`
	CheckID       int       `json:"check_id"`
	CheckIdent    string    `json:"check_ident"`
	IncidentNote  string    `json:"incident_note"`
	IncidentState string    `json:"incident_state"`
	CheckName     string    `json:"check_name"`
	CheckURL      string    `json:"check_url"`
	Opened        string    `json:"opened"`
	Closed        string    `json:"closed"`
	Duration      string    `json:"duration"`
	ResponseCodes []string  `json:"response_codes"`
	Requests      []Request `json:"requests"`
}

// Request struct
type Request struct {
	Region             string    `json:"region"`
	Status             int       `json:"status"`
	RequestURL         string    `json:"request_url"`
	RequestMethod      string    `json:"request_method"`
	ResponseStatusCode string    `json:"response_status"`
	Timings            Timings   `json:"timings"`
	Timestamp          time.Time `json:"timestamp"`
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
	incidentListFlagCheck string
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
	incidentListCmd.Flags().StringVarP(&incidentListFlagCheck, "check", "c", "", "list only incidents of this check")

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
		spin.Start()
		spin.Suffix = " loading incident..."
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

		// Table "main"

		tableMainIncidentCellContent := `Check ID: ` + respJSON.CheckIdent + `
Check: ` + respJSON.CheckName + ` 
URL: ` + respJSON.CheckURL + `
Incident State: ` + respJSON.IncidentState + `
Response Codes: ` + strings.Join(respJSON.ResponseCodes, "\n") + `

Opened: ` + respJSON.Opened + `
Closed: ` + respJSON.Closed + `
Duration: ` + util.OutputDurationWithDays(respJSON.Duration)

		tableMainNotesCellContent := respJSON.IncidentNote
		if tableMainNotesCellContent == "" {
			tableMainNotesCellContent = "-"
		}

		tableMain := tablewriter.NewWriter(os.Stdout)
		tableMain.SetHeader([]string{"INCIDENT", "NOTES"})
		tableMain.SetAutoWrapText(false)
		tableMain.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
		tableMain.SetColumnAlignment([]int{tablewriter.ALIGN_DEFAULT, tablewriter.ALIGN_DEFAULT})
		tableMain.Append([]string{tableMainIncidentCellContent, tableMainNotesCellContent})

		// Table "requests"

		tableRequests := tablewriter.NewWriter(os.Stdout)
		if len(respJSON.Requests) > 0 {
			tz := respJSON.Requests[0].Timestamp.Format("-07:00")
			tableRequests.SetHeader([]string{"CHECKED AT (" + tz + ")", "CHECKED FROM", "RESPONSE CODE", "RESPONSE TIME", "DNS LOOKUP", "CONNECTION", "TLS", "WAITING", "TRANSFER"})
			tableRequests.SetAutoWrapText(false)
			tableRequests.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
			tableRequests.SetColumnAlignment([]int{tablewriter.ALIGN_LEFT, tablewriter.ALIGN_LEFT, tablewriter.ALIGN_LEFT, tablewriter.ALIGN_RIGHT, tablewriter.ALIGN_RIGHT, tablewriter.ALIGN_RIGHT, tablewriter.ALIGN_RIGHT, tablewriter.ALIGN_RIGHT, tablewriter.ALIGN_RIGHT})
			for _, request := range respJSON.Requests {
				if strings.Contains(request.Timestamp.String(), "0001") {
					placeholder := "~"
					shortcut := request.RequestURL + " requests"
					shortcut = strings.Repeat(placeholder, 2) + " " + shortcut
					if len(shortcut) < 18 {
						shortcut = shortcut + " " + strings.Repeat(placeholder, 19-len(shortcut)-1)
					}
					tableRequests.Append([]string{shortcut, strings.Repeat(placeholder, 7), request.ResponseStatusCode, strings.Repeat(placeholder, 7), strings.Repeat(placeholder, 7), strings.Repeat(placeholder, 7), strings.Repeat(placeholder, 7), strings.Repeat(placeholder, 7), strings.Repeat(placeholder, 7)})
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
					tableRequests.Append([]string{request.Timestamp.Format("2006-01-02 15:04:05"), request.Region, request.ResponseStatusCode, responseTime, timingsDNSLookup, timingsConnection, timingsTLS, timingsWait, timingsTransfer})
				}
			}
		}

		spin.Stop()
		if user.CreditBalance == 0 {
			printZeroCreditsWarning()
		}
		tableMain.Render()
		if len(respJSON.Requests) > 0 {
			fmt.Println()
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
		spin.Start()
		spin.Suffix = " loading incidents..."
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
		incidents, err := fetchIncidents(urlValues)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		var tableData [][]string
		for _, v := range incidents {
			// Mon Jan 2 15:04:05 -0700 MST 2006
			opened, err := time.Parse("2006-01-02 15:04:05 -0700", v.Opened)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
			tableRow := []string{
				v.Ident, v.CheckName, util.Ellipsis(v.CheckURL, 50), v.IncidentState, opened.Format("2006-01-02 15:04:05"), v.Closed, util.OutputDurationWithDays(v.Duration), strings.Join(v.ResponseCodes, "\n"), v.IncidentNote,
			}
			tableData = append(tableData, tableRow)
		}
		table := tablewriter.NewWriter(os.Stdout)
		table.SetAutoWrapText(false)
		table.SetHeader([]string{"ID", "CHECK", "URL", "STATE", "OPENED", "CLOSED", "DURATION", "RESPONSE CODES", "NOTE"})
		table.SetColumnAlignment([]int{tablewriter.ALIGN_DEFAULT, tablewriter.ALIGN_DEFAULT, tablewriter.ALIGN_DEFAULT, tablewriter.ALIGN_DEFAULT, tablewriter.ALIGN_DEFAULT, tablewriter.ALIGN_DEFAULT, tablewriter.ALIGN_DEFAULT, tablewriter.ALIGN_LEFT, tablewriter.ALIGN_DEFAULT})
		for _, v := range tableData {
			table.Append(v)
		}
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
`,
	Args:              cobra.ExactArgs(1),
	DisableAutoGenTag: true,
	Run: func(cmd *cobra.Command, args []string) {
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
