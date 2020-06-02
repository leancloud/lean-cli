package api

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/leancloud/lean-cli/apps"
	"github.com/levigross/grequests"
)

// Log is EngineLogs's type structure
type Log struct {
	InstanceName string `json:"instanceName"`
	Content      string `json:"content"`
	Type         string `json:"type"`
	Time         string `json:"time"`
	GroupName    string `json:"groupName"`
	Production   int    `json:"prod"`
	Stream       string `json:"stream"`
	ID           string `json:"id"`
}

// LogReceiver is print func interface to PrintLogs
type LogReceiver func(*Log) error

// ReceiveLogsByLimit will poll the leanengine's log and print it to the giver io.Writer
func ReceiveLogsByLimit(printer LogReceiver, appID string, masterKey string, isProd bool, group string, limit int, follow bool) error {
	params := map[string]string{
		"limit": strconv.Itoa(limit),
		"prod":  "0",
		"group": group,
	}
	if isProd {
		params["prod"] = "1"
	}

	for {
		logs, err := fetchLogs(appID, masterKey, params, isProd)
		if err != nil {
			return err
		}
		for i := len(logs); i > 0; i-- {
			log := logs[i-1]

			err = printer(&log)
			if err != nil {
				fmt.Fprintf(os.Stderr, "error \"%v\" while parsing log: %v\r\n", err, log)
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

// ReceiveLogsByRange will poll the leanengine's log and print it to the giver io.Writer
func ReceiveLogsByRange(printer LogReceiver, appID string, masterKey string, isProd bool, group string, from time.Time, to time.Time) error {
	params := map[string]string{
		"ascend": "true",
		"since":  from.UTC().Format("2006-01-02T15:04:05.000000000Z"),
		"prod":   "0",
		"group":  group,
		"limit":  "1000",
	}

	if isProd {
		params["prod"] = "1"
	}

	for {
		logs, err := fetchLogs(appID, masterKey, params, isProd)
		if err != nil {
			return err
		}
		for _, log := range logs {
			logTime, err := time.Parse("2006-01-02T15:04:05.999999999Z", log.Time)
			if err != nil {
				return err
			}
			if to != (time.Time{}) && logTime.After(to) {
				// reached the end
				return nil
			}

			err = printer(&log)
			if err != nil {
				fmt.Fprintf(os.Stderr, "error \"%v\" while parsing log: %v\r\n", err, log)
			}
		}

		if len(logs) == 0 {
			// no more logs
			return nil
		}

		if len(logs) > 0 {
			params["since"] = logs[len(logs)-1].Time
		}
	}
}

func fetchLogs(appID string, masterKey string, params map[string]string, isProd bool) ([]Log, error) {
	region, err := apps.GetAppRegion(appID)
	if err != nil {
		return nil, err
	}

	url := NewClientByRegion(region).GetBaseURL() + "/1.1/engine/logs"

	options := &grequests.RequestOptions{
		Headers: map[string]string{
			"X-AVOSCloud-Application-Id": appID,
			"X-AVOSCloud-Master-Key":     masterKey,
			"Content-Type":               "application/json",
		},
		Params: params,
	}

	var resp *grequests.Response
	retryCount := 0
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
