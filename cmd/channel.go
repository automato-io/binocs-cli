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

	"github.com/AlecAivazis/survey/v2"
	util "github.com/automato-io/binocs-cli/util"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
)

// Channel comes from the API as a JSON
type Channel struct {
	ID        int      `json:"id,omitempty"`
	Ident     string   `json:"ident,omitempty"`
	Type      string   `json:"type,omitempty"`
	Alias     string   `json:"alias"`
	Handle    string   `json:"handle,omitempty"`
	UsedCount int      `json:"used_count,omitempty"`
	LastUsed  string   `json:"last_used,omitempty"`
	Verified  string   `json:"verified,omitempty"`
	Checks    []string `json:"checks,omitempty"`
}

// Identity method returns "Type - Alias (handle)" or "handle"
func (ch Channel) Identity() string {
	if len(ch.Alias) > 0 {
		return ch.Alias + " (" + ch.Handle + ")"
	}
	return ch.Handle
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
	channelAddFlagAttach []string
)

// `channel attach` flags
var (
	channelAttachFlagCheck []string
	channelAttachFlagAll   bool
)

// `channel detach` flags
var (
	channelDetachFlagCheck []string
	channelDetachFlagAll   bool
)

// `channel update` flags
var (
	channelUpdateFlagAlias  string
	channelUpdateFlagAttach []string
)

const (
	validChannelIdentPattern      = `^[a-f0-9]{5}$`
	validAliasPattern             = `^[a-zA-Z0-9_\s\/\-\.]{0,25}$`
	validTypePattern              = `^(email|slack|telegram)$`
	validChannelsIdentListPattern = `^(all|([a-f0-9]{5})(,[a-f0-9]{5})*)$`
	validNotificationTypePattern  = `^(response-change|status)$`
	channelTypeEmail              = "email"
	channelTypeTelegram           = "telegram"
	channelTypeSlack              = "slack"
)

var validHandlePattern = map[string]string{
	"email": `^(?:[a-z0-9!#$%&'*+/=?^_{|}~-]+(?:\.[a-z0-9!#$%&'*+/=?^_{|}~-]+)*|"(?:[\x01-\x08\x0b\x0c\x0e-\x1f\x21\x23-\x5b\x5d-\x7f]|\\[\x01-\x09\x0b\x0c\x0e-\x7f])*")@(?:(?:[a-z0-9](?:[a-z0-9-]*[a-z0-9])?\.)+[a-z0-9](?:[a-z0-9-]*[a-z0-9])?|\[(?:(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.){3}(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?|[a-z0-9-]*[a-z0-9]:(?:[\x01-\x08\x0b\x0c\x0e-\x1f\x21-\x5a\x53-\x7f]|\\[\x01-\x09\x0b\x0c\x0e-\x7f])+)\])$`,
}

func init() {
	rootCmd.AddCommand(channelsCmd)

	rootCmd.AddCommand(channelCmd)

	channelCmd.AddCommand(channelAddCmd)
	channelCmd.AddCommand(channelAttachCmd)
	channelCmd.AddCommand(channelDetachCmd)
	channelCmd.AddCommand(channelDeleteCmd)
	channelCmd.AddCommand(channelInspectCmd)
	channelCmd.AddCommand(channelListCmd)
	channelCmd.AddCommand(channelUpdateCmd)

	channelAttachCmd.Flags().StringSliceVarP(&channelAttachFlagCheck, "check", "c", []string{}, "check identifiers")
	channelAttachCmd.Flags().BoolVarP(&channelAttachFlagAll, "all", "a", false, "attach all checks to this channel")
	channelAttachCmd.Flags().SortFlags = false

	channelDetachCmd.Flags().StringSliceVarP(&channelDetachFlagCheck, "check", "c", []string{}, "check identifiers")
	channelDetachCmd.Flags().BoolVarP(&channelDetachFlagAll, "all", "a", false, "detach all checks from this channel")
	channelDetachCmd.Flags().SortFlags = false

	channelAddCmd.Flags().StringVarP(&channelAddFlagType, "type", "t", "", "channel type (E-mail, Slack, Telegram)")
	channelAddCmd.Flags().StringVar(&channelAddFlagHandle, "handle", "", "channel handle - an address for \"E-mail\" channel type; handles for Slack and Telegram will be obtained programmatically")
	channelAddCmd.Flags().StringVar(&channelAddFlagAlias, "alias", "", "channel alias (optional)")
	channelAddCmd.Flags().StringSliceVar(&channelAddFlagAttach, "attach", []string{}, "checks to attach to this channel (optional); can be either \"all\", or one or more check identifiers")
	channelAddCmd.Flags().SortFlags = false

	channelsCmd.Flags().StringVarP(&channelListFlagCheck, "check", "c", "", "list only notification channels attached to a specific check")
	channelListCmd.Flags().StringVarP(&channelListFlagCheck, "check", "c", "", "list only notification channels attached to a specific check")

	channelUpdateCmd.Flags().StringVar(&channelUpdateFlagAlias, "alias", "", "channel alias (optional)")
	channelUpdateCmd.Flags().StringSliceVar(&channelUpdateFlagAttach, "attach", []string{}, "checks to attach to this channel (optional); can be either \"all\", or one or more check identifiers")
	channelUpdateCmd.Flags().SortFlags = false
}

var channelCmd = &cobra.Command{
	Use:               "channel",
	Short:             "Manage notification channels",
	DisableAutoGenTag: true,
}

var channelsCmd = &cobra.Command{
	Use:               "channels",
	Args:              cobra.NoArgs,
	Short:             channelListCmd.Short,
	Long:              channelListCmd.Long,
	DisableAutoGenTag: true,
	Run: func(cmd *cobra.Command, args []string) {
		channelListCmd.Run(cmd, args)
	},
}

var channelAddCmd = &cobra.Command{
	Use:   "add",
	Short: "Add a new notifications channel",
	Long: `
Add a new notifications channel
`,
	Aliases:           []string{"create"},
	Args:              cobra.NoArgs,
	DisableAutoGenTag: true,
	Run: func(cmd *cobra.Command, args []string) {
		channelAddOrUpdate("add", "")
	},
}

var channelAttachCmd = &cobra.Command{
	Use:   "attach",
	Short: "Attach channel to check(s)",
	Long: `
Attach channel to check(s)
`,
	Aliases:           []string{"att"},
	Args:              cobra.ExactArgs(1),
	DisableAutoGenTag: true,
	Run: func(cmd *cobra.Command, args []string) {
		var err error
		var match bool

		spin.Start()
		spin.Suffix = " loading channel " + args[0]
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
			checks, err := fetchChecks(url.Values{})
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
			for _, c := range checks {
				checkIdents = append(checkIdents, c.Ident)
			}
		} else {
			// validate checks against pattern, single or slice, required
			if len(channelAttachFlagCheck) == 0 {
				fmt.Println("Set at least one check to attach to the channel")
				os.Exit(1)
			}
			checkIdents = channelAttachFlagCheck
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
		spin.Stop()
		prompt := &survey.Confirm{
			Message: "Attach " + currentRespJSON.Type + " notification channel " + currentRespJSON.Identity() + " to " + fmt.Sprintf("%d", len(checkIdents)) + " checks?",
		}
		var yes bool
		err = survey.AskOne(prompt, &yes)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		if yes {
			spin.Start()
			spin.Suffix = " attaching channel " + args[0] + " to " + strconv.Itoa(len(checkIdents)) + " checks"
			for _, c := range checkIdents {
				postData, err := json.Marshal(ChannelAttachment{})
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
			spin.Stop()
			fmt.Println("Successfully attached channel " + args[0] + " to " + strconv.Itoa(len(checkIdents)) + " checks")
		} else {
			fmt.Println("OK, skipping")
		}
	},
}

var channelDetachCmd = &cobra.Command{
	Use:   "detach",
	Short: "Detach channel from check(s)",
	Long: `
Detach channel from check(s)
`,
	Args:              cobra.ExactArgs(1),
	DisableAutoGenTag: true,
	Run: func(cmd *cobra.Command, args []string) {
		var err error
		var match bool

		spin.Start()
		spin.Suffix = " loading channel " + args[0]
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
			checks, err := fetchChecks(url.Values{})
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
			for _, c := range checks {
				for _, cc := range c.Channels {
					if cc == currentRespJSON.Ident {
						checkIdents = append(checkIdents, c.Ident)
					}
				}
			}
		} else {
			// validate checks against pattern, single or slice, required
			if len(channelDetachFlagCheck) == 0 {
				fmt.Println("Set at least one check to detach from the channel")
				os.Exit(1)
			}
			checkIdents = channelDetachFlagCheck
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
		spin.Stop()
		prompt := &survey.Confirm{
			Message: "Detach " + currentRespJSON.Type + " notification channel " + currentRespJSON.Identity() + " from " + fmt.Sprintf("%d", len(checkIdents)) + " checks?",
		}
		var yes bool
		err = survey.AskOne(prompt, &yes)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		if yes {
			spin.Start()
			spin.Suffix = " detaching channel " + args[0] + " from " + strconv.Itoa(len(checkIdents)) + " checks"
			for _, c := range checkIdents {
				deleteData, err := json.Marshal(ChannelAttachment{})
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
			spin.Stop()
			fmt.Println("Successfully detached channel " + args[0] + " from " + strconv.Itoa(len(checkIdents)) + " checks")
		} else {
			fmt.Println("OK, skipping")
		}
	},
}

var channelDeleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete notification channel(s)",
	Long: `
Delete notification channel(s).
`,
	Aliases:           []string{"del", "rm"},
	Args:              cobra.MatchAll(),
	DisableAutoGenTag: true,
	Run: func(cmd *cobra.Command, args []string) {
		for _, arg := range args {
			respData, err := util.BinocsAPI("/channels/"+arg, http.MethodGet, []byte{})
			if err != nil {
				fmt.Println("Error loading channel " + arg)
				continue
			}
			var respJSON Channel
			err = json.Unmarshal(respData, &respJSON)
			if err != nil {
				fmt.Println("Invalid response from binocs.sh")
				continue
			}
			prompt := &survey.Confirm{
				Message: "Delete " + respJSON.Type + " notification channel " + respJSON.Alias + " (" + respJSON.Handle + ")?",
			}
			var yes bool
			err = survey.AskOne(prompt, &yes)
			if err != nil {
				continue
			}
			if yes {
				_, err = util.BinocsAPI("/channels/"+arg, http.MethodDelete, []byte{})
				if err != nil {
					fmt.Println("Error deleting channel " + arg)
					continue
				} else {
					fmt.Println("Channel successfully deleted")
				}
			} else {
				fmt.Println("OK, skipping")
			}
		}
	},
}

var channelInspectCmd = &cobra.Command{
	Use:   "inspect",
	Short: "View channel details",
	Long: `
View channel details and attached checks.
`,
	Aliases:           []string{"view", "show", "info"},
	Args:              cobra.ExactArgs(1),
	DisableAutoGenTag: true,
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

		var handle, used, lastUsed, alias string
		if respJSON.Type == channelTypeEmail && respJSON.Verified == "nil" {
			handle = respJSON.Handle + " (unverified)"
		} else {
			handle = respJSON.Handle
		}
		if respJSON.UsedCount > 0 {
			lastUsed = respJSON.LastUsed
		}
		if respJSON.UsedCount == 0 {
			used = "never"
		} else {
			used = fmt.Sprintf("%d ×", respJSON.UsedCount)
		}
		if respJSON.Alias == "" {
			alias = "-"
		} else {
			alias = respJSON.Alias
		}

		// @todo show ID field in check and incident detail as well
		tableMainChannelCellContent := colorBold.Sprint(`ID: `) + respJSON.Ident + "\n" +
			colorBold.Sprint(`Type: `) + respJSON.Type + "\n" +
			colorBold.Sprint(`Alias: `) + alias + "\n" +
			colorBold.Sprint(`Handle: `) + handle + "\n" +
			colorBold.Sprint(`Used: `) + used + "\n" +
			colorBold.Sprint(`Last used: `) + lastUsed

		tableMain := tablewriter.NewWriter(os.Stdout)
		tableMain.SetHeader([]string{"CHANNEL"})
		tableMain.SetAutoWrapText(false)
		tableMain.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
		tableMain.SetHeaderColor(tablewriter.Colors{tablewriter.Bold})
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
	Aliases:           []string{"ls"},
	Args:              cobra.NoArgs,
	DisableAutoGenTag: true,
	Run: func(cmd *cobra.Command, args []string) {
		spin.Start()
		spin.Suffix = " loading channels..."

		urlValues := url.Values{}
		match, err := regexp.MatchString(validCheckIdentPattern, channelListFlagCheck)
		if err == nil && match {
			urlValues.Set("check", channelListFlagCheck)
		}
		channels, err := fetchChannels(urlValues)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		var tableData [][]string
		for _, v := range channels {
			var used, lastUsed, handle, alias, identSnippet string
			identSnippet = colorBold.Sprint(v.Ident)
			if v.UsedCount == 0 {
				used = colorFaint.Sprint("never")
				lastUsed = colorFaint.Sprint("n/a")
			} else {
				used = colorFaint.Sprintf("%d ×", v.UsedCount)
				lastUsed = colorFaint.Sprint(v.LastUsed)
			}
			if v.Type == channelTypeEmail && v.Verified == "nil" {
				handle = util.Ellipsis(v.Handle, 50) + " (unverified)"
			} else {
				handle = util.Ellipsis(v.Handle, 50)
			}
			if v.Alias == "" {
				alias = "-"
			} else {
				alias = v.Alias
			}
			tableRow := []string{
				identSnippet, v.Type, alias, handle, used, lastUsed,
			}
			tableData = append(tableData, tableRow)
		}
		table := tablewriter.NewWriter(os.Stdout)
		table.SetAutoWrapText(false)
		table.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
		table.SetHeader([]string{"ID", "TYPE", "ALIAS", "HANDLE", "USED", "LAST USED"})
		table.SetHeaderColor(tablewriter.Colors{tablewriter.Bold}, tablewriter.Colors{tablewriter.Bold}, tablewriter.Colors{tablewriter.Bold}, tablewriter.Colors{tablewriter.Bold}, tablewriter.Colors{tablewriter.Bold}, tablewriter.Colors{tablewriter.Bold})
		table.SetColumnAlignment([]int{tablewriter.ALIGN_DEFAULT, tablewriter.ALIGN_DEFAULT, tablewriter.ALIGN_DEFAULT, tablewriter.ALIGN_DEFAULT, tablewriter.ALIGN_RIGHT, tablewriter.ALIGN_RIGHT})
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
	Args:              cobra.ExactArgs(1),
	DisableAutoGenTag: true,
	Run: func(cmd *cobra.Command, args []string) {
		channelAddOrUpdate("update", args[0])
	},
}

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
		flagAttach []string
	)

	switch mode {
	case "add":
		flagAlias = channelAddFlagAlias
		flagHandle = channelAddFlagHandle
		flagType = channelAddFlagType
		flagAttach = channelAddFlagAttach
	case "update":
		flagAlias = channelUpdateFlagAlias
		flagAttach = channelUpdateFlagAttach
	}

	var currentChannel Channel
	if mode == "update" {
		spin.Start()
		spin.Suffix = " loading channel..."
		respData, err := util.BinocsAPI("/channels/"+channelIdent, http.MethodGet, []byte{})
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		err = json.Unmarshal(respData, &currentChannel)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		spin.Stop()
	}

	if mode == "update" {
		// pass; never update channel type
	} else {
		match, err = regexp.MatchString(validTypePattern, flagType)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		} else if !match {
			prompt := &survey.Select{
				Message: "Choose type:",
				Options: []string{channelTypeEmail, channelTypeSlack, channelTypeTelegram},
			}
			err = survey.AskOne(prompt, &flagType)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
		}
	}

	if mode == "update" {
		// pass; never update channel handle
	} else {
		if flagType == channelTypeEmail {
			match, err = regexp.MatchString(validHandlePattern[channelTypeEmail], flagHandle)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			} else if !match {
				validate := func(val interface{}) error {
					match, err = regexp.MatchString(validHandlePattern[channelTypeEmail], val.(string))
					if err != nil {
						return err
					} else if !match {
						return errors.New("invalid handle format")
					}
					return nil
				}
				prompt := &survey.Input{
					Message: "Enter a valid " + flagType + " handle:",
				}
				err = survey.AskOne(prompt, &flagHandle, survey.WithValidator(validate))
				if err != nil {
					fmt.Println(err)
					os.Exit(1)
				}
			}
		} else if flagType == channelTypeSlack {
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
		} else if flagType == channelTypeTelegram {
			telegramIntegrationToken, err := requestTelegramIntegrationToken()
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
			telegramInstallURL := "https://t.me/binocs_bot?start=" + telegramIntegrationToken.Token
			fmt.Println("Visit the following URL to add @binocs_bot to your Telegram:")
			fmt.Println(telegramInstallURL)
			spin.Start()
			spin.Suffix = " waiting for your action ..."
			for {
				pollResult, err := pollTelegramIntegrationStatus(telegramIntegrationToken.Token)
				if err != nil {
					fmt.Println(err)
					os.Exit(1)
				}
				if pollResult.Updated != "nil" {
					flagHandle = fmt.Sprintf("%d", pollResult.ChatID)
					break
				}
				time.Sleep(3 * time.Second)
			}
			fmt.Println("Successfully associated with Telegram.")
			spin.Stop()
		}
	}

	match, err = regexp.MatchString(validAliasPattern, flagAlias)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	} else if !match || flagAlias == "" {
		validate := func(val interface{}) error {
			match, err = regexp.MatchString(validAliasPattern, val.(string))
			if err != nil {
				return err
			} else if !match {
				return errors.New("invalid alias format")
			}
			return nil
		}
		prompt := &survey.Input{
			Message: "Channel alias (optional):",
		}
		if mode == "update" {
			prompt.Default = currentChannel.Alias
		}
		err = survey.AskOne(prompt, &flagAlias, survey.WithValidator(validate))
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	}

	spin.Start()
	spin.Suffix = " loading checks..."
	checks, err := fetchChecks(url.Values{})
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	spin.Stop()

	match, err = regexp.MatchString(validChecksIdentListPattern, strings.Join(flagAttach, ","))
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	} else if !match {
		var options = []string{}
		for _, c := range checks {
			options = append(options, c.Ident+" "+c.Identity())
		}
		var defaultOptions = []string{}
		if mode == "update" {
			for _, cc := range currentChannel.Checks {
				for _, c := range checks {
					if c.Ident == cc {
						defaultOption := c.Ident + " " + c.Identity()
						defaultOptions = append(defaultOptions, defaultOption)
					}
				}
			}
		}
		prompt := &survey.MultiSelect{
			Message:  "Checks to attach (optional):",
			Options:  options,
			Default:  defaultOptions,
			PageSize: 9,
		}
		err = survey.AskOne(prompt, &flagAttach)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	} else if strings.Join(flagAttach, ",") == "all" {
		flagAttach = []string{}
		for _, c := range checks {
			flagAttach = append(flagAttach, c.Ident)
		}
	}

	channel := Channel{
		Alias:  flagAlias,
		Handle: flagHandle,
		Type:   flagType,
	}
	postData, err := json.Marshal(channel)
	if err != nil {
		fmt.Println(err)
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
	spin.Start()
	spin.Suffix = " saving channel..."
	respData, err := util.BinocsAPI(reqURL, reqMethod, postData)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
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
			tpl = "[" + channel.Ident + "] " + channelDescription + ` added successfully`
		}
		if mode == "update" {
			tpl = "[" + channel.Ident + "] " + channelDescription + ` updated successfully`
		}
		spin.Suffix = " attaching channel to " + fmt.Sprintf("%d", len(flagAttach)) + " check(s)..."
		var detachCheckIdents = []string{}
		for _, c := range checks {
			for _, cc := range c.Channels {
				if cc == channel.Ident {
					detachCheckIdents = append(detachCheckIdents, c.Ident)
				}
			}
		}
		for _, c := range detachCheckIdents {
			deleteData, err := json.Marshal(ChannelAttachment{})
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
			_, err = util.BinocsAPI("/channels/"+channel.Ident+"/check/"+c, http.MethodDelete, deleteData)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
		}
		for _, fa := range flagAttach {
			attachIdent := strings.Split(fa, " ")[0]
			postData, err := json.Marshal(ChannelAttachment{})
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
			_, err = util.BinocsAPI("/channels/"+channel.Ident+"/check/"+attachIdent, http.MethodPost, postData)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
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
	spin.Stop()
	fmt.Println(tpl)
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

// TelegramIntegrationToken struct
type TelegramIntegrationToken struct {
	Token string `json:"token"`
}

func requestTelegramIntegrationToken() (TelegramIntegrationToken, error) {
	var token TelegramIntegrationToken
	respData, err := util.BinocsAPI("/integration/telegram/request-integration-token", http.MethodPost, []byte{})
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

// TelegramIntegrationStatus struct
type TelegramIntegrationStatus struct {
	Token   string `json:"token"`
	ChatID  int64  `json:"chat_id"`
	Updated string `json:"updated,omitempty"`
}

func pollTelegramIntegrationStatus(token string) (TelegramIntegrationStatus, error) {
	var status TelegramIntegrationStatus
	respData, err := util.BinocsAPI("/integration/telegram/status/"+token, http.MethodGet, []byte{})
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
