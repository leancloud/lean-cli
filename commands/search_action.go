package commands

import (
	"fmt"
	"github.com/algolia/algoliasearch-client-go/algoliasearch"
	"github.com/urfave/cli"
	"strings"
)


func searchAction(c *cli.Context) error {
	if c.NArg() < 1 {
		noHelp := cli.ShowCommandHelp(c, "search")
		if noHelp != nil {
			return noHelp
		}
		return cli.NewExitError("", 1)
	}

	appID := "BH4D9OD16A"  // case sensitive
	apiKey := "357b777ed18e79673a2c1de3f6c64478"
	client := algoliasearch.NewClient(appID, apiKey)
	index := client.InitIndex("leancloud")

	params := algoliasearch.Map{
		"hitsPerPage": 20,
	}
	keyword := strings.Join(c.Args(), " ")
	res, err := index.Search(keyword, params)

	if err != nil {
		return err
	}

	displayHits(res)
	return nil
}

func displayHits(res algoliasearch.QueryRes) {
	if res.NbHits == 0 {
		fmt.Printf("No results found for query '%s'\n", res.Query)
	} else {
		for _, hit := range res.Hits {
			// We have to use `interface{}` because the value may contain `nil` or `string`.
			hierarchy := hit["hierarchy"].(map[string]interface{})
			for _, lvl := range []string{"lvl0", "lvl1", "lvl2", "lvl3", "lvl4", "lvl5", "lvl6"} {
				if hierarchy[lvl] == nil {
					fmt.Print("> ")
					break
				} else {
					fmt.Printf(" %s >", hierarchy[lvl])
				}
			}
			fmt.Println(hit["url"])
		}
	}
}


