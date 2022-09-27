package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"regexp"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/automato-io/binocs-cli/util"
	"github.com/automato-io/s3update"
	"github.com/automato-io/tablewriter"
	"github.com/briandowns/spinner"
	"github.com/fatih/color"
	"github.com/gdamore/tcell/v2"
	"github.com/muesli/reflow/ansi"
	"github.com/rivo/tview"
	"github.com/spf13/cobra"
	"golang.org/x/term"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/viper"

	_ "time/tzdata"
)

// BinocsVersion semver
const BinocsVersion = "v0.7.1"

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

const (
	watchInterval = time.Duration(5 * time.Second)
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

var (
	Verbose bool
	Quiet   bool
)

var (
	autoUpgradeInterval = 3600 * 24 * 2
	autoUpdaterConfig   = s3update.Updater{
		CurrentVersion: BinocsVersion,
		S3VersionKey:   "VERSION",
		S3Bucket:       "binocs-download-website",
		S3ReleaseKey:   "binocs_{{VERSION}}_{{OS}}_{{ARCH}}",
		ChecksumKey:    "binocs_{{VERSION}}_{{OS}}_{{ARCH}}_checksum.txt",
		Verbose:        false,
	}
)

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
	err := rootCmd.Execute()
	if err != nil {
		// fmt.Println(err)
		os.Exit(1)
	}
}

func runAsWatch() {
	// windows are unsupported because github.com/iamacarpet/go-winpty crashes during build in CI
	if runtime.GOOS == "windows" {
		fmt.Println("The --watch flag is not currently supported on Windows.")
		os.Exit(1)
	}
	app, screen, viewer, err := initWatchEnv()
	if err != nil {
		screen.Suspend()
		fmt.Println(err)
		os.Exit(1)
	}
	args := os.Args
	args = append(args, "--quiet")
	for i, v := range args {
		if v == "--watch" {
			args = append(args[:i], args[i+1:]...)
		}
	}
	go func() {
		for {
			var buf bytes.Buffer
			var cmdStart, cmdEnd time.Time
			var watchIntervalRemainder time.Duration
			cmdStart = time.Now()
			cmd := exec.Command(args[0], args[1:]...)
			err := util.CmdOutput(cmd, &buf)
			if err != nil {
				screen.Suspend()
				fmt.Println(err)
				os.Exit(1)
			}
			app.QueueUpdateDraw(func() {
				screen.Clear()
				viewer.SetText(tview.TranslateANSI(buf.String()))
			})
			cmdEnd = time.Now()
			watchIntervalRemainder = watchInterval - cmdEnd.Sub(cmdStart)
			if watchIntervalRemainder > 0 {
				time.Sleep(watchIntervalRemainder)
			}
		}
	}()
	err = app.Run()
	if err != nil {
		screen.Suspend()
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initRuntimeSystemSpecifics)
	cobra.OnInitialize(initConfig)
	cobra.OnInitialize(initAutoUpgrader)
	cobra.OnInitialize(initGlobalFlags)
	cobra.OnInitialize(initClientKey)
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.binocs/config.json)")
	rootCmd.PersistentFlags().BoolVarP(&Quiet, "quiet", "q", false, "enable quiet mode (hide spinners and progress bars)")
	rootCmd.PersistentFlags().BoolVarP(&Verbose, "verbose", "v", false, "verbose output")
}

func initRuntimeSystemSpecifics() {
	if runtime.GOOS == "windows" {
		color.NoColor = true
	}
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
			err = os.MkdirAll(home+"/.binocs", 0755)
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
	configContent := []byte("{\"client_key\": \"\"}")
	return ioutil.WriteFile(path, configContent, 0600)
}

// sister function of upgradeCmd.Run()
func initAutoUpgrader() {
	currentTimestamp := int(time.Now().UTC().Unix())
	// ignore error, proceed, and try to write the correct value after a successful check
	lastUpgraded, _ := strconv.Atoi(viper.GetString("upgrade_last_checked"))
	if lastUpgraded+autoUpgradeInterval < currentTimestamp {
		upgradeAvailable, versionAvailable, err := s3update.IsUpdateAvailable(autoUpdaterConfig)
		if err != nil {
			fmt.Println(err)
			return
		}
		if upgradeAvailable {
			if !Quiet {
				upgradeMessage := fmt.Sprintf("Binocs CLI %s is available. You are currently using version %s.\n", versionAvailable, BinocsVersion)
				upgradeMessage = upgradeMessage + "Run " + colorBold.Sprint("binocs upgrade") + " to get the latest version.\n"
				fmt.Print(upgradeMessage)
			}
		} else {
			viper.Set("upgrade_last_checked", fmt.Sprintf("%v", currentTimestamp))
			err = viper.WriteConfigAs(viper.ConfigFileUsed())
			if err != nil {
				fmt.Println(err)
			}
		}
	}
}

func initClientKey() {
	clientKey := viper.Get("client_key")
	if clientKey == nil || len(clientKey.(string)) != 40 { // sha1 hash length
		viper.Set("client_key", generateClientKey())
		err := viper.WriteConfigAs(viper.ConfigFileUsed())
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		err = util.ResetAccessToken()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	}
}

func initGlobalFlags() {
	if Quiet {
		spin.Disable()
	}
}

func initWatchEnv() (*tview.Application, tcell.Screen, *tview.TextView, error) {
	screen, err := tcell.NewScreen()
	if err != nil {
		return nil, screen, nil, err
	}
	err = screen.Init()
	if err != nil {
		return nil, screen, nil, err
	}
	app := tview.NewApplication()
	app.SetScreen(screen)
	viewer := tview.NewTextView().SetDynamicColors(true).SetScrollable(true).SetTextColor(tcell.ColorDefault)
	viewer.SetBackgroundColor(tcell.ColorDefault)
	flex := tview.NewFlex().SetDirection(tview.FlexRow)
	flex.AddItem(viewer, 0, 1, true)
	app.SetRoot(flex, true)
	return app, screen, viewer, err
}

//

func printZeroCreditsWarning() {
	creditsBalanceWarning := color.RedString("WARNING: ") + "Your credit balance reached zero and all your checks were paused.\nIf you wish to continue using Binocs, please visit the Settings page at " + colorUnderline.Sprint("https://binocs.sh/settings") + " to purchase additional credits.\nYour checks will resume once you top up credits."
	tableCreditsBalanceWarning := tablewriter.NewWriter(os.Stdout)
	tableCreditsBalanceWarning.SetBorders(tablewriter.Border{Left: true, Top: true, Right: true, Bottom: true})
	tableCreditsBalanceWarning.SetBorderSymbols(tablewriter.BorderSymbols{
		Horizontal:  colorFaint.Sprint("─"),
		Vertical:    colorFaint.Sprint("│"),
		Center:      colorFaint.Sprint("┼"),
		Top:         colorFaint.Sprint("┬"),
		TopRight:    colorFaint.Sprint("┐"),
		Right:       colorFaint.Sprint("┤"),
		BottomRight: colorFaint.Sprint("┘"),
		Bottom:      colorFaint.Sprint("┴"),
		BottomLeft:  colorFaint.Sprint("└"),
		Left:        colorFaint.Sprint("├"),
		TopLeft:     colorFaint.Sprint("┌"),
	})
	tableCreditsBalanceWarning.SetAutoWrapText(false)
	tableCreditsBalanceWarning.Append([]string{creditsBalanceWarning})
	tableCreditsBalanceWarning.Render()
}

//

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
	table.SetBorderSymbols(tablewriter.BorderSymbols{
		Horizontal:  colorFaint.Sprint("─"),
		Vertical:    colorFaint.Sprint("│"),
		Center:      colorFaint.Sprint("┼"),
		Top:         colorFaint.Sprint("┬"),
		TopRight:    colorFaint.Sprint("┐"),
		Right:       colorFaint.Sprint("┤"),
		BottomRight: colorFaint.Sprint("┘"),
		Bottom:      colorFaint.Sprint("┴"),
		BottomLeft:  colorFaint.Sprint("└"),
		Left:        colorFaint.Sprint("├"),
		TopLeft:     colorFaint.Sprint("┌"),
	})
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
		if !color.NoColor {
			table.SetHeaderColor(tableHeadersColor...)
		}
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

//

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
	sort.Strings(v)
	return v
}

func getDefaultRegionAliases() []string {
	v := []string{}
	for k, a := range regionAliases {
		if util.StringInSlice(k, supportedRegions) && util.StringInSlice(k, defaultRegions) {
			v = append(v, a)
		}
	}
	sort.Strings(v)
	return v
}

func getRegionAliasesByIds(ids []string) []string {
	v := []string{}
	for k, a := range regionAliases {
		if util.StringInSlice(k, supportedRegions) && util.StringInSlice(k, ids) {
			v = append(v, a)
		}
	}
	sort.Strings(v)
	return v
}

func getRegionIdByAlias(a string) string {
	for k, v := range regionAliases {
		if util.StringInSlice(k, supportedRegions) && strings.EqualFold(v, a) {
			return k
		}
	}
	return ""
}

func getRegionIdsByAliases(as []string) []string {
	res := []string{}
	for _, a := range as {
		r := getRegionIdByAlias(a)
		if len(r) > 0 {
			res = append(res, r)
		}
	}
	sort.Strings(res)
	return res
}

func isValidRegionAlias(a string) bool {
	for k, v := range regionAliases {
		if util.StringInSlice(k, supportedRegions) && strings.EqualFold(v, a) {
			return true
		}
	}
	return false
}
