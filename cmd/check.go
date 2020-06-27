package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"

	util "github.com/automato-io/binocs-cli/util"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
)

// CheckResponse comes from the API as a JSON
type CheckResponse struct {
	ID                         int      `json:"id"`
	Name                       string   `json:"name"`
	URL                        string   `json:"url"`
	Method                     string   `json:"method"`
	Interval                   int      `json:"interval"`
	Target                     float64  `json:"target"`
	Regions                    []string `json:"regions"`
	UpCodes                    string   `json:"up_codes"`
	UpConfirmationsThreshold   int      `json:"up_confirmations_threshold"`
	UpConfirmations            int      `json:"up_confirmations"`
	DownConfirmationsThreshold int      `json:"down_confirmations_threshold"`
	DownConfirmations          int      `json:"down_confirmations"`
	LastStatus                 int      `json:"last_status"`
	LastStatusCode             string   `json:"last_status_code"`
	Created                    string   `json:"created"`
	Updated                    string   `json:"updated"`
}

// MetricsResponse comes from the API as a JSON
type MetricsResponse struct {
	Apdex  string `json:"apdex"`
	MRT    string `json:"mrt"`
	Uptime string `json:"uptime"`
}

// ApdexResponse comes from the API as a JSON
type ApdexResponse struct {
	Apdex string `json:"apdex"`
	From  string `json:"from"`
	To    string `json:"to"`
}

func init() {
	rootCmd.AddCommand(checkCmd)
	checkCmd.AddCommand(checkAddCmd)
	checkCmd.AddCommand(checkInspectCmd)
	checkCmd.AddCommand(checkUpdateCmd)
	checkCmd.AddCommand(checkDeleteCmd)
}

var checkCmd = &cobra.Command{
	Use:     "check",
	Short:   "Manage your checks - checks are the endpoints monitored by binocs",
	Long:    `...`,
	Aliases: []string{"checks"},
	Example: "",
	Run: func(cmd *cobra.Command, args []string) {
		respData, err := util.BinocsAPI("/checks", http.MethodGet, []byte{})
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		respJSON := make([]CheckResponse, 0)
		decoder := json.NewDecoder(bytes.NewBuffer(respData))
		err = decoder.Decode(&respJSON)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		var tableData [][]string
		for i, v := range respJSON {
			// @todo accept period, region params
			metricsData, err := util.BinocsAPI("/checks/"+strconv.Itoa(v.ID)+"/metrics?period=day", http.MethodGet, []byte{})
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
			var metrics MetricsResponse
			err = json.Unmarshal(metricsData, &metrics)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
			if metrics.Uptime == "100.00" {
				metrics.Uptime = "100"
			}

			apdexData, err := util.BinocsAPI("/checks/"+strconv.Itoa(v.ID)+"/apdex?period=day", http.MethodGet, []byte{})
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
			apdex := make([]ApdexResponse, 0)
			decoder := json.NewDecoder(bytes.NewBuffer(apdexData))
			err = decoder.Decode(&apdex)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}

			apdexChart := drawCompactApdexChart(apdex, 3)

			tableRow := []string{
				strconv.Itoa(i + 1), v.Name, v.URL, statusName[v.LastStatus], v.LastStatusCode, strconv.Itoa(v.Interval) + " s", fmt.Sprintf("%.3f s", v.Target), metrics.MRT + " s", fmt.Sprintf("%v %%", metrics.Uptime), metrics.Apdex, apdexChart,
			}
			tableData = append(tableData, tableRow)
		}
		table := tablewriter.NewWriter(os.Stdout)
		table.SetHeader([]string{"#", "NAME", "URL", "STATUS", "HTTP CODE", "INTERVAL", "TARGET", "MRT", "UPTIME", "APDEX", "APDEX 24 h"})

		for _, v := range tableData {
			table.Append(v)
		}
		table.Render()
	},
}

var checkAddCmd = &cobra.Command{
	Use: "add",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("check add")
	},
}

var checkInspectCmd = &cobra.Command{
	Use:     "inspect",
	Aliases: []string{"view", "show"},
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			fmt.Println("RTFM")
			os.Exit(1)
		} else if _, err := strconv.Atoi(args[0]); err != nil {
			fmt.Println("only reference by id allowed atm")
			os.Exit(1)
		}
		respData, err := util.BinocsAPI("/checks/"+args[0], http.MethodGet, []byte{})
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		var respJSON CheckResponse
		err = json.Unmarshal(respData, &respJSON)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		tpl := `Name: ` + respJSON.Name + `
URL: ` + respJSON.URL + `
Method: ` + respJSON.Method + `
Status: ` + respJSON.LastStatusCode + `
Interval: ` + strconv.Itoa(respJSON.Interval) + ` s
Target response time: ` + fmt.Sprintf("%.3f", respJSON.Target) + ` s
Check from: ` + strings.Join(respJSON.Regions, ", ") + `
`
		fmt.Print(tpl)
	},
}

var checkUpdateCmd = &cobra.Command{
	Use: "update",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("check update")
	},
}

var checkDeleteCmd = &cobra.Command{
	Use:     "delete",
	Aliases: []string{"del"},
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("check delete")
	},
}

func drawCompactApdexChart(apdex []ApdexResponse, compress int) string {
	var compressed []float64
	var compressStack []float64
	for i, v := range apdex {
		vf, _ := strconv.ParseFloat(v.Apdex, 32)
		compressStack = append(compressStack, vf)
		i = i + 1
		if i%compress == 0 || i == len(apdex) {
			var sum float64
			for _, f := range compressStack {
				sum = sum + f
			}
			compressed = append(compressed, sum/float64(len(compressStack)))
			compressStack = []float64{}
		}
	}

	var chart string
	for _, v := range compressed {
		var dot string
		if v < 0.125 {
			dot = "▁"
		} else if v < 0.250 {
			dot = "▂"
		} else if v < 0.375 {
			dot = "▃ "
		} else if v < 0.500 {
			dot = "▄ "
		} else if v < 0.625 {
			dot = "▅"
		} else if v < 0.750 {
			dot = "▆"
		} else if v < 0.875 {
			dot = "▇"
		} else {
			dot = "█"
		}
		chart = chart + dot
	}
	return chart
}
