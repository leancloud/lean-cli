package api

import (
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

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
}

// PrintLogs will poll the leanengine's log and print it to the giver io.Writer
func PrintLogs(writer io.Writer, appID string, masterKey string, follow bool, isProd bool) error {
	var url string
	var prod int

	limit := 100 // TODO
	params := map[string]string{}

	if !follow {
		params["limit"] = strconv.Itoa(limit)
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
			fmt.Fprintf(writer, "%s - %s\r\n", t.Local().Format("15:04:05"), content)
		}

		if !follow {
			break
		}

		if len(logs) > 0 {
			params["since"] = logs[0].Time
		}

		time.Sleep(5 * time.Second)

	}

	return nil
}
