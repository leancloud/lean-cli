package api

import (
	"io"
	"os"
	"strings"
	"time"

	"github.com/leancloud/lean-cli/lean/output"
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
func PollEvents(appID string, tok string, writer io.Writer) (bool, error) {
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
	op := output.NewOutput(os.Stdout)
	for {
		time.Sleep(3 * time.Second)
		url := "/1.1/functions/_ops/events/poll/" + tok
		if from != "" {
			url = url + "?from=" + from
		}
		resp, err := client.get(url, opts)
		if err != nil {
			return false, err
		}
		event := new(deployEvent)
		err = resp.JSON(&event)
		if err != nil {
			return false, err
		}
		for i := len(event.Events) - 1; i >= 0; i-- {
			e := event.Events[i]

			op.Write(e.Content)

			from = e.Time
			if strings.ToLower(e.Level) == "error" {
				ok = false
			}
		}
		if !event.MoreEvent {
			op.Successed()
			break
		}
	}
	return ok, nil
}
