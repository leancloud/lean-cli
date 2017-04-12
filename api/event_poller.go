package api

import (
	"github.com/aisk/chrysanthemum"
	"strings"
	"time"
)

type deployEvent struct {
	MoreEvent bool `json:"moreEvent"`
	Events    []struct {
		Content    string `json:"content"`
		Level      string `json:"level"`
		Production int    `json:"production"`
		Time       string `json:"time"`
	} `json:"events"`
}

// PollEvents will poll the server's event logs and print the result to the given io.Writer
func PollEvents(appID string, tok string) (bool, error) {
	region, err := GetAppRegion(appID)
	if err != nil {
		return false, err
	}
	client := NewClient(region)

	opts, err := client.options()
	if err != nil {
		return false, err
	}
	opts.Headers["X-LC-Id"] = appID

	from := ""
	ok := true
	retryCount := 0
	var spinner *chrysanthemum.Chrysanthemum
	for {
		time.Sleep(1 * time.Second)
		url := "/1.1/engine/events/poll/" + tok
		if from != "" {
			url = url + "?from=" + from
		}
		resp, err := client.get(url, opts)
		if err != nil {
			retryCount++
			if retryCount > 3 {
				return false, err
			}
			continue
		}
		event := new(deployEvent)
		err = resp.JSON(&event)
		if err != nil {
			return false, err
		}
		for i := len(event.Events) - 1; i >= 0; i-- {
			e := event.Events[i]

			if spinner != nil {
				if ok {
					spinner.Successed()
				} else {
					spinner.Failed()
				}
			}

			spinner = chrysanthemum.New("[REMOTE] " + e.Content).Start()
			from = e.Time
			ok = strings.ToLower(e.Level) != "error"
		}
		if !event.MoreEvent {
			break
		}
	}
	if spinner != nil {
		spinner.End()
	}
	return ok, nil
}
