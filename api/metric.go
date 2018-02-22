package api

import (
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/aisk/logp"
)

var (
	ErrNoEnoughData = errors.New("没有足够的数据")
)

type ReqStat struct {
	Date             string `json:"date"`
	ExceedTimes      int    `json:"exceed_times"`
	MaxDurationTime  int    `json:"max_duration_ms"`
	MeanQPS          int    `json:"meanQPS"`
	P95DurationTime  int    `json:"p95_duration_ms"`
	P90DurationTime  int    `json:"p90_duration_ms"`
	MaxConcurrent    int    `json:"max_concurrent"`
	MeanDurationTime int    `json:"mean_duration_ms"`
	// ExceptionPercentage string `json:"exception_percentage"`
	MeanConcurrent  int    `json:"mean_concurrent"`
	MaxQPS          int    `json:"max_qps"`
	P80DurationTime int    `json:"p80_duration_ms"`
	Error           string `json:"error,omitempty"`
	ApiReqCount     int    `json:"apiReqCount"`
}

type Status []ReqStat

func (S Status) Len() int {
	return len(S)
}

func (S Status) Swap(i, j int) {
	S[i], S[j] = S[j], S[i]
}

func (S Status) Less(i, j int) bool {
	flag := strings.Compare(S[i].Date, S[j].Date)
	if flag >= 0 {
		return false
	}
	return true
}

func FetchReqStat(appID string, from time.Time, to time.Time) (Status, error) {
	queryString := "?from=" + from.Format("20060102") + "&to=" + to.Format("20060102")
	appInfo, err := GetAppInfo(appID)
	if err != nil {
		return nil, err
	}
	logp.Info(fmt.Sprintf("正在获取 %s 储存报告", appInfo.AppName))
	client := NewClientByApp(appID)
	resp, err := client.get("/1.1/clients/self/apps/"+appID+"/reqStats"+queryString, nil)
	if err != nil {
		return nil, err
	}
	var js struct {
		Results map[string]ReqStat `json:"results"`
	}
	err = resp.JSON(&js)
	if err != nil {
		return nil, err
	}
	results := js.Results
	var status Status
	for date, item := range results {
		if item.Error != "" {
			return nil, ErrNoEnoughData
		}
		item.Date = date
		status = append(status, item)
	}
	sort.Sort(status)
	options, err := client.options()
	if err != nil {
		return nil, err
	}
	options.Headers["x-avoscloud-application-id"] = appInfo.AppID
	options.Headers["x-avoscloud-application-key"] = appInfo.AppKey
	resp, err = client.get("/1/statistics/details"+queryString, options)
	if err != nil {
		return nil, err
	}
	var jsapi struct {
		Results []int `json:"results"`
	}
	err = resp.JSON(&jsapi)
	if err != nil {
		return nil, err
	}
	index := 0
	for _, value := range jsapi.Results {
		if value == 0 {
			continue
		}
		status[index].ApiReqCount = value
		index++
	}
	return status, nil
}
