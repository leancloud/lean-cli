package main

import (
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"github.com/aisk/wizard"
	"github.com/chzyer/readline"
	"github.com/codegangsta/cli"
	"github.com/leancloud/lean-cli/lean/api"
	"github.com/leancloud/lean-cli/lean/apps"
	"github.com/leancloud/lean-cli/lean/utils"
)

func selectCache(cacheList []*api.GetCacheListResult) (*api.GetCacheListResult, error) {
	var selectedCache *api.GetCacheListResult
	question := wizard.Question{
		Content: "请选择 LeanCache 实例",
		Answers: []wizard.Answer{},
	}
	for _, cache := range cacheList {
		answer := wizard.Answer{
			Content: fmt.Sprintf("%s - %dM", cache.Instance, cache.MaxMemory),
		}
		// for scope problem
		func(cache *api.GetCacheListResult) {
			answer.Handler = func() {
				selectedCache = cache
			}
		}(cache)
		question.Answers = append(question.Answers, answer)
	}
	err := wizard.Ask([]wizard.Question{question})
	return selectedCache, err
}

func enterLeanCacheREPL(appID string, instance string, db int) error {
	l, err := readline.NewEx(&readline.Config{
		Prompt:      "LeanCache >> ",
		HistoryFile: filepath.Join(utils.ConfigDir(), "leancloud", "leancache_history"),
		// AutoComplete:    completer,
		InterruptPrompt: "^C",
		EOFPrompt:       "exit",
	})
	if err != nil {
		return err
	}
	defer l.Close()

	for {
		line, err := l.Readline()
		if err == readline.ErrInterrupt {
			if len(line) == 0 {
				break
			} else {
				continue
			}
		} else if err == io.EOF {
			break
		}

		line = strings.TrimSpace(line)

		result, err := api.ExecuteCacheCommand(appID, instance, db, line)
		if err != nil {
			return newCliError(err)
		}

		fmt.Println(result.Result)
	}
	return nil
}

func cacheAction(c *cli.Context) error {
	db := c.Int("db")

	appID, err := apps.GetCurrentAppID(".")
	if err != nil {
		return newCliError(err)
	}

	caches, err := api.GetCacheList(appID)
	if err != nil {
		return newCliError(err)
	}

	if len(caches) == 0 {
		return cli.NewExitError("该应用没有 LeanCache 实例", 1)
	}

	cache, err := selectCache(caches)
	if err != nil {
		return newCliError(err)
	}

	err = enterLeanCacheREPL(appID, cache.Instance, db)
	if err != nil {
		return newCliError(err)
	}

	return nil
}
