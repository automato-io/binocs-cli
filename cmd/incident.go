package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// Incident comes from the API as a JSON
type Incident struct {
}

func init() {
	rootCmd.AddCommand(incidentCmd)
	incidentCmd.AddCommand(incidentInspectCmd)
}

var incidentCmd = &cobra.Command{
	Use:     "incident",
	Short:   "Manage incidents",
	Long:    `...`,
	Aliases: []string{"incidents"},
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("display incidents")
	},
}

var incidentInspectCmd = &cobra.Command{
	Use:     "inspect",
	Aliases: []string{"view", "show"},
	Run: func(cmd *cobra.Command, args []string) {

	},
}
