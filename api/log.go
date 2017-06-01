package api

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/leancloud/lean-cli/api/regions"
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

// LogPrinter is print func interface to PrintLogs
type LogPrinter func(*Log) error

// PrintLogs will poll the leanengine's log and print it to the giver io.Writer
func PrintLogs(printer LogPrinter, appID string, masterKey string, follow bool, isProd bool, limit int) error {
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
	case regions.TAB:
		url = fmt.Sprintf("https://e1-api.leancloud.cn/1.1/tables/EngineLogs?production=%d", prod)
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
			time.Sleep(1100 * time.Millisecond)
			continue
		}

		var logs []Log
		err = resp.JSON(&logs)
		if err != nil {
			return err
		}

		for i := len(logs); i > 0; i-- {
			log := logs[i-1]

			err = printer(&log)
			if err != nil {
				fmt.Fprintf(os.Stderr, "error \"%v\" while parsing log: %s\r\n", err, resp)
			}
		}

		if !follow {
			break
		}

		// limit is not necessary in second round of fetch
		delete(params, "limit")

		if len(logs) > 0 {
			params["since"] = logs[0].Time
		}

		time.Sleep(5 * time.Second)
	}

	return nil
}
