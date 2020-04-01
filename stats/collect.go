package stats

import (
	"fmt"
	"runtime"

	"github.com/levigross/grequests"
)

const GA_TRACK_ID = "UA-42629236-12"

// Client is current user's client info
var Client ClientType

func Init() error {
	deviceID, err := GetDeviceID()
	if err != nil {
		return err
	}

	Client.ID = deviceID
	Client.Platform = runtime.GOOS
	return err
}

// Event is collect payload's evnets field type
type Event struct {
	Event string `json:"event"`
}

// ClientType is collect payload's client field type
type ClientType struct {
	ID         string `json:"id"`
	Platform   string `json:"platform"`
	AppVersion string `json:"app_version"`
	AppChannel string `json:"app_channel"`
}

// Collect the user's stats
func Collect(event Event) {
	_, err := grequests.Post("https://www.google-analytics.com/collect", &grequests.RequestOptions{
		Data: map[string]string{
			"aid": Client.Platform,
			"aiid": Client.AppChannel,
			"an": "lean",
			"av": Client.AppVersion,
			"cid": Client.ID,
			"ea": event.Event,
			"ec": "run",
			"t": "event",
			"tid": GA_TRACK_ID,
			"v": "1",
		},
	})
	if err != nil {
		fmt.Println("Failed to send statistics to Google Analytics.")
	}
}
