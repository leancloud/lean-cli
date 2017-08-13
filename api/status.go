package api

import (
	"fmt"
	"encoding/json"
	"strings"
	"sort"
	"errors"
)

var(
	ErrNoEnoughData = errors.New("the project has no enough data in this day.")
)

type ReqStat struct {
	Date string `json:"date"`
	ExceedTimes string `json:"exceedTimes"`
	MaxDurationTime string `json:"maxDurationTime"`
	MeanQPS string `json:"meanQPS"`
	P95DurationTime string `json:"p95DurationTime"`
	P90DurationTime string `json:"p90DurationTime"`
	MaxConcurrent string `json:"maxConcurrent"`
	MeanDurationTime string `json:"meanDurationTime"`
	// ExceptionPercentage string `json:"exception_percentage"`
	MeanConcurrent string `json:"meanConcurrent"`
	MaxQPS string `json:"maxQPS"`
	P80DurationTime string `json:"p80DurationTime"`
}


type Status []ReqStat

func (S Status)Len()int{
	return len(S)
}

func (S Status)Swap(i, j int){
	S[i], S[j] = S[j], S[i]
}

func (S Status)Less(i, j int) bool {
	flag := strings.Compare(S[i].Date, S[j].Date)
	if flag >= 0{
		return false
	}
	return true
}

func FetchReqStat(appID string, from string, to string)(Status, error){
	queryString := "?from=" + from + "&to=" + to
	region, err := GetAppRegion(appID)
	if err != nil{
		panic(err)
	}
	client := NewClient(region)
	resp, err := client.get("/1.1/clients/self/apps/"+appID+"/reqStats"+queryString, nil)
	if err != nil{
		return nil, errors.New("没有足够的数据。")
	}
	convertMap := map[string]string{
		"maxDurationTime":     "max_duration_ms",
		"meanQPS":             "mean_qps",
		"p95DurationTime":     "p95_duration_ms",
		"p90DurationTime":     "p90_duration_ms",
		"maxConcurrent":       "max_concurrent",
		"meanDurationTime":    "mean_duration_ms",
		// "exceptionPercentage": "exception_percentage",
		"exceedTimes":         "exceed_times",
		"meanConcurrent":      "mean_concurrent",
		"maxQPS":              "max_qps",
		"p80DurationTime":     "p80_duration_ms",
	}
	var js map[string]interface{}
	resp.JSON(&js)
	r := js["results"]
	results, _ := r.(map[string]interface{})
	var status Status
	for date, item := range results{
		item, _ := item.(map[string]interface{})
		if item["error"] != nil{
			return nil, ErrNoEnoughData
		}
		js := make(map[string]string)
		js["date"] = date
		for k, v := range convertMap{
			var s string
			if item[v] != nil {
				s = fmt.Sprintf("%v", item[v])
			} else {
				s = "0"
			}
			js[k] = s
		}
		s, err := json.Marshal(js)
		if err != nil{
			return nil, err
		}
		var field ReqStat
		json.Unmarshal(s, &field)
		status = append(status, field)
	}
	sort.Sort(status)
	return status, nil
}