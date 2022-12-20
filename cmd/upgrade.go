package cmd

import (
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/automato-io/s3update"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func init() {
	rootCmd.AddCommand(upgradeCmd)
}

var upgradeCmd = &cobra.Command{
	Use:   "upgrade",
	Short: "Upgrade Binocs to the latest version",
	Long: `
Upgrade Binocs to the latest version
`,
	DisableAutoGenTag: true,
	// sister function of initAutoUpgrader()
	Run: func(cmd *cobra.Command, args []string) {
		currentTimestamp := int(time.Now().UTC().Unix())
		upgradeAvailable, versionAvailable, err := s3update.IsUpdateAvailable(autoUpdaterConfig)
		if err != nil {
			handleErr(err)
		}
		if upgradeAvailable {
			upgradeMessage := fmt.Sprintf("Binocs CLI %s is available. You are currently using version %s.\n", versionAvailable, BinocsVersion)
			upgradeMessage = upgradeMessage + "Attempting upgrade..."
			fmt.Println(upgradeMessage)
			filePath, err := os.Executable()
			if err != nil {
				handleErr(errors.New("Cannot determine current binary location."))
			}
			f, err := os.OpenFile(filePath, os.O_RDWR, 0755)
			if err != nil && os.IsPermission(err) {
				defer f.Close()
				handleErr(fmt.Errorf("Please execute this command as sudo for %s to be replaced.\n", filePath))
			}
			f.Close()
			err = s3update.AutoUpdate(autoUpdaterConfig)
			if err != nil {
				handleErr(err)
			}
			fmt.Println("Thank you for using Binocs 🙏")
		} else {
			viper.Set("upgrade_last_checked", fmt.Sprintf("%v", currentTimestamp))
			err = viper.WriteConfigAs(viper.ConfigFileUsed())
			if err != nil {
				handleErr(err)
			}
			fmt.Printf("You are currently using Binocs CLI %s.\n", versionAvailable)
			fmt.Println("Binocs CLI is up to date.")
		}
	},
}
