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

// PrintLogsByLimit will poll the leanengine's log and print it to the giver io.Writer
func PrintLogsByLimit(printer LogPrinter, appID string, masterKey string, follow bool, isProd bool, limit int) error {
	params := map[string]string{
		"limit": strconv.Itoa(limit),
	}

	for {
		logs, err := FetchLogs(appID, masterKey, params, isProd)
		if err != nil {
			return err
		}
		for i := len(logs); i > 0; i-- {
			log := logs[i-1]

			err = printer(&log)
			if err != nil {
				fmt.Fprintf(os.Stderr, "error \"%v\" while parsing log: %s\r\n", err, log)
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

func FetchLogs(appID string, masterKey string, params map[string]string, isProd bool) ([]Log, error) {
	region, err := GetAppRegion(appID)
	if err != nil {
		return nil, err
	}

	var prod int
	if isProd {
		prod = 1
	} else {
		prod = 0
	}

	var url string
	switch region {
	case regions.CN:
		url = fmt.Sprintf("https://api.leancloud.cn/1.1/tables/EngineLogs?production=%d", prod)
	case regions.US:
		url = fmt.Sprintf("https://us-api.leancloud.cn/1.1/tables/EngineLogs?production=%d", prod)
	case regions.TAB:
		url = fmt.Sprintf("https://e1-api.leancloud.cn/1.1/tables/EngineLogs?production=%d", prod)
	}

	options := &grequests.RequestOptions{
		Headers: map[string]string{
			"X-AVOSCloud-Application-Id": appID,
			"X-AVOSCloud-Master-Key":     masterKey,
			"Content-Type":               "application/json",
		},
		Params: params,
	}

	var resp *grequests.Response
	var retryCount int = 0
	for {
		resp, err = grequests.Get(url, options)
		if err == nil {
			break
		}
		if retryCount >= 3 {
			return nil, err
		}
		retryCount++
		time.Sleep(1123 * time.Millisecond) // 1123 is a prime number, prime number makes less bugs.
	}

	var logs []Log
	err = resp.JSON(&logs)
	return logs, err
}
