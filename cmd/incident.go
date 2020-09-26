package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
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
	IncidentNote  string    `json:"incident_note"`
	IncidentState string    `json:"incident_state"`
	CheckName     string    `json:"check_name"`
	CheckURL      string    `json:"check_url"`
	Opened        string    `json:"opened"`
	Closed        string    `json:"closed"`
	Duration      string    `json:"duration"`
	ResponseCodes string    `json:"response_codes"`
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
	DSNLookup  float64 `json:"dns_lookup"`
	Connection float64 `json:"connection"`
	TLS        float64 `json:"tls"`
	Wait       float64 `json:"wait"`
	Transfer   float64 `json:"transfer"`
}

// `incident ls` flags
var (
	incidentListCheckNote string
)

// `incident update` flags
var (
	incidentUpdateFlagNote string
)

func init() {
	rootCmd.AddCommand(incidentCmd)
	incidentCmd.AddCommand(incidentInspectCmd)
	incidentCmd.AddCommand(incidentListCmd)
	incidentCmd.AddCommand(incidentUpdateCmd)

	// @todo implement
	incidentListCmd.Flags().StringVarP(&incidentListCheckNote, "check", "c", "", "list only incidents of this check")

	incidentUpdateCmd.Flags().StringVarP(&incidentUpdateFlagNote, "note", "n", "", "incident note")
}

var incidentCmd = &cobra.Command{
	Use:     "incident",
	Short:   "Manage incidents",
	Long:    ``,
	Aliases: []string{"incidents"},
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			cmd.Run(incidentListCmd, args)
		} else if len(args) == 1 && true { // @todo true ~ incident id validity regex
			cmd.Run(incidentInspectCmd, args)
		} else {
			fmt.Println("Unsupported command/arguments combination, please see --help")
			os.Exit(1)
		}
	},
}

var incidentInspectCmd = &cobra.Command{
	Use:   "inspect",
	Short: "view incident details",
	Long: `
View incident details, notes and associated requests.
`,
	Aliases: []string{"view", "show", "info"},
	Args:    cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		spin.Start()
		spin.Suffix = " loading incident..."
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

		tableMainIncidentCellContent := `State: ` + respJSON.IncidentState + `
URL: ` + respJSON.CheckURL + `
Response Codes: ` + respJSON.ResponseCodes + `

Opened: ` + respJSON.Opened + `
Closed: ` + respJSON.Closed + `
Duration: ` + respJSON.Duration

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
			tz := respJSON.Requests[0].Timestamp.Format("-07:00 MST")
			tableRequests.SetHeader([]string{"CHECKED (" + tz + ")", "FROM", "RESPONSE CODE", "RESPONSE TIME", "DNS LOOKUP", "CONNECTION", "TLS", "WAITING", "TRANSFER"})
			tableRequests.SetAutoWrapText(false)
			tableRequests.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
			tableRequests.SetColumnAlignment([]int{tablewriter.ALIGN_LEFT})
			for _, request := range respJSON.Requests {
				responseTime := fmt.Sprintf("%.3f s", request.Timings.DSNLookup+request.Timings.Connection+request.Timings.TLS+request.Timings.Wait+request.Timings.Transfer)
				timingsDNSLookup := fmt.Sprintf("%.3f s", request.Timings.DSNLookup)
				timingsConnection := fmt.Sprintf("%.3f s", request.Timings.Connection)
				timingsTLS := fmt.Sprintf("%.3f s", request.Timings.DSNLookup)
				timingsWait := fmt.Sprintf("%.3f s", request.Timings.DSNLookup)
				timingsTransfer := fmt.Sprintf("%.3f s", request.Timings.DSNLookup)
				tableRequests.Append([]string{request.Timestamp.Format("2006-01-02 15:04:05"), request.Region, request.ResponseStatusCode, responseTime, timingsDNSLookup, timingsConnection, timingsTLS, timingsWait, timingsTransfer})
			}
		}

		spin.Stop()
		tableMain.Render()
		if len(respJSON.Requests) > 0 {
			fmt.Println()
			tableRequests.Render()
		}
	},
}

var incidentListCmd = &cobra.Command{
	Use:   "list",
	Short: "list all past incidents",
	Long: `
List all past incidents.
`,
	Aliases: []string{"ls"},
	Args:    cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		spin.Start()
		spin.Suffix = " loading incidents..."

		respData, err := util.BinocsAPI("/incidents", http.MethodGet, []byte{})
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		respJSON := make([]Incident, 0)
		decoder := json.NewDecoder(bytes.NewBuffer(respData))
		err = decoder.Decode(&respJSON)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		var tableData [][]string
		for _, v := range respJSON {
			// Mon Jan 2 15:04:05 -0700 MST 2006
			opened, err := time.Parse("2006-01-02 15:04:05 -0700 MST", v.Opened)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
			tableRow := []string{
				v.Ident, v.CheckName, v.CheckURL, v.IncidentState, opened.Format("2006-01-02 15:04:05"), v.Closed, util.OutputDurationWithDays(v.Duration), v.ResponseCodes, v.IncidentNote,
			}
			tableData = append(tableData, tableRow)
		}
		table := tablewriter.NewWriter(os.Stdout)
		table.SetAutoWrapText(false)
		table.SetHeader([]string{"ID", "CHECK", "URL", "STATE", "OPENED", "CLOSED", "DURATION", "RESPONSE CODES", "NOTE"})
		for _, v := range tableData {
			table.Append(v)
		}
		spin.Stop()
		table.Render()
	},
}

var incidentUpdateCmd = &cobra.Command{
	Use: "update",
	// Short:   "View info about incident",
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return fmt.Errorf("please provide the identifier of the incident you want to inspect")
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {

	},
}
