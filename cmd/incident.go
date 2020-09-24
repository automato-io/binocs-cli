package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"

	util "github.com/automato-io/binocs-cli/util"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
)

// Incident comes from the API as a JSON
type Incident struct {
	ID            int    `json:"id"`
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

func init() {
	rootCmd.AddCommand(incidentCmd)
	incidentCmd.AddCommand(incidentInspectCmd)
}

var incidentCmd = &cobra.Command{
	Use:     "incident",
	Short:   "Manage incidents",
	Long:    ``,
	Aliases: []string{"incidents"},
	Run: func(cmd *cobra.Command, args []string) {
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
			tableRow := []string{
				strconv.Itoa(v.ID), v.CheckName, v.CheckURL, v.IncidentState, v.Opened, v.Closed, util.OutputDurationWithDays(v.Duration), v.ResponseCodes, v.IncidentNote,
			}
			tableData = append(tableData, tableRow)
		}
		table := tablewriter.NewWriter(os.Stdout)
		table.SetAutoWrapText(false)
		table.SetHeader([]string{"ID", "CHECK", "URL", "STATE", "OPENED", "CLOSED", "DURATION", "RESPONSE CODES", "NOTE"})
		for _, v := range tableData {
			table.Append(v)
		}
		table.Render()
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
	},
}
