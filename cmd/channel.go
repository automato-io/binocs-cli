package cmd

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	util "github.com/automato-io/binocs-cli/util"
	"github.com/manifoldco/promptui"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
)

// Channel comes from the API as a JSON
type Channel struct {
	ID        int    `json:"id,omitempty"`
	Ident     string `json:"ident,omitempty"`
	Type      string `json:"type,omitempty"`
	Alias     string `json:"alias,omitempty"`
	Handle    string `json:"handle,omitempty"`
	UsedCount int    `json:"used_count,omitempty"`
	LastUsed  string `json:"last_used,omitempty"`
}

// ChannelAttachment struct is used to attach/detach a channel to/trom a check
type ChannelAttachment struct {
	NotificationType string `json:"notification_type"`
}

// `channel ls` flags
var (
	channelListFlagCheck string
)

// `channel add` flags
var (
	channelAddFlagAlias  string
	channelAddFlagHandle string
	channelAddFlagType   string
)

// `channel attach` flags
var (
	channelAttachFlagCheck string
	channelAttachFlagType  string
	channelAttachFlagAll   bool
)

// `channel detach` flags
var (
	channelDetachFlagCheck string
	channelDetachFlagType  string
	channelDetachFlagAll   bool
)

// `channel update` flags
var (
	channelUpdateFlagAlias  string
	channelUpdateFlagHandle string
)

const (
	validChannelIdentPattern     = `^[a-f0-9]{5}$`
	validAliasPattern            = `^[a-zA-Z0-9_\ \/\-\.]{0,25}$`
	validTypePattern             = `^(email|slack|telegram)$`
	validNotificationTypePattern = `^(http-code-change|status)$`
	channelTypeEmail             = "email"
	channelTypeTelegram          = "telegram"
	channelTypeSlack             = "slack"
)

var supportedNotificationTypes = []string{"http-code-change", "status"}

var validHandlePattern = map[string]string{
	"email":    `^(?:[a-z0-9!#$%&'*+/=?^_{|}~-]+(?:\.[a-z0-9!#$%&'*+/=?^_{|}~-]+)*|"(?:[\x01-\x08\x0b\x0c\x0e-\x1f\x21\x23-\x5b\x5d-\x7f]|\\[\x01-\x09\x0b\x0c\x0e-\x7f])*")@(?:(?:[a-z0-9](?:[a-z0-9-]*[a-z0-9])?\.)+[a-z0-9](?:[a-z0-9-]*[a-z0-9])?|\[(?:(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.){3}(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?|[a-z0-9-]*[a-z0-9]:(?:[\x01-\x08\x0b\x0c\x0e-\x1f\x21-\x5a\x53-\x7f]|\\[\x01-\x09\x0b\x0c\x0e-\x7f])+)\])$`,
	"slack":    `^https?:\/\/(www\.)?[-a-zA-Z0-9@:%._\+~#=]{1,256}\.[a-zA-Z0-9()]{1,6}\b([-a-zA-Z0-9()@:%_\+.~#?&//=]*)$`,
	"telegram": `^https?:\/\/(www\.)?[-a-zA-Z0-9@:%._\+~#=]{1,256}\.[a-zA-Z0-9()]{1,6}\b([-a-zA-Z0-9()@:%_\+.~#?&//=]*)$`,
}

func init() {
	rootCmd.AddCommand(channelCmd)
	channelCmd.AddCommand(channelAddCmd)
	channelCmd.AddCommand(channelAttachCmd)
	channelCmd.AddCommand(channelDetachCmd)
	channelCmd.AddCommand(channelDeleteCmd)
	channelCmd.AddCommand(channelInspectCmd)
	channelCmd.AddCommand(channelListCmd)
	channelCmd.AddCommand(channelUpdateCmd)

	channelAttachCmd.Flags().StringVarP(&channelAttachFlagCheck, "check", "c", "", "check identifier, using multiple comma-separated identifiers is supported")
	channelAttachCmd.Flags().StringVarP(&channelAttachFlagType, "type", "t", "", "notification type, \"status\" or \"http-code-change\" or both, defaults to \"http-code-change,status\"")
	channelAttachCmd.Flags().BoolVarP(&channelAttachFlagAll, "all", "a", false, "attach all checks to this channel")
	channelAttachCmd.Flags().SortFlags = false

	channelDetachCmd.Flags().StringVarP(&channelDetachFlagCheck, "check", "c", "", "check identifier, using multiple comma-separated identifiers is supported")
	channelDetachCmd.Flags().StringVarP(&channelDetachFlagType, "type", "t", "", "notification type, \"status\" or \"http-code-change\" or both, defaults to \"http-code-change,status\"")
	channelDetachCmd.Flags().BoolVarP(&channelDetachFlagAll, "all", "a", false, "detach all checks from this channel")
	channelDetachCmd.Flags().SortFlags = false

	channelAddCmd.Flags().StringVarP(&channelAddFlagType, "type", "t", "", "channel type (E-mail, Slack, Telegram)")
	channelAddCmd.Flags().StringVarP(&channelAddFlagHandle, "handle", "", "", "channel handle - e-mail address for E-mail, Slack URL for Slack")
	channelAddCmd.Flags().StringVarP(&channelAddFlagAlias, "alias", "", "", "channel alias - how we're gonna refer to it; optional")
	channelAddCmd.Flags().SortFlags = false

	channelListCmd.Flags().StringVarP(&channelListFlagCheck, "check", "c", "", "list only notification channels attached to a specific check")

	channelUpdateCmd.Flags().StringVarP(&channelUpdateFlagHandle, "handle", "", "", "channel handle - e-mail address for E-mail, Slack URL for Slack")
	channelUpdateCmd.Flags().StringVarP(&channelUpdateFlagAlias, "alias", "", "", "channel alias - how we're gonna refer to it; optional")
	channelUpdateCmd.Flags().SortFlags = false
}

var channelCmd = &cobra.Command{
	Use:   "channel",
	Short: "Manage notification channels",
	Long: `
Manage notification channels
`,
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

var channelAddCmd = &cobra.Command{
	Use:   "add",
	Short: "Add a new notification channel",
	Long: `
Add a new notification channel
`,
	Aliases: []string{"create"},
	Args:    cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		channelAddOrUpdate("add", "")
	},
}

var channelAttachCmd = &cobra.Command{
	Use:   "attach",
	Short: "Attach channel to one or more checks",
	Long: `
Attach channel to one or more checks, either for "status", "http-code-change" or both types of notifications
`,
	Aliases: []string{"att"},
	Args:    cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		var err error
		var match bool

		spin.Start()
		spin.Suffix = " attaching channel " + args[0]

		// validate channels ident
		channelRespData, err := util.BinocsAPI("/channels/"+args[0], http.MethodGet, []byte{})
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		var currentRespJSON Channel
		err = json.Unmarshal(channelRespData, &currentRespJSON)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		checkIdents := []string{}

		if channelAttachFlagAll {
			// get all checks
			checkRespData, err := util.BinocsAPI("/checks", http.MethodGet, []byte{})
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
			checksRespJSON := make([]Check, 0)
			decoder := json.NewDecoder(bytes.NewBuffer(checkRespData))
			err = decoder.Decode(&checksRespJSON)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
			for _, c := range checksRespJSON {
				checkIdents = append(checkIdents, c.Ident)
			}
		} else {
			// validate checks against pattern, single or slice, required
			if len(channelAttachFlagCheck) == 0 {
				fmt.Println("Set at least one check to attach to the channel")
				os.Exit(1)
			}
			checkIdents = strings.Split(channelAttachFlagCheck, ",")
			for _, c := range checkIdents {
				match, err = regexp.MatchString(validCheckIdentPattern, c)
				if err != nil {
					fmt.Println(err)
					os.Exit(1)
				} else if match == false {
					fmt.Println("Provided check identifier is invalid")
					os.Exit(1)
				}
			}
		}

		spin.Suffix = " attaching channel " + args[0] + " to " + strconv.Itoa(len(checkIdents)) + " checks"

		// validate types against pattern, single or slice, default all
		notificationTypes := []string{}
		if len(channelAttachFlagType) == 0 {
			notificationTypes = supportedNotificationTypes
		} else {
			notificationTypes = strings.Split(channelAttachFlagType, ",")
		}
		for _, nt := range notificationTypes {
			match, err = regexp.MatchString(validNotificationTypePattern, nt)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			} else if !match {
				fmt.Println("Provided notification type is invalid. Supported notification types: " + strings.Join(supportedNotificationTypes, ", "))
				os.Exit(1)
			}
		}

		for _, c := range checkIdents {
			for _, nt := range notificationTypes {
				postData, err := json.Marshal(ChannelAttachment{
					NotificationType: nt,
				})
				if err != nil {
					fmt.Println(err)
					os.Exit(1)
				}
				_, err = util.BinocsAPI("/channels/"+args[0]+"/check/"+c, http.MethodPost, postData)
				if err != nil {
					fmt.Println(err)
					os.Exit(1)
				}
			}
		}

		spin.Stop()
		fmt.Println("Successfully attached channel " + args[0] + " to " + strconv.Itoa(len(checkIdents)) + " checks")
	},
}

var channelDetachCmd = &cobra.Command{
	Use:   "detach",
	Short: "Detach channel from one or more checks",
	Long: `
Detach channel from one or more checks, either for "status", "http-code-change" or both types of notifications
`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		var err error
		var match bool

		spin.Start()
		spin.Suffix = " detaching channel " + args[0]

		// validate channels ident
		channelRespData, err := util.BinocsAPI("/channels/"+args[0], http.MethodGet, []byte{})
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		var currentRespJSON Channel
		err = json.Unmarshal(channelRespData, &currentRespJSON)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		checkIdents := []string{}

		if channelDetachFlagAll {
			// get all checks
			checkRespData, err := util.BinocsAPI("/checks", http.MethodGet, []byte{})
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
			checksRespJSON := make([]Check, 0)
			decoder := json.NewDecoder(bytes.NewBuffer(checkRespData))
			err = decoder.Decode(&checksRespJSON)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
			for _, c := range checksRespJSON {
				checkIdents = append(checkIdents, c.Ident)
			}
		} else {
			// validate checks against pattern, single or slice, required
			if len(channelDetachFlagCheck) == 0 {
				fmt.Println("Set at least one check to detach from the channel")
				os.Exit(1)
			}
			checkIdents = strings.Split(channelDetachFlagCheck, ",")
			for _, c := range checkIdents {
				match, err = regexp.MatchString(validCheckIdentPattern, c)
				if err != nil {
					fmt.Println(err)
					os.Exit(1)
				} else if !match {
					fmt.Println("Provided check identifier is invalid")
					os.Exit(1)
				}
			}
		}

		spin.Suffix = " detaching channel " + args[0] + " from " + strconv.Itoa(len(checkIdents)) + " checks"

		// validate types against pattern, single or slice, default all
		notificationTypes := []string{}
		if len(channelDetachFlagType) == 0 {
			notificationTypes = supportedNotificationTypes
		} else {
			notificationTypes = strings.Split(channelDetachFlagType, ",")
		}
		for _, nt := range notificationTypes {
			match, err = regexp.MatchString(validNotificationTypePattern, nt)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			} else if match == false {
				fmt.Println("Provided notification type is invalid. Supported notification types: " + strings.Join(supportedNotificationTypes, ", "))
				os.Exit(1)
			}
		}

		for _, c := range checkIdents {
			for _, nt := range notificationTypes {
				deleteData, err := json.Marshal(ChannelAttachment{
					NotificationType: nt,
				})
				if err != nil {
					fmt.Println(err)
					os.Exit(1)
				}
				_, err = util.BinocsAPI("/channels/"+args[0]+"/check/"+c, http.MethodDelete, deleteData)
				if err != nil {
					fmt.Println(err)
					os.Exit(1)
				}
			}
		}

		spin.Stop()
		fmt.Println("Successfully detached channel " + args[0] + " from " + strconv.Itoa(len(checkIdents)) + " checks")
	},
}

var channelDeleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete a notification channel",
	Long: `
Delete a notification channel.
`,
	Aliases: []string{"del", "rm"},
	Args:    cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		respData, err := util.BinocsAPI("/channels/"+args[0], http.MethodGet, []byte{})
		if err != nil {
			// @todo verbose error
			fmt.Println(err)
			os.Exit(1)
		}
		var respJSON Channel
		err = json.Unmarshal(respData, &respJSON)
		if err != nil {
			// @todo verbose error
			fmt.Println(err)
			os.Exit(1)
		}

		prompt := promptui.Prompt{
			Label:     "Delete " + respJSON.Type + " notification channel " + respJSON.Alias + " (" + respJSON.Handle + ")",
			IsConfirm: true,
		}
		_, err = prompt.Run()
		if err != nil {
			fmt.Println("Aborting")
			os.Exit(0)
		}
		_, err = util.BinocsAPI("/channels/"+args[0], http.MethodDelete, []byte{})
		if err != nil {
			// @todo verbose error
			fmt.Println(err)
			os.Exit(1)
		}
		tpl := `Channel successfully deleted
`
		fmt.Print(tpl)
	},
}

var channelInspectCmd = &cobra.Command{
	Use:   "inspect",
	Short: "View channel details",
	Long: `
View channel details and attached checks.
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
Alias: ` + respJSON.Alias + `
Handle: ` + respJSON.Handle + `
Used ` + strconv.Itoa(respJSON.UsedCount) + `x` + lastUsed + ``

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
	Short: "List all notification channels",
	Long: `
List all notification channels.
`,
	Aliases: []string{"ls"},
	Args:    cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		spin.Start()
		spin.Suffix = " loading channels..."

		urlValues := url.Values{}
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
				v.Ident, v.Type, v.Alias, util.Ellipsis(v.Handle, 50), strconv.Itoa(v.UsedCount), v.LastUsed,
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

var channelUpdateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update existing notification channel",
	Long: `
Update existing notification channel.
`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		channelAddOrUpdate("update", args[0])
	},
}

// mode = add|update
func channelAddOrUpdate(mode string, channelIdent string) {
	if mode != "add" && mode != "update" {
		fmt.Println("Unknown mode: " + mode)
		os.Exit(1)
	}

	var err error
	var match bool
	var tpl string

	var (
		flagAlias  string
		flagHandle string
		flagType   string
	)

	switch mode {
	case "add":
		flagAlias = channelAddFlagAlias
		flagHandle = channelAddFlagHandle
		flagType = channelAddFlagType
	case "update":
		flagAlias = channelUpdateFlagAlias
		flagHandle = channelUpdateFlagHandle
	}

	if mode == "update" {
		// pass
	} else {
		// check if Type is one from a set, empty not allowed
		match, err = regexp.MatchString(validTypePattern, flagType)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		} else if match == false {
			prompt := promptui.Select{
				Label: "Type",
				Items: []string{channelTypeEmail, channelTypeSlack, channelTypeTelegram},
			}
			_, flagType, err = prompt.Run()
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
		}
	}

	if mode == "update" && flagAlias == "" {
		// pass
	} else {
		// check if Alias is alphanum, space & normal chars, empty OK
		match, err = regexp.MatchString(validAliasPattern, flagAlias)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		} else if match == false || flagAlias == "" {
			validate := func(input string) error {
				match, err = regexp.MatchString(validAliasPattern, input)
				if err != nil {
					return errors.New("Invalid input")
				} else if match == false {
					return errors.New("Invalid input value")
				}
				return nil
			}
			prompt := promptui.Prompt{
				Label:    "Channel alias (optional)",
				Validate: validate,
			}
			flagAlias, err = prompt.Run()
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
		}
	}

	if mode == "update" && flagHandle == "" {
		// pass
	} else if mode == "update" {
		currentRespData, err := util.BinocsAPI("/channels/"+channelIdent, http.MethodGet, []byte{})
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		var currentRespJSON Channel
		err = json.Unmarshal(currentRespData, &currentRespJSON)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		match, err = regexp.MatchString(validHandlePattern[currentRespJSON.Type], flagHandle)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		} else if match == false {
			validate := func(input string) error {
				match, err = regexp.MatchString(validHandlePattern[currentRespJSON.Type], input)
				if err != nil {
					return errors.New("Invalid input")
				} else if match == false {
					return errors.New("Invalid input value")
				}
				return nil
			}
			prompt := promptui.Prompt{
				Label:    "Enter a valid " + currentRespJSON.Type + " handle",
				Validate: validate,
			}
			flagHandle, err = prompt.Run()
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
		}
	} else {
		match, err = regexp.MatchString(validHandlePattern[flagType], flagHandle)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		} else if match == false {
			if flagType == channelTypeSlack {
				slackIntegrationToken, err := requestSlackIntegrationToken()
				if err != nil {
					fmt.Println(err)
					os.Exit(1)
				}
				slackScope := "incoming-webhook"
				slackRedirectURI := "https://binocs.sh/integration/slack/callback"
				slackClientID := "1145502794960.1106931893399"
				slackAuthorizeURL := "https://slack.com/oauth/v2/authorize?scope=" + url.QueryEscape(slackScope) + "&client_id=" + url.QueryEscape(slackClientID) + "&redirect_uri=" + url.QueryEscape(slackRedirectURI) + "&state=" + slackIntegrationToken.Token
				fmt.Println("Visit the following URL to choose where we should send your notifications:")
				fmt.Println(slackAuthorizeURL)
				spin.Start()
				spin.Suffix = " waiting for your action ..."
				for {
					pollResult, err := pollSlackIntegrationStatus(slackIntegrationToken.Token)
					if err != nil {
						fmt.Println(err)
						os.Exit(1)
					}
					if pollResult.Updated != "nil" {
						flagHandle = pollResult.IncomingWebhookURL
						break
					}
					time.Sleep(3 * time.Second)
				}
				spin.Stop()
			} else {
				validate := func(input string) error {
					match, err = regexp.MatchString(validHandlePattern[flagType], input)
					if err != nil {
						return errors.New("Invalid input")
					} else if match == false {
						return errors.New("Invalid input value")
					}
					return nil
				}
				prompt := promptui.Prompt{
					Label:    "Enter a valid " + flagType + " handle",
					Validate: validate,
				}
				flagHandle, err = prompt.Run()
				if err != nil {
					fmt.Println(err)
					os.Exit(1)
				}
			}
		}
	}

	// all clear, we can call the API and confirm adding new channel
	channel := Channel{
		Alias:  flagAlias,
		Handle: flagHandle,
		Type:   flagType,
	}
	// @hack channel flags
	postData, err := json.Marshal(channel)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	if bytes.Equal(postData, []byte("{}")) {
		fmt.Printf("provide at least one parameter that you want to update\n")
		os.Exit(1)
	}
	var reqURL, reqMethod string
	if mode == "add" {
		reqURL = "/channels"
		reqMethod = http.MethodPost
	}
	if mode == "update" {
		reqURL = "/channels/" + channelIdent
		reqMethod = http.MethodPut
	}
	// @todo verbose print postData
	respData, err := util.BinocsAPI(reqURL, reqMethod, postData)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	// @toto verbose print respData
	err = json.Unmarshal(respData, &channel)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	if channel.ID > 0 {
		var channelDescription string
		if len(channel.Alias) > 0 {
			channelDescription = channel.Alias + " (" + channel.Handle + ")"
		} else {
			channelDescription = channel.Handle
		}
		if mode == "add" {
			tpl = "[" + channel.Ident + "] " + channelDescription + ` added successfully
`
		}
		if mode == "update" {
			tpl = "[" + channel.Ident + "] " + channelDescription + ` updated successfully
`
		}
	} else {
		if mode == "add" {
			fmt.Println("Error adding channel")
			os.Exit(1)
		}
		if mode == "update" {
			fmt.Println("Error updating channel")
			os.Exit(1)
		}
	}
	fmt.Print(tpl)
}

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

// SlackIntegrationToken struct
type SlackIntegrationToken struct {
	Token string `json:"token"`
}

func requestSlackIntegrationToken() (SlackIntegrationToken, error) {
	var token SlackIntegrationToken
	respData, err := util.BinocsAPI("/integration/slack/request-integration-token", http.MethodPost, []byte{})
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	err = json.Unmarshal(respData, &token)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	return token, nil
}

// SlackIntegrationStatus struct
type SlackIntegrationStatus struct {
	Token              string `json:"token"`
	IncomingWebhookURL string `json:"incoming_webhook_url"`
	Updated            string `json:"updated,omitempty"`
}

func pollSlackIntegrationStatus(token string) (SlackIntegrationStatus, error) {
	var status SlackIntegrationStatus
	respData, err := util.BinocsAPI("/integration/slack/status/"+token, http.MethodGet, []byte{})
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	err = json.Unmarshal(respData, &status)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	return status, nil
}
