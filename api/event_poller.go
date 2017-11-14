package api

import (
	"fmt"
	"os"
	"os/signal"
	"strings"
	"time"
	"sync"

	"github.com/aisk/logp"
	"github.com/fatih/color"
	"github.com/mattn/go-colorable"
)

var (
	m  = &sync.Mutex{}
	ch = make(chan os.Signal)
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

func monitorInterrupt(appId, eventTok string) {
	for i := 0; ; i++ {
		_, ok := <-ch
		if !ok {
			return
		}

		switch i {
		case 0:
			logp.Warn("正在取消部署...")
			go func() {
				m.Lock()
				err := CancelDeployByToken(appId, eventTok)
				if err != nil {
					logp.Error(err)
				} else {
					logp.Info("取消部署成功！")
				}
				m.Unlock()
			}()
		case 1:
			signal.Stop(ch)
			close(ch)
			os.Exit(1)
		}
	}
}

// PollEvents will poll the server's event logs and print the result to the given io.Writer
func PollEvents(appID string, tok string) (bool, error) {
	signal.Notify(ch, os.Interrupt)

	defer func() {
		signal.Stop(ch)
		close(ch)
	}()

	go monitorInterrupt(appID, tok)

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
			m.Lock()
			if ok {
				fmt.Fprintf(colorable.NewColorableStderr(), color.YellowString("[REMOTE] ")+e.Content+"\r\n")
			} else {
				fmt.Fprintf(colorable.NewColorableStderr(), color.YellowString("[REMOTE] ")+color.RedString("[ERROR] ")+e.Content+"\r\n")
			}
			m.Unlock()
		}
		if !event.MoreEvent {
			break
		}
	}
	return ok, nil
}
