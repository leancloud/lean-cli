package commands

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"text/tabwriter"

	"github.com/chzyer/readline"
	"github.com/leancloud/lean-cli/api"
	"github.com/leancloud/lean-cli/apps"
	"github.com/leancloud/lean-cli/utils"
	"github.com/urfave/cli"
)

const (
	printCQLResultFormatInvalid = iota
	printCQLResultFormatTable
	printCQLResultFormatJSON
)

func enterCQLREPL(appInfo *api.GetAppInfoResult, format int) error {
	region, err := api.GetAppRegion(appInfo.AppID)
	if err != nil {
		return err
	}

	l, err := readline.NewEx(&readline.Config{
		Prompt:          "CQL > ",
		HistoryFile:     filepath.Join(utils.ConfigDir(), "leancloud", "cql_history"),
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
		} else if err == io.EOF {
			break
		}

		line = strings.TrimSpace(line)

		if line == "" {
			continue
		}
		if strings.HasSuffix(line, ";") {
			line = line[:len(line)-1]
		}

		result, err := api.ExecuteCQL(appInfo.AppID, appInfo.MasterKey, region, line)
		if err != nil {
			fmt.Println(err)
			continue
		}
		if format == printCQLResultFormatJSON {
			printJSONCQLResult(result)
		} else {
			printTableCQLResult(result)
		}
	}
	return nil
}

func printTableCQLResult(result *api.ExecuteCQLResult) {
	t := tabwriter.NewWriter(os.Stdout, 0, 1, 3, ' ', 0)
	if result.Count != -1 { // This is a count query
		fmt.Fprintf(t, "count\r\n")
		fmt.Fprintf(t, "%d\r\n", result.Count)
		t.Flush()
		return
	}

	if len(result.Results) == 0 {
		fmt.Println("** EMPTY **")
		return
	}

	keysSet := map[string]bool{
		"objectId":  true,
		"createdAt": true,
		"updatedAt": true,
	} // add this keys latter after sort
	keys := []string{}
	for _, obj := range result.Results {
		for key := range obj {
			if _, ok := keysSet[key]; !ok {
				keysSet[key] = true
				keys = append(keys, key)
			}
		}
	}
	sort.Strings(keys)

	// add this keys after sort
	keys = append([]string{"objectId"}, keys...)
	keys = append(keys, []string{"createdAt", "updatedAt"}...)

	// print table header
	fmt.Fprintf(t, "%s\r\n", strings.Join(keys, "\t"))

	for _, obj := range result.Results {
		for _, key := range keys {
			switch field := obj[key].(type) {
			case map[string]interface{}:
				if field["__type"] == "Date" {
					fmt.Fprintf(t, "%s\t", field["iso"])
				} else if field["__type"] == "GeoPoint" {
					fmt.Fprintf(t, "<GeoPoint(%v,%v)>\t", field["longitude"], field["latitude"])
				} else if field["__type"] == "Pointer" {
					if field["className"] == "_File" {
						fmt.Fprintf(t, "<File(%s)>\t", field["objectId"])
					} else {
						fmt.Fprintf(t, "<Pointer(%s:%s)>\t", field["className"], field["objectId"])
					}
				} else if field["__type"] == "Relation" {
					fmt.Fprintf(t, "<Relation>\t")
				} else {
					fmt.Fprintf(t, "<Object>\t")
				}
			case []interface{}:
				fmt.Fprintf(t, "<Array>\t")
			case nil:
				fmt.Fprintf(t, "<null>\t")
			default:
				fmt.Fprintf(t, "%v\t", obj[key])
			}
		}
		fmt.Fprintln(t)
	}
	t.Flush()
}

func printJSONCQLResult(result *api.ExecuteCQLResult) {
	if result.Count != -1 { // This is a count query
		encoded, err := json.MarshalIndent(map[string]interface{}{
			"count": result.Count,
		}, "", "  ")
		if err != nil {
			panic(err)
		}
		fmt.Printf("%s\r\n", encoded)
		return
	}

	encoded, err := json.MarshalIndent(result.Results, "", "  ")
	if err != nil {
		panic(err)
	}
	fmt.Printf("%s\r\n", encoded)
}

func cqlAction(c *cli.Context) error {
	eval := c.String("eval")
	format := printCQLResultFormatInvalid
	_format := c.String("format")
	switch _format {
	case "json", "JSON", "j", "J":
		format = printCQLResultFormatJSON
	case "table", "tab", "t", "T":
		format = printCQLResultFormatTable
	default:
		return cli.NewExitError("invalid format argument", 1)
	}

	appID, err := apps.GetCurrentAppID(".")
	if err != nil {
		return newCliError(err)
	}
	appInfo, err := api.GetAppInfo(appID)
	if err != nil {
		return newCliError(err)
	}

	if eval != "" {
		region, err := api.GetAppRegion(appInfo.AppID)
		if err != nil {
			return newCliError(err)
		}
		result, err := api.ExecuteCQL(appInfo.AppID, appInfo.MasterKey, region, eval)
		if err != nil {
			return newCliError(err)
		}
		if format == printCQLResultFormatJSON {
			printJSONCQLResult(result)
		} else {
			printTableCQLResult(result)
		}
	} else {
		err = enterCQLREPL(appInfo, format)
		if err != nil {
			return newCliError(err)
		}
	}
	return nil
}
