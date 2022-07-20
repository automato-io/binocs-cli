package cmd

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/automato-io/binocs-cli/util"
	"github.com/automato-io/s3update"
	"github.com/briandowns/spinner"
	"github.com/fatih/color"
	"github.com/muesli/reflow/ansi"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
	"golang.org/x/term"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/viper"
)

// BinocsVersion semver
const BinocsVersion = "v0.5.3"

const (
	statusUnknown  = 0
	statusStepUp   = 1
	statusUp       = 2
	statusStepDown = 3
	statusDown     = 4
)

const (
	statusNameUnknown  = "UNKNOWN"
	statusNameStepUp   = "UP (tentative)"
	statusNameUp       = "UP"
	statusNameStepDown = "DOWN (tentative)"
	statusNameDown     = "DOWN"
)

var statusName = map[int]string{
	statusUnknown:  statusNameUnknown,
	statusStepUp:   statusNameStepUp,
	statusUp:       statusNameUp,
	statusStepDown: statusNameStepDown,
	statusDown:     statusNameDown,
}

const (
	periodHour  = "hour"
	periodDay   = "day"
	periodWeek  = "week"
	periodMonth = "month"
)

var supportedPeriods = map[string]time.Duration{
	periodHour:  time.Duration(1 * time.Hour),
	periodDay:   time.Duration(24 * time.Hour),
	periodWeek:  time.Duration(7 * 24 * time.Hour),
	periodMonth: time.Duration(30 * 24 * time.Hour),
}

const (
	protocolHTTP  = "HTTP"
	protocolHTTPS = "HTTPS"
	protocolICMP  = "ICMP"
	protocolTCP   = "TCP"
)

const (
	incidentStateOpen     = "open"
	incidentStateResolved = "resolved"
)

var (
	colorBold      = color.New(color.Bold)
	colorFaint     = color.New(color.Faint)
	colorUnderline = color.New(color.Underline)
	colorFaintBold = color.New(color.Faint, color.Bold)
)

var (
	supportedRegions = []string{}
	defaultRegions   = []string{ // @todo fetch via API
		"us-east-1",
		"us-west-1",
		"ap-northeast-1",
		"ap-southeast-1",
		"eu-central-1",
		"eu-west-1",
	}
	regionAliases = map[string]string{
		"af-south-1":     "South Africa",
		"ap-east-1":      "Hong Kong",
		"ap-northeast-1": "Japan",
		"ap-south-1":     "India",
		"ap-southeast-1": "Singapore",
		"ap-southeast-2": "Australia",
		"eu-central-1":   "Germany",
		"eu-west-1":      "Ireland",
		"sa-east-1":      "Brazil",
		"us-east-1":      "US East",
		"us-west-1":      "US West",
	}
)

// Verbose flag
var Verbose bool

var AutoUpdateInterval = 3600 * 24 * 2

var cfgFile string

var spin = spinner.New(spinner.CharSets[53], 100*time.Millisecond, spinner.WithColor("faint"))

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "binocs",
	Short: "Monitoring tool for websites, applications and APIs",
	Long: `
Binocs is a CLI-first uptime and performance monitoring tool for websites, applications and APIs.

Binocs servers continuously measure uptime and performance of HTTP(S) or TCP endpoints. 

Get insight into current state of your endpoints and metrics history, and receive notifications about any incidents in real-time.

`,
	DisableAutoGenTag: true,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	//	Run: func(cmd *cobra.Command, args []string) { },
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		// fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)
	cobra.OnInitialize(initAutoUpdater)

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.binocs/config.json)")
	rootCmd.PersistentFlags().BoolVarP(&Verbose, "verbose", "v", false, "verbose output")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	var err error
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		home, err := homedir.Dir()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		if _, err = os.Stat(home + "/.binocs/config.json"); os.IsNotExist(err) {
			err = os.Mkdir(home+"/.binocs", 0755)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
			err = writeConfigTemplate(home + "/.binocs/config.json")
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
		}
		viper.AddConfigPath(home + "/.binocs/")
		viper.SetConfigName("config")
		viper.SetConfigType("json")
	}

	viper.AutomaticEnv()
	err = viper.ReadInConfig()
	if err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			fmt.Println("Config file not found")
		} else {
			fmt.Println("Cannot use config file:", viper.ConfigFileUsed())
		}
	} else if Verbose {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}

func writeConfigTemplate(path string) error {
	configContent := []byte("{\"access_key\": \"\", \"secret_key\": \"\"}")
	return ioutil.WriteFile(path, configContent, 0600)
}

func initAutoUpdater() {
	currentUnix := int(time.Now().UTC().Unix())
	lastUpdatedRaw := viper.GetString("update_last_checked")
	lastUpdated, err := strconv.Atoi(lastUpdatedRaw)
	if err != nil {
		if Verbose {
			fmt.Println(err)
		}
	}
	if lastUpdated+AutoUpdateInterval < currentUnix {
		err := s3update.AutoUpdate(s3update.Updater{
			CurrentVersion: BinocsVersion,
			S3VersionKey:   "VERSION",
			S3Bucket:       "binocs-download-website",
			S3ReleaseKey:   "binocs_{{VERSION}}_{{OS}}_{{ARCH}}.tgz",
			ChecksumKey:    "binocs_{{VERSION}}_{{OS}}_{{ARCH}}_checksum.txt",
			Verbose:        true,
		})
		if err != nil {
			fmt.Println("Error loading auto updater (2)")
			if Verbose {
				fmt.Println(err)
			}
		} else {
			viper.Set("update_last_checked", fmt.Sprintf("%v", currentUnix))
			err = viper.WriteConfigAs(viper.ConfigFileUsed())
			if err != nil {
				fmt.Println("Error loading auto updater (3)")
				if Verbose {
					fmt.Println(err)
				}
			}
		}
	}
}

func printZeroCreditsWarning() {
	creditsBalanceWarning := color.RedString("WARNING: ") + "Your credit balance reached zero and all your checks were paused.\nIf you wish to continue using Binocs, please visit the Settings page at " + colorUnderline.Sprint("https://binocs.sh/settings") + " to purchase additional credits.\nYour checks will resume once you top up credits."
	tableCreditsBalanceWarning := tablewriter.NewWriter(os.Stdout)
	tableCreditsBalanceWarning.SetBorders(tablewriter.Border{Left: true, Top: true, Right: true, Bottom: true})
	tableCreditsBalanceWarning.SetCenterSeparator(colorFaint.Sprint("┼"))
	tableCreditsBalanceWarning.SetColumnSeparator(colorFaint.Sprint("│"))
	tableCreditsBalanceWarning.SetRowSeparator(colorFaint.Sprint("─"))
	tableCreditsBalanceWarning.SetAutoWrapText(false)
	tableCreditsBalanceWarning.Append([]string{creditsBalanceWarning})
	tableCreditsBalanceWarning.Render()
}

type tableColumnDefinition struct {
	Header    string
	Priority  int8
	Alignment int
	hidden    bool
}

func composeTable(data [][]string, columnDefs []tableColumnDefinition) *tablewriter.Table {
	physicalWidth, _, _ := term.GetSize(int(os.Stdout.Fd()))
	tableCellWidths := make([]int, len(columnDefs))
	for _, v := range data {
		for i, w := range v {
			lines := regexp.MustCompile("\r?\n").Split(w, -1)
			lineMaxWidth := ansi.PrintableRuneWidth(columnDefs[i].Header)
			for _, line := range lines {
				l := ansi.PrintableRuneWidth(line)
				if l > lineMaxWidth {
					lineMaxWidth = l
				}
			}
			if tableCellWidths[i] < lineMaxWidth {
				tableCellWidths[i] = lineMaxWidth
			}
		}
	}
	for {
		currentColumnsCount := 0
		currentTableWidth := 0
		currentLowestPriority := 0
		currentShortestWidth := 0
		nextToHide := -1
		for i, c := range columnDefs {
			if !c.hidden {
				currentColumnsCount++
				currentTableWidth = currentTableWidth + tableCellWidths[i] + 2 // 2 for cell padding
				if currentLowestPriority < int(c.Priority) {
					currentLowestPriority = int(c.Priority)
				}
			}
		}
		currentTableWidth = currentTableWidth + currentColumnsCount + 1 // borders
		if currentTableWidth <= physicalWidth {
			break
		}
		for i, c := range columnDefs {
			if !c.hidden && c.Priority == int8(currentLowestPriority) {
				if currentShortestWidth == 0 || tableCellWidths[i] < currentShortestWidth {
					currentShortestWidth = tableCellWidths[i]
					nextToHide = i
				}
			}
		}
		if nextToHide < 0 {
			fmt.Println("Error drawing a table, cannot hide another column")
			os.Exit(1)
		}
		columnDefs[nextToHide].hidden = true
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetBorders(tablewriter.Border{Left: true, Top: true, Right: true, Bottom: true})
	table.SetCenterSeparator(colorFaint.Sprint("┼"))
	table.SetColumnSeparator(colorFaint.Sprint("│"))
	table.SetRowSeparator(colorFaint.Sprint("─"))
	table.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
	table.SetAutoWrapText(false)

	var tableHeaders []string
	var tableHeadersEnabled bool
	var tableHeadersColor []tablewriter.Colors
	var tableColumnAlignments []int
	for _, c := range columnDefs {
		if !c.hidden {
			tableHeaders = append(tableHeaders, c.Header)
			if len(c.Header) > 0 {
				tableHeadersEnabled = true
			}
			tableHeadersColor = append(tableHeadersColor, tablewriter.Colors{tablewriter.Bold})
			tableColumnAlignments = append(tableColumnAlignments, c.Alignment)
		}
	}

	if tableHeadersEnabled {
		table.SetHeader(tableHeaders)
		table.SetHeaderColor(tableHeadersColor...)
	}
	table.SetColumnAlignment(tableColumnAlignments)

	for _, v := range data {
		row := []string{}
		for i, c := range columnDefs {
			if !c.hidden {
				row = append(row, v[i])
			}
		}
		table.Append(row)
	}

	return table
}

func loadSupportedRegions() {
	respData, err := util.BinocsAPI("/regions", http.MethodGet, []byte{})
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	regionsResponse := RegionsResponse{}
	err = json.Unmarshal(respData, &regionsResponse)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	supportedRegions = regionsResponse.Regions
	sort.Strings(supportedRegions)
}

func getSupportedRegionAliases() []string {
	v := []string{}
	for k, a := range regionAliases {
		if util.StringInSlice(k, supportedRegions) {
			v = append(v, a)
		}
	}
	return v
}

func getDefaultRegionAliases() []string {
	v := []string{}
	for k, a := range regionAliases {
		if util.StringInSlice(k, supportedRegions) && util.StringInSlice(k, defaultRegions) {
			v = append(v, a)
		}
	}
	return v
}

func getRegionAliasesByIds(ids []string) []string {
	v := []string{}
	for k, a := range regionAliases {
		if util.StringInSlice(k, supportedRegions) && util.StringInSlice(k, ids) {
			v = append(v, a)
		}
	}
	return v
}

func getRegionIdByAlias(a string) string {
	for r, v := range regionAliases {
		if strings.EqualFold(v, a) {
			return r
		}
	}
	return ""
}

func getRegionIdsByAliases(as []string) []string {
	v := []string{}
	for k, a := range regionAliases {
		if util.StringInSlice(k, supportedRegions) && util.StringInSlice(a, as) {
			v = append(v, k)
		}
	}
	return v
}

func isValidRegionAlias(a string) bool {
	for k, v := range regionAliases {
		if strings.EqualFold(v, a) && util.StringInSlice(k, supportedRegions) {
			return true
		}
	}
	return false
}
