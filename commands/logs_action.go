package commands

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/leancloud/lean-cli/api"
	"github.com/leancloud/lean-cli/apps"
	"github.com/urfave/cli"
)

func extractDateParams(c *cli.Context) (time.Time, time.Time, error) {
	from := time.Time{}
	var err error
	if c.String("from") != "" {
		from, err = time.ParseInLocation("2006-01-02", c.String("from"), time.Now().Location())
		if err != nil {
			err = fmt.Errorf("from format error: %s. The correct format is YYYY-MM-DD, e.g., 1926-08-17", c.String("from"))
			return time.Time{}, time.Time{}, err
		}
	}
	to := time.Time{}
	if c.String("to") != "" {
		to, err = time.ParseInLocation("2006-01-02", c.String("to"), time.Now().Location())
		if err != nil {
			err = fmt.Errorf("to format error: %s. The correct format is YYYY-MM-DD, e.g., 1926-08-17", c.String("to"))
			return time.Time{}, time.Time{}, err
		}
	}
	return from, to, nil
}

func logsAction(c *cli.Context) error {
	follow := c.Bool("f")
	env := c.String("e")
	limit := c.Int("limit")
	format := c.String("format")
	isProd := false

	groupName, err := apps.GetCurrentGroup(".")
	if err != nil {
		return err
	}

	from, to, err := extractDateParams(c)
	if err != nil {
		return err
	}

	if env == "staging" || env == "stag" {
		isProd = false
	} else if env == "production" || env == "" || env == "prod" {
		isProd = true
	} else {
		return cli.NewExitError("environment must be staging or production", 1)
	}

	appID, err := apps.GetCurrentAppID("")
	if err == apps.ErrNoAppLinked {
		return cli.NewExitError("Please use `lean checkout` designate a LeanCloud app first.", 1)
	}
	if err != nil {
		return err
	}
	info, err := api.GetAppInfo(appID)
	if err != nil {
		return err
	}

	var printer api.LogReceiver
	if format == "default" {
		printer = getDefaultLogPrinter(isProd)
	} else if strings.ToLower(format) == "json" {
		printer = jsonLogPrinter
	} else {
		return cli.NewExitError("format must be json or default.", 1)
	}

	if from != (time.Time{}) {
		return api.ReceiveLogsByRange(printer, info.AppID, info.MasterKey, isProd, groupName, from, to)
	}
	return api.ReceiveLogsByLimit(printer, info.AppID, info.MasterKey, isProd, groupName, limit, follow)
}

func getDefaultLogPrinter(isProd bool) api.LogReceiver {
	// 根据文档描述，有些类型的日志中的 production 字段，不论生产环境还是预备环境都会为 1，因此不能以此字段
	// 为依据来决定展示样式。
	return func(log *api.Log) error {
		t, err := time.Parse(time.RFC3339, log.Time)
		if err != nil {
			return err
		}
		content := strings.TrimSuffix(log.Content, "\n")
		stream := log.Stream
		var streamSprintf func(string, ...interface{}) string
		if stream == "stdout" {
			streamSprintf = color.New(color.BgGreen, color.FgWhite).SprintfFunc()
		} else {
			streamSprintf = color.New(color.BgRed, color.FgWhite).SprintfFunc()
		}
		var instance string
		if log.InstanceName == "" {
			instance = "    "
		} else {
			instance = log.InstanceName
		}

		if isProd {
			fmt.Fprintf(color.Output, "%s %s %s\r\n", instance, streamSprintf(" %s ", formatTime(&t)), content)
		} else {
			// no instance column
			fmt.Fprintf(color.Output, "%s %s\r\n", streamSprintf(" %s ", formatTime(&t)), content)
		}

		return nil
	}
}

func jsonLogPrinter(log *api.Log) error {
	content, err := json.Marshal(log)
	if err != nil {
		return err
	}
	fmt.Println(string(content))
	return nil
}

func isTimeInToday(t *time.Time) bool {
	now := time.Now()
	beginOfToday := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	endOfToday := beginOfToday.AddDate(0, 0, 1)
	return t.After(beginOfToday) && t.Before(endOfToday)
}

func formatTime(t *time.Time) string {
	if isTimeInToday(t) {
		return t.Local().Format("15:04:05")
	} else {
		return t.Local().Format("2006-01-02 15:04:05")
	}
}
