package cmd

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/AlecAivazis/survey/v2"
	util "github.com/automato-io/binocs-cli/util"
	"github.com/automato-io/tablewriter"
	"github.com/spf13/cobra"
)

// Channel comes from the API as a JSON
type Channel struct {
	ID        int      `json:"id,omitempty"`
	Ident     string   `json:"ident,omitempty"`
	Type      string   `json:"type,omitempty"`
	Alias     string   `json:"alias,omitempty"`
	Handle    string   `json:"handle,omitempty"`
	UsedCount int      `json:"used_count,omitempty"`
	LastUsed  string   `json:"last_used,omitempty"`
	Verified  string   `json:"verified,omitempty"`
	Checks    []string `json:"checks,omitempty"`
}

// Identity method returns "Type - Alias (handle)" or "handle"
func (ch Channel) Identity() string {
	// @todo remove in 0.8.x, alias always set
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
	validAliasPattern             = `^[\p{L}\p{N}_\s\/\-\.]{1,25}$`
	validTypePattern              = `^(email|slack|telegram|sms)$`
	validChannelsIdentListPattern = `^(all|([a-f0-9]{5})(,[a-f0-9]{5})*)$`
	validNotificationTypePattern  = `^(response-change|status)$`
	channelTypeEmail              = "email"
	channelTypeSms                = "sms"
	channelTypeSlack              = "slack"
	channelTypeTelegram           = "telegram"
)

var validHandlePattern = map[string]string{
	"email": `^(?:[a-z0-9!#$%&'*+/=?^_{|}~-]+(?:\.[a-z0-9!#$%&'*+/=?^_{|}~-]+)*|"(?:[\x01-\x08\x0b\x0c\x0e-\x1f\x21\x23-\x5b\x5d-\x7f]|\\[\x01-\x09\x0b\x0c\x0e-\x7f])*")@(?:(?:[a-z0-9](?:[a-z0-9-]*[a-z0-9])?\.)+[a-z0-9](?:[a-z0-9-]*[a-z0-9])?|\[(?:(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.){3}(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?|[a-z0-9-]*[a-z0-9]:(?:[\x01-\x08\x0b\x0c\x0e-\x1f\x21-\x5a\x53-\x7f]|\\[\x01-\x09\x0b\x0c\x0e-\x7f])+)\])$`,
	"sms":   `^\+?[1-9][0-9]{7,14}$`,
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

	channelAddCmd.Flags().StringVarP(&channelAddFlagType, "type", "t", "", "channel type (E-mail, Slack, Telegram, SMS)")
	channelAddCmd.Flags().StringVar(&channelAddFlagHandle, "handle", "", "channel handle - an address for \"E-mail\" channel type; a phone number for \"SMS\" channel type; handles for Slack and Telegram will be obtained programmatically")
	channelAddCmd.Flags().StringVar(&channelAddFlagAlias, "alias", "", "channel alias")
	channelAddCmd.Flags().StringSliceVar(&channelAddFlagAttach, "attach", []string{}, "checks to attach to this channel (optional); can be either \"all\", or one or more check identifiers")
	channelAddCmd.Flags().SortFlags = false

	channelsCmd.Flags().StringVarP(&channelListFlagCheck, "check", "c", "", "list only notification channels attached to a specific check")
	channelListCmd.Flags().StringVarP(&channelListFlagCheck, "check", "c", "", "list only notification channels attached to a specific check")

	channelUpdateCmd.Flags().StringVar(&channelUpdateFlagAlias, "alias", "", "channel alias")
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
Add a new notifications channel.

This command is interactive and asks user for parameters that were not provided as flags.
`,
	Aliases:           []string{"create"},
	Args:              cobra.NoArgs,
	DisableAutoGenTag: true,
	Run: func(cmd *cobra.Command, args []string) {
		util.VerifyAuthenticated()
		channelAddOrUpdate("add", "")
	},
}

var channelAttachCmd = &cobra.Command{
	Use:   "attach",
	Short: "Attach channel to check(s)",
	Long: `
Attach channel to check(s).

This command is interactive and asks user for confirmation.
`,
	Aliases:           []string{"att"},
	Args:              cobra.ExactArgs(1),
	DisableAutoGenTag: true,
	Run: func(cmd *cobra.Command, args []string) {
		util.VerifyAuthenticated()

		var err error

		if channelAttachFlagAll && len(channelAttachFlagCheck) > 0 {
			handleErr(fmt.Errorf("Cannot combine --all and --check flags"))
		}

		spin.Start()
		defer spin.Stop()
		spin.Suffix = colorFaint.Sprintf(" loading channel %s", args[0])
		channelRespData, err := util.BinocsAPI("/channels/"+args[0], http.MethodGet, []byte{})
		if err != nil {
			handleErr(err)
		}
		var currentRespJSON Channel
		err = json.Unmarshal(channelRespData, &currentRespJSON)
		if err != nil {
			handleErr(err)
		}
		checkIdents := []string{}
		if channelAttachFlagAll {
			checks, err := fetchChecks(url.Values{})
			if err != nil {
				handleErr(err)
			}
			for _, c := range checks {
				checkIdents = append(checkIdents, c.Ident)
			}
		} else {
			// validate checks against pattern, single or slice, required
			if len(channelAttachFlagCheck) == 0 {
				handleErr(fmt.Errorf("Set at least one check to attach to the channel"))
			}
			checkIdents = channelAttachFlagCheck
			for _, c := range checkIdents {
				match, err := regexp.MatchString(validCheckIdentPattern, c)
				if err != nil {
					handleErr(err)
				} else if !match {
					handleErr(fmt.Errorf("Provided check identifier is invalid"))
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
			handleErr(err)
		}
		if yes {
			spin.Start()
			defer spin.Stop()
			spin.Suffix = colorFaint.Sprintf(" attaching channel %s to %d checks", args[0], len(checkIdents))
			for _, c := range checkIdents {
				postData, err := json.Marshal(ChannelAttachment{})
				if err != nil {
					handleErr(err)
				}
				_, err = util.BinocsAPI("/channels/"+args[0]+"/check/"+c, http.MethodPost, postData)
				if err != nil {
					handleErr(err)
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
Detach channel from check(s).

This command is interactive and asks user for confirmation.
`,
	Args:              cobra.ExactArgs(1),
	DisableAutoGenTag: true,
	Run: func(cmd *cobra.Command, args []string) {
		util.VerifyAuthenticated()

		var err error

		if channelDetachFlagAll && len(channelDetachFlagCheck) > 0 {
			handleErr(fmt.Errorf("Cannot combine --all and --check flags"))
		}

		spin.Start()
		defer spin.Stop()
		spin.Suffix = colorFaint.Sprintf(" loading channel %s", args[0])
		channelRespData, err := util.BinocsAPI("/channels/"+args[0], http.MethodGet, []byte{})
		if err != nil {
			handleErr(err)
		}
		var currentRespJSON Channel
		err = json.Unmarshal(channelRespData, &currentRespJSON)
		if err != nil {
			handleErr(err)
		}
		checkIdents := []string{}
		if channelDetachFlagAll {
			checks, err := fetchChecks(url.Values{})
			if err != nil {
				handleErr(err)
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
				handleErr(fmt.Errorf("Set at least one check to detach from the channel"))
			}
			checkIdents = channelDetachFlagCheck
			for _, c := range checkIdents {
				match, err := regexp.MatchString(validCheckIdentPattern, c)
				if err != nil {
					handleErr(err)
				} else if !match {
					handleErr(fmt.Errorf("Provided check identifier is invalid"))
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
			handleErr(err)
		}
		if yes {
			spin.Start()
			defer spin.Stop()
			spin.Suffix = colorFaint.Sprintf(" detaching channel %s from %d checks", args[0], len(checkIdents))
			for _, c := range checkIdents {
				deleteData, err := json.Marshal(ChannelAttachment{})
				if err != nil {
					handleErr(err)
				}
				_, err = util.BinocsAPI("/channels/"+args[0]+"/check/"+c, http.MethodDelete, deleteData)
				if err != nil {
					handleErr(err)
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

This command is interactive and asks for confirmation.
`,
	Aliases:           []string{"del", "rm"},
	Args:              cobra.MatchAll(),
	DisableAutoGenTag: true,
	Run: func(cmd *cobra.Command, args []string) {
		util.VerifyAuthenticated()

		for _, arg := range args {
			respData, err := util.BinocsAPI("/channels/"+arg, http.MethodGet, []byte{})
			if err != nil {
				handleWarn("Error loading channel " + arg)
				continue
			}
			var respJSON Channel
			err = json.Unmarshal(respData, &respJSON)
			if err != nil {
				handleWarn("Invalid response from binocs.sh")
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
					handleWarn("Error deleting channel " + arg)
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
		util.VerifyAuthenticated()

		spin.Start()
		defer spin.Stop()
		spin.Suffix = colorFaint.Sprint(" loading channel...")
		respData, err := util.BinocsAPI("/channels/"+args[0], http.MethodGet, []byte{})
		if err != nil {
			handleErr(err)
		}
		var respJSON Channel
		err = json.Unmarshal(respData, &respJSON)
		if err != nil {
			handleErr(err)
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
		// @todo remove in 0.8.x, alias always set
		if respJSON.Alias == "" {
			alias = "-"
		} else {
			alias = respJSON.Alias
		}

		tableMainChannelCellContent := colorBold.Sprint(`ID: `) + respJSON.Ident + "\n" +
			colorBold.Sprint(`Type: `) + respJSON.Type + "\n" +
			colorBold.Sprint(`Alias: `) + alias + "\n" +
			colorBold.Sprint(`Handle: `) + handle + "\n" +
			colorBold.Sprint(`Last used: `) + lastUsed + "\n" +
			colorBold.Sprint(`Used: `) + used

		var tableMainChecksCellContent []string
		if len(respJSON.Checks) > 0 {
			checks, err := fetchChecks(url.Values{})
			if err != nil {
				handleErr(err)
			}
			for _, c := range checks {
				for _, cc := range respJSON.Checks {
					if c.Ident == cc {
						tableMainChecksCellContent = append(tableMainChecksCellContent, c.Ident+" - "+c.Identity())
					}
				}
			}
		} else {
			tableMainChecksCellContent = []string{"-"}
		}

		columnDefinitions := []tableColumnDefinition{
			{
				Header:    "CHANNEL",
				Priority:  1,
				Alignment: tablewriter.ALIGN_LEFT,
			},
			{
				Header:    "ATTACHED CHECKS",
				Priority:  2,
				Alignment: tablewriter.ALIGN_LEFT,
			},
		}

		var tableData [][]string
		tableData = append(tableData, []string{tableMainChannelCellContent, strings.Join(tableMainChecksCellContent, "\n")})
		tableMain := composeTable(tableData, columnDefinitions)

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
		util.VerifyAuthenticated()

		spin.Start()
		defer spin.Stop()
		spin.Suffix = colorFaint.Sprint(" loading channels...")

		urlValues := url.Values{}
		match, err := regexp.MatchString(validCheckIdentPattern, channelListFlagCheck)
		if err == nil && match {
			urlValues.Set("check", channelListFlagCheck)
		}
		channels, err := fetchChannels(urlValues)
		if err != nil {
			handleErr(err)
		}

		var tableData [][]string
		for _, v := range channels {
			var attached, used, lastUsed, handle, alias, identSnippet string
			identSnippet = colorBold.Sprint(v.Ident)
			if v.UsedCount == 0 {
				lastUsed = "n/a"
			} else {
				lastUsed = v.LastUsed
			}
			used = fmt.Sprintf("%d ×", v.UsedCount)
			attached = fmt.Sprintf("%d check(s)", len(v.Checks))
			if v.Type == channelTypeEmail && v.Verified == "nil" {
				handle = util.Ellipsis(v.Handle, 50) + " (unverified)"
			} else {
				handle = util.Ellipsis(v.Handle, 50)
			}
			// @todo remove in 0.8.x, alias always set
			if v.Alias == "" {
				alias = "-"
			} else {
				alias = v.Alias
			}
			tableRow := []string{
				identSnippet, v.Type, alias, handle, attached, used, lastUsed,
			}
			tableData = append(tableData, tableRow)
		}

		columnDefinitions := []tableColumnDefinition{
			{
				Header:    "ID",
				Priority:  1,
				Alignment: tablewriter.ALIGN_LEFT,
			},
			{
				Header:    "TYPE",
				Priority:  2,
				Alignment: tablewriter.ALIGN_LEFT,
			},
			{
				Header:    "ALIAS",
				Priority:  3,
				Alignment: tablewriter.ALIGN_LEFT,
			},
			{
				Header:    "HANDLE",
				Priority:  2,
				Alignment: tablewriter.ALIGN_LEFT,
			},
			{
				Header:    "ATTACHED",
				Priority:  3,
				Alignment: tablewriter.ALIGN_RIGHT,
			},
			{
				Header:    "USED",
				Priority:  3,
				Alignment: tablewriter.ALIGN_RIGHT,
			},
			{
				Header:    "LAST USED",
				Priority:  3,
				Alignment: tablewriter.ALIGN_RIGHT,
			},
		}

		table := composeTable(tableData, columnDefinitions)
		spin.Stop()
		table.Render()
	},
}

var channelUpdateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update existing notification channel",
	Long: `
Update existing notification channel.

This command is interactive and asks user for parameters that were not provided as flags.
`,
	Args:              cobra.ExactArgs(1),
	DisableAutoGenTag: true,
	Run: func(cmd *cobra.Command, args []string) {
		util.VerifyAuthenticated()
		channelAddOrUpdate("update", args[0])
	},
}

func channelAddOrUpdate(mode string, channelIdent string) {
	if mode != "add" && mode != "update" {
		handleErr(fmt.Errorf("Unknown mode: " + mode))
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
		defer spin.Stop()
		spin.Suffix = colorFaint.Sprint(" loading channel...")
		respData, err := util.BinocsAPI("/channels/"+channelIdent, http.MethodGet, []byte{})
		if err != nil {
			handleErr(err)
		}
		err = json.Unmarshal(respData, &currentChannel)
		if err != nil {
			handleErr(err)
		}
		spin.Stop()
	}

	if mode == "update" {
		// pass; never update channel type
	} else {
		match, err = regexp.MatchString(validTypePattern, flagType)
		if err != nil {
			handleErr(err)
		} else if !match {
			prompt := &survey.Select{
				Message: "Choose type:",
				Options: []string{channelTypeEmail, channelTypeSlack, channelTypeTelegram, channelTypeSms},
			}
			err = survey.AskOne(prompt, &flagType)
			if err != nil {
				handleErr(err)
			}
		}
	}

	if mode == "update" {
		// pass; never update channel handle
	} else {
		if flagType == channelTypeEmail {
			match, err = regexp.MatchString(validHandlePattern[channelTypeEmail], flagHandle)
			if err != nil {
				handleErr(err)
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
					Message: "Enter a valid e-mail address:",
				}
				err = survey.AskOne(prompt, &flagHandle, survey.WithValidator(validate))
				if err != nil {
					handleErr(err)
				}
			}
		} else if flagType == channelTypeSms {
			match, err = regexp.MatchString(validHandlePattern[channelTypeSms], flagHandle)
			if err != nil {
				handleErr(err)
			} else if !match {
				validate := func(val interface{}) error {
					match, err = regexp.MatchString(validHandlePattern[channelTypeSms], val.(string))
					if err != nil {
						return err
					} else if !match {
						return errors.New("invalid handle format")
					}
					return nil
				}
				prompt := &survey.Input{
					Message: "Enter a valid phone number:",
				}
				err = survey.AskOne(prompt, &flagHandle, survey.WithValidator(validate))
				if err != nil {
					handleErr(err)
				}
			}
			var smsVerificationInput string
			smsVerificationCode, err := sendVerificationSms(flagHandle)
			if err != nil {
				handleErr(err)
			}
			validate := func(val interface{}) error {
				if val.(string) == smsVerificationCode.Token {
					return nil
				}
				return errors.New("invalid code")
			}
			fmt.Println("We sent you a verification SMS with a 4 digit code.")
			smsVerifyPrompt := &survey.Input{
				Message: "SMS verification code:",
			}
			err = survey.AskOne(smsVerifyPrompt, &smsVerificationInput, survey.WithValidator(validate))
			if err != nil {
				handleErr(err)
			}
		} else if flagType == channelTypeSlack {
			slackIntegrationToken, err := requestSlackIntegrationToken()
			if err != nil {
				handleErr(err)
			}
			slackScope := "incoming-webhook"
			slackRedirectURI := "https://binocs.sh/integration/slack/callback"
			slackClientID := "1145502794960.1106931893399"
			slackAuthorizeURL := "https://slack.com/oauth/v2/authorize?scope=" + url.QueryEscape(slackScope) + "&client_id=" + url.QueryEscape(slackClientID) + "&redirect_uri=" + url.QueryEscape(slackRedirectURI) + "&state=" + slackIntegrationToken.Token
			fmt.Println("Visit the following URL to choose where we should send your notifications:")
			fmt.Println(slackAuthorizeURL)
			spin.Start()
			defer spin.Stop()
			spin.Suffix = colorFaint.Sprint(" waiting for your action ...")
			for {
				pollResult, err := pollSlackIntegrationStatus(slackIntegrationToken.Token)
				if err != nil {
					handleErr(err)
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
				handleErr(err)
			}
			telegramInstallURL := "https://t.me/binocs_bot?start=" + telegramIntegrationToken.Token
			fmt.Println("Visit the following URL to add @binocs_bot to your Telegram:")
			fmt.Println(telegramInstallURL)
			spin.Start()
			defer spin.Stop()
			spin.Suffix = colorFaint.Sprint(" waiting for your action ...")
			for {
				pollResult, err := pollTelegramIntegrationStatus(telegramIntegrationToken.Token)
				if err != nil {
					handleErr(err)
				}
				if pollResult.Updated != "nil" {
					flagHandle = fmt.Sprintf("%d", pollResult.ChatID)
					break
				}
				time.Sleep(3 * time.Second)
			}
			spin.Stop()
			fmt.Println("Successfully associated with Telegram.")
		}
	}

	match, err = regexp.MatchString(validAliasPattern, flagAlias)
	if err != nil {
		handleErr(err)
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
			Message: "Channel alias:",
		}
		if mode == "update" {
			prompt.Default = currentChannel.Alias
		}
		err = survey.AskOne(prompt, &flagAlias, survey.WithValidator(validate))
		if err != nil {
			handleErr(err)
		}
	}

	spin.Start()
	defer spin.Stop()
	spin.Suffix = colorFaint.Sprint(" loading checks...")
	checks, err := fetchChecks(url.Values{})
	if err != nil {
		handleErr(err)
	}
	spin.Stop()

	if len(checks) > 0 {
		match, err = regexp.MatchString(validChecksIdentListPattern, strings.Join(flagAttach, ","))
		if err != nil {
			handleErr(err)
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
				handleErr(err)
			}
		} else if strings.Join(flagAttach, ",") == "all" {
			flagAttach = []string{}
			for _, c := range checks {
				flagAttach = append(flagAttach, c.Ident)
			}
		}
	}

	channel := Channel{
		Alias:  flagAlias,
		Handle: flagHandle,
		Type:   flagType,
	}
	postData, err := json.Marshal(channel)
	if err != nil {
		handleErr(err)
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
	defer spin.Stop()
	spin.Suffix = colorFaint.Sprint(" saving channel...")
	respData, err := util.BinocsAPI(reqURL, reqMethod, postData)
	if err != nil {
		spin.Stop()
		handleErr(err)
	}
	err = json.Unmarshal(respData, &channel)
	if err != nil {
		spin.Stop()
		handleErr(err)
	}
	if channel.ID > 0 {
		var channelDescription string
		// @todo remove in 0.8.x, alias always set
		if len(channel.Alias) > 0 {
			channelDescription = `"` + channel.Alias + `"` + " (" + channel.Handle + ")"
		} else {
			channelDescription = `"` + channel.Handle + `"`
		}
		if mode == "add" {
			tpl = channel.Type + " notifications channel " + channelDescription + " [" + channel.Ident + `] added successfully`
		}
		if mode == "update" {
			tpl = channel.Type + " notifications channel " + channelDescription + " [" + channel.Ident + `] updated successfully`
		}
		if len(flagAttach) > 0 {
			spin.Suffix = colorFaint.Sprintf(" attaching channel to %d check(s)...", len(flagAttach))
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
					spin.Stop()
					handleErr(err)
				}
				_, err = util.BinocsAPI("/channels/"+channel.Ident+"/check/"+c, http.MethodDelete, deleteData)
				if err != nil {
					spin.Stop()
					handleErr(err)
				}
			}
			for _, fa := range flagAttach {
				attachIdent := strings.Split(fa, " ")[0]
				postData, err := json.Marshal(ChannelAttachment{})
				if err != nil {
					spin.Stop()
					handleErr(err)
				}
				_, err = util.BinocsAPI("/channels/"+channel.Ident+"/check/"+attachIdent, http.MethodPost, postData)
				if err != nil {
					spin.Stop()
					handleErr(err)
				}
			}
		}
	} else {
		if mode == "add" {
			handleErr(fmt.Errorf("Error adding channel"))
		}
		if mode == "update" {
			handleErr(fmt.Errorf("Error updating channel"))
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
		handleErr(err)
	}
	err = json.Unmarshal(respData, &token)
	if err != nil {
		handleErr(err)
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
		handleErr(err)
	}
	err = json.Unmarshal(respData, &status)
	if err != nil {
		handleErr(err)
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
		handleErr(err)
	}
	err = json.Unmarshal(respData, &token)
	if err != nil {
		handleErr(err)
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
		handleErr(err)
	}
	err = json.Unmarshal(respData, &status)
	if err != nil {
		handleErr(err)
	}
	return status, nil
}

// SmsVerificationRequest as expected input
type SmsVerificationRequest struct {
	Number string `json:"number"`
}

// SmsVerificationToken as expected input
type SmsVerificationToken struct {
	Token string `json:"token"`
}

func sendVerificationSms(num string) (SmsVerificationToken, error) {
	var token SmsVerificationToken
	postData, err := json.Marshal(&SmsVerificationRequest{Number: num})
	if err != nil {
		return token, err
	}
	respData, err := util.BinocsAPI("/integration/sms/request-verification-token", http.MethodPost, postData)
	if err != nil {
		return token, err
	}
	err = json.Unmarshal(respData, &token)
	if err != nil {
		return token, err
	}
	return token, nil
}
