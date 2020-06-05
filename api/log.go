package api

import (
	"fmt"
	"os"
	"strconv"
	"time"

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
		"limit":     strconv.Itoa(limit),
		"prod":      "0",
		"groupName": group,
	}
	if isProd {
		params["prod"] = "1"
	}

	logIDSet := map[string]bool{}
	for {
		logs, err := fetchLogs(appID, masterKey, params, isProd)
		if err != nil {
			return err
		}

		for i := len(logs) - 1; i >= 0; i-- {
			log := logs[i]
			if _, ok := logIDSet[log.ID]; ok {
				continue
			}
			logIDSet[log.ID] = true
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
			params["to"] = logs[0].Time
		}
		params["from"] = time.Now().UTC().Format("2006-01-02T15:04:05.000000000Z")

		time.Sleep(5 * time.Second)
	}

	return nil
}

// ReceiveLogsByRange will poll the leanengine's log and print it to the giver io.Writer
func ReceiveLogsByRange(printer LogReceiver, appID string, masterKey string, isProd bool, group string, from time.Time, to time.Time) error {
	params := map[string]string{
		"prod":      "0",
		"groupName": group,
	}

	if isProd {
		params["prod"] = "1"
	}

	if from != (time.Time{}) {
		from = from.UTC()
		params["from"] = from.Format("2006-01-02T15:04:05.000000000Z")
	}
	if to != (time.Time{}) {
		to = to.UTC()
		params["to"] = to.Format("2006-01-02T15:04:05.000000000Z")
	}

	// 边界
	min := from
	max := to
	if from.After(to) {
		min, max = max, min
	}

	logIDSet := map[string]bool{}
	for {
		logs, err := fetchLogs(appID, masterKey, params, isProd)
		if err != nil {
			return err
		}

		if from.After(to) {
			// 反转，比倒序遍历好看点
			for i, j := 0, len(logs)-1; i < j; i, j = i+1, j-1 {
				logs[i], logs[j] = logs[j], logs[i]
			}
		}

		// 去重后的日志数量
		uniqueLogsCount := 0
		for i := 0; i < len(logs); i++ {
			log := logs[i]

			logTime, err := time.Parse("2006-01-02T15:04:05.999999999Z", log.Time)
			if err != nil {
				return err
			}

			if logTime.Before(min) || logTime.After(max) {
				return nil
			}

			if _, ok := logIDSet[log.ID]; ok {
				continue
			}
			logIDSet[log.ID] = true
			uniqueLogsCount++
			err = printer(&log)
			if err != nil {
				fmt.Fprintf(os.Stderr, "error \"%v\" while parsing log: %v\r\n", err, log)
			}
		}

		if uniqueLogsCount == 0 {
			return nil
		}

		if from != (time.Time{}) {
			params["from"] = logs[len(logs)-1].Time
		} else {
			params["to"] = logs[len(logs)-1].Time
		}
	}
}

func fetchLogs(appID string, masterKey string, params map[string]string, isProd bool) ([]Log, error) {
	client := NewClientByApp(appID)
	url := "/1.1/engine/logs"

	opts, err := client.options()
	if err != nil {
		return nil, err
	}
	opts.Headers["X-LC-Id"] = appID
	opts.Params = params

	var resp *grequests.Response
	retryCount := 0
	for {
		resp, err = client.get(url, opts)
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
