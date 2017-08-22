package commands

import (
	"text/tabwriter"
	"time"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/leancloud/lean-cli/api"
	"github.com/leancloud/lean-cli/apps"
	"github.com/urfave/cli"
)


type metricPrinter func(api.Status) error

func parseDate(d string) string {
	tmp, _ := time.Parse("20060102", d)
	return tmp.Format("2006-01-02")
}

func jsonMetricPrinter(status api.Status) error {
	content, err := json.MarshalIndent(status, "", "  ")
	if err != nil {
		return err
	}
	fmt.Println(string(content))
	return nil
}

func statusPrinter(status api.Status) error {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
	if status.Len() > 8 {
		fmt.Fprintln(w, "Date\tAPI Requests\tMax Concurrent\tMean Concurrent\tExceed Time\tMax QPS\tMean Duration Time\t80% Duration Time\t95% Duration Time\t")
		for _, item := range status {
			fmt.Fprintln(w, fmt.Sprintf(
				"%v\t%v\t%v\t%v\t%v\t%v\t%vms\t%vms\t%vms\t",
				parseDate(item.Date), item.ApiReqCount, item.MaxConcurrent, item.MeanConcurrent,
				item.ExceedTimes, item.MaxQPS, item.MeanDurationTime,
				item.P80DurationTime, item.P95DurationTime),
			)
		}
	} else {
		formatString := strings.Repeat("%v\t", status.Len()+1)
		fieldTitle := []string{
			"Date", "API Requests", "Max Concurrent", "Mean Concurrent", "Exceed Time", "Max QPS", "Mean Duration Time",
			"80% Duration Time", "95% Duration Time",
		}
		for _, field := range fieldTitle {
			var printString []interface{}
			printString = append(printString, field)
			for _, item := range status {
				switch field {
				case "Date":
					printString = append(printString, parseDate(item.Date))
				case "Max Concurrent":
					printString = append(printString, item.MaxConcurrent)
				case "Mean Concurrent":
					printString = append(printString, item.MeanConcurrent)
				case "Exceed Time":
					printString = append(printString, item.ExceedTimes)
				case "Max QPS":
					printString = append(printString, item.MaxQPS)
				case "Mean Duration Time":
					printString = append(printString, strconv.Itoa(item.MeanDurationTime)+"ms")
				case "80% Duration Time":
					printString = append(printString, strconv.Itoa(item.P80DurationTime)+"ms")
				case "95% Duration Time":
					printString = append(printString, strconv.Itoa(item.P95DurationTime)+"ms")
				case "API Requests":
					printString = append(printString, item.ApiReqCount)
				}
			}
			fmt.Fprintln(w, fmt.Sprintf(formatString, printString...))
		}
	}
	w.Flush()
	return nil
}

func statusAction(c *cli.Context) error {
	from, to, err := extractDateParams(c)
	if err != nil {
		return err
	}
	if from == (time.Time{}) {
		from = time.Now().Add(time.Duration(-1 * 7 * 24 * time.Hour))
	}
	if to == (time.Time{}) {
		to = time.Now()
	}
	appID, err := apps.GetCurrentAppID("./")
	if err == apps.ErrNoAppLinked {
		return cli.NewExitError("没有关联任何 app，请使用 lean checkout 来关联应用。", 1)
	}
	ReqStats, err := api.FetchReqStat(appID, from, to)
	if err != nil {
		return err
	}
	var p metricPrinter
	switch c.String("format") {
	case "":
		fallthrough
	case "default":
		p = statusPrinter
	case "json":
		p = jsonMetricPrinter
	}
	err = p(ReqStats)
	if err != nil {
		return err
	}
	return nil
}
