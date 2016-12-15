package main

import (
	"encoding/json"
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"github.com/aisk/wizard"
	"github.com/chzyer/readline"
	"github.com/codegangsta/cli"
	"github.com/leancloud/lean-cli/lean/api"
	"github.com/leancloud/lean-cli/lean/apps"
	"github.com/leancloud/lean-cli/lean/rediscommands"
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

func selectDb() (int, error) {
	selectedDb := 0
	question := wizard.Question{
		Content: "请选择要操作 LeanCache 的 db （默认为 0）",
		Answers: []wizard.Answer{},
	}
	for i := 0; i < 16; i++ {
		answer := wizard.Answer{
			Content: fmt.Sprintf("db %d", i),
		}
		func(i int) {
			answer.Handler = func() {
				selectedDb = i
			}
		}(i)
		question.Answers = append(question.Answers, answer)
	}
	err := wizard.Ask([]wizard.Question{question})
	return selectedDb, err
}

func getRedisCommandCompleter() *readline.PrefixCompleter {
	var items []readline.PrefixCompleterInterface

	for _, c := range rediscommands.Commands {
		// ignore some unsupported command
		switch c {
		case "select":
			continue
		default:
			items = append(items, readline.PcItem(c))
		}
	}

	return readline.NewPrefixCompleter(items...)
}

func enterLeanCacheREPL(appID string, instance string, db int) error {
	l, err := readline.NewEx(&readline.Config{
		Prompt:          fmt.Sprintf("LeanCache (db %d) > ", db),
		HistoryFile:     filepath.Join(utils.ConfigDir(), "leancloud", "leancache_history"),
		AutoComplete:    getRedisCommandCompleter(),
		InterruptPrompt: "^C",
		EOFPrompt:       "quit",
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
		} else if line == "exit" || line == "quit" {
			break
		} else if err == io.EOF {
			break
		}

		line = strings.TrimSpace(line)

		if line == "" {
			continue
		} else if strings.HasPrefix(line, "select ") {
		}

		result, err := api.ExecuteCacheCommand(appID, instance, db, line)
		if e, ok := err.(api.Error); ok {
			fmt.Println(e.Content)
		} else if err != nil {
			return newCliError(err)
		} else {
			err = printCacheReult(result)
			if err != nil {
				return newCliError(err)
			}
		}
	}
	return nil
}

func printCacheReult(result *api.ExecuteCacheCommandResult) error {
	data, err := json.MarshalIndent(result.Result, "", "  ")
	if err != nil {
		return err
	}
	fmt.Println(string(data))
	return nil
}

func cacheAction(c *cli.Context) error {
	db := c.Int("db")
	instanceName := c.String("name")
	command := c.String("eval")

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

	if instanceName == "" {
		cache, err := selectCache(caches)
		if err != nil {
			return newCliError(err)
		}
		instanceName = cache.Instance
	}

	if db == -1 {
		db, err = selectDb()
		if err != nil {
			return newCliError(err)
		}
	}

	if command == "" {
		err = enterLeanCacheREPL(appID, instanceName, db)
		if err != nil {
			return newCliError(err)
		}
	} else {
		result, err := api.ExecuteCacheCommand(appID, instanceName, db, command)
		if e, ok := err.(api.Error); ok {
			fmt.Println(e.Content)
			return cli.NewExitError("", 1)
		} else if err != nil {
			return newCliError(err)
		} else {
			err = printCacheReult(result)
			if err != nil {
				return newCliError(err)
			}
		}
	}

	return nil
}
