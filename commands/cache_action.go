package commands

import (
	"encoding/json"
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"github.com/aisk/wizard"
	"github.com/chzyer/readline"
	"github.com/leancloud/lean-cli/api"
	"github.com/leancloud/lean-cli/apps"
	"github.com/leancloud/lean-cli/rediscommands"
	"github.com/leancloud/lean-cli/utils"
	"github.com/urfave/cli"
)

func selectDb() (int, error) {
	selectedDb := 0
	question := wizard.Question{
		Content: "Please choose a LeanCache DB (Default: 0)",
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

func printCacheReult(result *api.ExecuteCacheCommandResult) error {
	data, err := json.MarshalIndent(result.Result, "", "  ")
	if err != nil {
		return err
	}
	fmt.Println(string(data))
	return nil
}

func cacheAction(c *cli.Context) error {
	clusterName := c.String("name")
	db := c.Int("db")
	command := c.String("eval")

	appID, err := apps.GetCurrentAppID(".")
	if err != nil {
		return err
	}

	return runClusterAction(appID, clusterName, db, command)
}

func runClusterAction(appID string, clusterName string, db int, command string) error {
	clusters, err := api.GetClusterList(appID)

	var clusterID int

	if err != nil {
		return err
	}

	if len(clusters) == 0 {
		return cli.NewExitError("This app doesn't have any LeanDB instance", 1)
	}

	if clusterName == "" {
		instance, err := selectCluster(clusters)
		if err != nil {
			return err
		}
		clusterID = instance.ID
	} else {
		for _, cluster := range clusters {
			if cluster.Name == clusterName {
				clusterID = cluster.ID
			}
		}
	}

	if clusterID == 0 {
		return cli.NewExitError(fmt.Sprintf("LeanCache named %s is not found", clusterName), 1)
	}

	if db == -1 {
		db, err = selectDb()
		if err != nil {
			return err
		}
	}

	if command == "" {
		err = enterCacheREPL(appID, clusterID, db)
		if err != nil {
			return err
		}
	} else {
		result, err := api.ExecuteClusterCommand(appID, clusterID, db, command)
		if e, ok := err.(api.Error); ok {
			fmt.Println(e.Content)
			return cli.NewExitError("", 1)
		} else if err != nil {
			return err
		} else {
			err = printCacheReult(result)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func selectCluster(cacheList []*api.LeanCacheCluster) (*api.LeanCacheCluster, error) {
	var selectedInstance *api.LeanCacheCluster
	question := wizard.Question{
		Content: "Please choose a LeanCache instance",
		Answers: []wizard.Answer{},
	}
	for _, instance := range cacheList {
		answer := wizard.Answer{
			Content: fmt.Sprintf("%s - %s", instance.Name, instance.NodeQuota),
		}
		// for scope problem
		func(cache *api.LeanCacheCluster) {
			answer.Handler = func() {
				selectedInstance = cache
			}
		}(instance)
		question.Answers = append(question.Answers, answer)
	}
	err := wizard.Ask([]wizard.Question{question})
	return selectedInstance, err
}

func enterCacheREPL(appID string, clusterID int, db int) error {
	l, err := readline.NewEx(&readline.Config{
		Prompt:          fmt.Sprintf("LeanDB (db %d) > ", db),
		HistoryFile:     filepath.Join(utils.ConfigDir(), "leancloud", "leandb_history"),
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

		result, err := api.ExecuteClusterCommand(appID, clusterID, db, line)
		if e, ok := err.(api.Error); ok {
			fmt.Println(e.Content)
		} else if err != nil {
			return err
		} else {
			err = printCacheReult(result)
			if err != nil {
				return err
			}
		}
	}
	return nil
}
