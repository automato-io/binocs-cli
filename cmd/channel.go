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

	util "github.com/automato-io/binocs-cli/util"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
)

// Channel comes from the API as a JSON
type Channel struct {
	ID        int    `json:"id"`
	Ident     string `json:"ident"`
	Type      string `json:"type"`
	Alias     string `json:"alias"`
	Handle    string `json:"handle"`
	UsedCount int    `json:"used_count"`
	LastUsed  string `json:"last_used"`
}

// `channel ls` flags
var (
	channelListFlagCheck string
)

const (
	validChannelIdentPattern = `^[a-f0-9]{5}$`
)

func init() {
	rootCmd.AddCommand(channelCmd)
	// channelCmd.AddCommand(channelAddCmd)
	// channelCmd.AddCommand(channelAssociateCmd)
	// channelCmd.AddCommand(channelDeassociateCmd)
	channelCmd.AddCommand(channelInspectCmd)
	channelCmd.AddCommand(channelListCmd)
	// channelCmd.AddCommand(channelUpdateCmd)
}

var channelCmd = &cobra.Command{
	Use:     "channel",
	Short:   "Manage notification channels",
	Long:    ``,
	Aliases: []string{"channels"},
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			cmd.Run(channelListCmd, args)
		} else if len(args) == 1 && true { // @todo true ~ channel id validity regex
			cmd.Run(channelInspectCmd, args)
		} else {
			fmt.Println("Unsupported command/arguments combination, please see --help")
			os.Exit(1)
		}
	},
}

var channelInspectCmd = &cobra.Command{
	Use:   "inspect",
	Short: "view channel details",
	Long: `
View channel details and associated checks.
`,
	Aliases: []string{"view", "show", "info"},
	Args:    cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		spin.Start()
		spin.Suffix = " loading channel..."
		respData, err := util.BinocsAPI("/channels/"+args[0], http.MethodGet, []byte{})
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		var respJSON Channel
		err = json.Unmarshal(respData, &respJSON)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		// Table "main"
		lastUsed := ""
		if respJSON.UsedCount > 0 {
			lastUsed = `, last time at ` + respJSON.LastUsed
		}

		// @todo show ID field in check and incident detail as well
		tableMainChannelCellContent := `ID: ` + respJSON.Ident + `
Type: ` + respJSON.Type + `
Alias: ` + respJSON.Type + `
Handle: ` + respJSON.Handle + `
Used: ` + strconv.Itoa(respJSON.UsedCount) + ` times` + lastUsed + ``

		tableMain := tablewriter.NewWriter(os.Stdout)
		tableMain.SetHeader([]string{"CHANNEL"})
		tableMain.SetAutoWrapText(false)
		tableMain.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
		tableMain.SetColumnAlignment([]int{tablewriter.ALIGN_DEFAULT})
		tableMain.Append([]string{tableMainChannelCellContent})

		spin.Stop()
		tableMain.Render()
	},
}

var channelListCmd = &cobra.Command{
	Use:   "list",
	Short: "list all notification channels",
	Long: `
List all notification channels.
`,
	Aliases: []string{"ls"},
	Args:    cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		spin.Start()
		spin.Suffix = " loading channels..."

		urlValues := url.Values{
			"check": []string{""},
		}
		match, err := regexp.MatchString(validCheckIdentPattern, channelListFlagCheck)
		if err == nil && match == true {
			urlValues.Set("check", channelListFlagCheck)
		}
		channels, err := fetchChannels(urlValues)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		var tableData [][]string
		for _, v := range channels {
			tableRow := []string{
				v.Ident, v.Type, v.Alias, v.Handle, strconv.Itoa(v.UsedCount), v.LastUsed,
			}
			tableData = append(tableData, tableRow)
		}
		table := tablewriter.NewWriter(os.Stdout)
		table.SetAutoWrapText(false)
		table.SetHeader([]string{"ID", "TYPE", "ALIAS", "HANDLE", "USED", "LAST USED"})
		for _, v := range tableData {
			table.Append(v)
		}
		spin.Stop()
		table.Render()
	},
}

// var incidentUpdateCmd = &cobra.Command{
// 	Use:   "update",
// 	Short: "provide incident with a note",
// 	Long: `
// Provide incident with a note. This note would be visible on incident page.
// `,
// 	Args: cobra.ExactArgs(1),
// 	Run: func(cmd *cobra.Command, args []string) {
// 		// @todo implement
// 	},
// }

func fetchChannels(urlValues url.Values) ([]Channel, error) {
	var channels []Channel
	respData, err := util.BinocsAPI("/channels?"+urlValues.Encode(), http.MethodGet, []byte{})
	if err != nil {
		return channels, err
	}
	channels = make([]Channel, 0)
	decoder := json.NewDecoder(bytes.NewBuffer(respData))
	err = decoder.Decode(&channels)
	if err != nil {
		return channels, err
	}
	return channels, nil
}
