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
	ID            int    `json:"id"`
	Ident         string `json:"ident"`
	CheckID       int    `json:"check_id"`
	IncidentNote  string `json:"incident_note"`
	IncidentState string `json:"incident_state"`
	CheckName     string `json:"check_name"`
	CheckURL      string `json:"check_url"`
	Opened        string `json:"opened"`
	Closed        string `json:"closed"`
	Duration      string `json:"duration"`
	ResponseCodes string `json:"response_codes"`
}

// `incident update` flags
var (
	incidentUpdateFlagNote string
)

func init() {
	rootCmd.AddCommand(incidentCmd)
	incidentCmd.AddCommand(incidentInspectCmd)
	incidentCmd.AddCommand(incidentListCmd)
	incidentCmd.AddCommand(incidentUpdateCmd)

	incidentUpdateCmd.Flags().StringVarP(&incidentUpdateFlagNote, "note", "n", "", "Incident note")
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
	Use:     "inspect",
	Short:   "View info about incident",
	Aliases: []string{"view", "show", "info"},
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return fmt.Errorf("please provide the identifier of the incident you want to inspect")
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		spin.Start()
		spin.Stop()
	},
}

var incidentListCmd = &cobra.Command{
	Use:     "list",
	Short:   "View info about incident",
	Aliases: []string{"ls"},
	Args: func(cmd *cobra.Command, args []string) error {
		return nil
	},
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
