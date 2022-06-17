package cmd

import (
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(regionsCmd)
}

var regionsCmd = &cobra.Command{
	Use:   "regions",
	Short: "List supported regions",
	Long: `
List the regions Binocs makes requests from
`,
	DisableAutoGenTag: true,
	Run: func(cmd *cobra.Command, args []string) {
		spin.Start()
		defer spin.Stop()
		spin.Suffix = colorFaint.Sprint(" loading regions...")

		loadSupportedRegions()

		var tableData [][]string
		for _, v := range supportedRegions {
			tableRow := []string{v}
			tableData = append(tableData, tableRow)
		}

		columnDefinitions := []tableColumnDefinition{
			{
				Header:    "REGIONS",
				Priority:  1,
				Alignment: tablewriter.ALIGN_LEFT,
			},
		}

		table := composeTable(tableData, columnDefinitions)
		spin.Stop()
		table.Render()
	},
}
