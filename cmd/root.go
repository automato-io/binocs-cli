package cmd

import (
	"fmt"
	"io/ioutil"
	"os"
	"time"

	"github.com/briandowns/spinner"
	"github.com/spf13/cobra"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/viper"
)

// BinocsVersion semver
const BinocsVersion = "0.3.0"

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

// Verbose flag
var Verbose bool

var cfgFile string

var spin = spinner.New(spinner.CharSets[7], 100*time.Millisecond)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "binocs-cli",
	Short: "Monitoring tool for websites, applications and APIs",
	Long: `
binocs is a devops-oriented monitoring tool for websites, applications and APIs. 

binocs continuously measures uptime and performance of http or tcp endpoints 
and provides insight into current state and metrics history. 
Get notified via Slack, Telegram and SMS.
`,
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
	// if the config's missing, assume it's the first run and create empty config, which will be used later

	cobra.OnInitialize(initConfig)

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.binocs-cli.json)")
	rootCmd.PersistentFlags().BoolVarP(&Verbose, "verbose", "v", false, "verbose output")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
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
	configContent := []byte("{\"access_key_id\": \"\", \"secret_access_key\": \"\"}")
	return ioutil.WriteFile(path, configContent, 0600)
}
