package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"
)

const docOutputBase = "./docs"

func init() {
	rootCmd.AddCommand(docGenCmd)
}

var docGenCmd = &cobra.Command{
	Use:               "docgen",
	Hidden:            true,
	DisableAutoGenTag: true,
	Run: func(cmd *cobra.Command, args []string) {
		_ = os.RemoveAll(docOutputBase)
		fmt.Println("purged: " + docOutputBase)
		err := os.Mkdir(docOutputBase, 0755)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		err = doc.GenMarkdownTree(rootCmd, docOutputBase)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		fmt.Println("documentation generated successfully")
	},
}
