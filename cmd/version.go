package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(versionCmd)
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of binocs",
	Long:  `All software has versions. This is binocs's`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("binocs v0.2")
	},
}
