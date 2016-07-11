package main

import (
	"fmt"

	"github.com/codegangsta/cli"
	"github.com/leancloud/lean-cli/lean/api"
)

func whoamiAction(c *cli.Context) error {
	info, err := api.UserInfo()
	if err == api.ErrNotLogined {
		return cli.NewExitError("未登录，请先使用 `lean login` 命令登录 LeanCloud。", 1)
	}
	if err != nil {
		return cli.NewExitError(err.Error(), 1)
	}
	fmt.Printf("用户名: %s\r\n", info.Get("username").MustString())
	fmt.Printf("邮箱: %s\r\n", info.Get("email").MustString())
	return nil
}
