package api

import (
	"fmt"
	"encoding/json"
)

type ReqStat struct {
	Date string `json:"date"`
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

func FetchReqStat(appID string, from string, to string)([]ReqStat, error){
	queryString := "?from=" + from + "&to=" + to
	region, err := GetAppRegion(appID)
	if err != nil{
		panic(err)
	}
	client := NewClient(region)
	resp, err := client.get("/1.1/clients/self/apps/"+appID+"/reqStats"+queryString, nil)
	if err != nil{
		return nil, err
	}
	convertMap := map[string]string{
		"maxDurationTime":     "max_duration_ms",
		"meanQPS":             "mean_qps",
		"p95DurationTime":     "p95_duration_ms",
		"p90DurationTime":     "p90_duration_ms",
		"maxConcurrent":       "max_concurrent",
		"meanDurationTime":    "mean_duration_ms",
		// "exceptionPercentage": "exception_percentage",
		"meanConcurrent":      "mean_concurrent",
		"maxQPS":              "max_qps",
		"p80DurationTime":     "p80_duration_ms",
	}
	var js map[string]interface{}
	resp.JSON(&js)
	r := js["results"]
	results, _ := r.(map[string]interface{})
	var status []ReqStat
	for date, item := range results{
		item, _ := item.(map[string]interface{})
		js := make(map[string]string)
		js["date"] = date
		for k, v := range convertMap{
			s := fmt.Sprintf("%v",item[v])
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
	return status, nil
}