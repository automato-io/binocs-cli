package cmd

import (
	"os"

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
		spin.Suffix = " loading regions..."

		loadSupportedRegions()

		var tableData [][]string
		for _, v := range supportedRegions {
			tableRow := []string{
				v,
			}
			tableData = append(tableData, tableRow)
		}
		table := tablewriter.NewWriter(os.Stdout)
		table.SetAutoWrapText(false)
		table.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
		table.SetHeaderColor(tablewriter.Colors{tablewriter.Bold})
		table.SetHeader([]string{"REGIONS"})
		for _, v := range tableData {
			table.Append(v)
		}
		spin.Stop()
		table.Render()
	},
}
