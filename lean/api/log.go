package api

import (
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/leancloud/lean-cli/lean/api/regions"
	"github.com/levigross/grequests"
)

// Log is EngineLogs's type structure
type Log struct {
	InstanceName string `json:"instanceName"`
	Content      string `json:"content"`
	Type         string `json:"type"`
	Time         string `json:"time"`
	GroupName    string `json:"groupName"`
	Production   int    `json:"production"`
	OID          string `json:"oid"`
	Level        string `json:"level"`
	Instance     string `json:"instance"`
}

// PrintLogs will poll the leanengine's log and print it to the giver io.Writer
func PrintLogs(writer io.Writer, appID string, masterKey string, follow bool, isProd bool, limit int) error {
	var url string
	var prod int

	params := map[string]string{
		"limit": strconv.Itoa(limit),
	}

	if isProd {
		prod = 1
	} else {
		prod = 0
	}

	region, err := GetAppRegion(appID)
	if err != nil {
		return err
	}

	switch region {
	case regions.CN:
		url = fmt.Sprintf("https://api.leancloud.cn/1.1/tables/EngineLogs?production=%d", prod)
	case regions.US:
		url = fmt.Sprintf("https://us-api.leancloud.cn/1.1/tables/EngineLogs?production=%d", prod)
	}

	retryCount := 0

	for {
		resp, err := grequests.Get(url, &grequests.RequestOptions{
			Headers: map[string]string{
				"X-AVOSCloud-Application-Id": appID,
				"X-AVOSCloud-Master-Key":     masterKey,
				"Content-Type":               "application/json",
			},
			Params: params,
		})
		if err != nil {
			retryCount++
			if retryCount > 3 {
				return err
			}
			time.Sleep(5 * time.Second)
			continue
		}

		var logs []Log
		err = resp.JSON(&logs)
		if err != nil {
			return err
		}

		for i := len(logs); i > 0; i-- {
			log := logs[i-1]
			t, err := time.Parse(time.RFC3339, log.Time)
			if err != nil {
				return err
			}
			content := strings.TrimSuffix(log.Content, "\n")
			// fmt.Println(log)
			level := log.Level
			var levelSprintf func(string, ...interface{}) string
			if level == "info" {
				levelSprintf = color.New(color.BgGreen, color.FgWhite).SprintfFunc()
			} else {
				levelSprintf = color.New(color.BgRed, color.FgWhite).SprintfFunc()
			}
			var instance string
			if log.Instance == "" {
				instance = "    "
			} else {
				instance = log.Instance
			}

			if isProd {
				fmt.Fprintf(writer, "%s %s %s\r\n", instance, levelSprintf(" %s ", t.Local().Format("15:04:05")), content)
			} else {
				// no instance column
				fmt.Fprintf(writer, "%s %s\r\n", levelSprintf(" %s ", t.Local().Format("15:04:05")), content)
			}
		}

		if !follow {
			break
		}

		// limit is not necessary in second fetch
		delete(params, "limit")

		if len(logs) > 0 {
			params["since"] = logs[0].Time
		}

		time.Sleep(5 * time.Second)

	}

	return nil
}
