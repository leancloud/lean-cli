package commands

import (
	"time"

	"github.com/urfave/cli"
	"github.com/leancloud/lean-cli/api"
	"github.com/leancloud/lean-cli/apps"
	"encoding/json"
	"os"
)

func statusAction(c *cli.Context) error {
	fromPtr, toPtr, err := extractDateParams(c)
	if err != nil{
		return err
	}
	if fromPtr == nil{
		from := time.Now().Add(time.Duration(-1*7*24*60*60*1000*1000*1000)) // for a week
		fromPtr = &from
	}
	if toPtr == nil{
		to := time.Now()
		toPtr = &to
	}
	appID, err := apps.GetCurrentAppID("./")
	if err == apps.ErrNoAppLinked {
		return cli.NewExitError("没有关联任何 app，请使用 lean checkout 来关联应用。", 1)
	}
	ReqStats , err := api.FetchReqStat(appID, fromPtr.Format("20060102"), toPtr.Format("20060102"))
	if err != nil{
		return err
	}
	content, err := json.Marshal(ReqStats)
	if err != nil {
		return err
	}
	os.Stdout.Write(content)
	return nil
}