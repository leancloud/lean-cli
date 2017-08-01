package api

import (
	"fmt"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/mattn/go-colorable"
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
	for {
		time.Sleep(700 * time.Millisecond)
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
			ok = strings.ToLower(e.Level) != "error"
			from = e.Time
			if ok {
				fmt.Fprintf(colorable.NewColorableStderr(), color.YellowString("[REMOTE] ")+e.Content+"\r\n")
			} else {
				fmt.Fprintf(colorable.NewColorableStderr(), color.YellowString("[REMOTE] ")+color.RedString("[ERROR] ")+e.Content+"\r\n")
			}
		}
		if !event.MoreEvent {
			break
		}
	}
	return ok, nil
}
