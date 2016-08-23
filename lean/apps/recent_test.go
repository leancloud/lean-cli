package apps

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/leancloud/lean-cli/lean/api"
)

func TestSetRecentLinkedApp(t *testing.T) {
	dir, err := ioutil.TempDir("", "leancloud-test")
	if err != nil {
		t.Error(err)
	}
	defer os.RemoveAll(dir)

	if err = os.Mkdir(filepath.Join(dir, ".leancloud"), 0766); err != nil {
		t.Fatal(err)
	}

	test := func(appID string, expected string) {
		if err = setRecentLinkedApp(dir, appID); err != nil {
			t.Fatal(err)
		}

		content, err := ioutil.ReadFile(filepath.Join(dir, ".leancloud", "recent_linked_apps"))
		if err != nil {
			t.Fatal(err)
		}

		if string(content) != expected {
			t.Fatal(fmt.Sprintf("expected: %s, got: %s", expected, content))
		}
	}

	test("xxx", `["xxx"]`)
	test("xxx", `["xxx"]`)
	test("ooo", `["ooo","xxx"]`)
	test("xxx", `["xxx","ooo"]`)
	test("x01", `["x01","xxx","ooo"]`)
	test("x02", `["x02","x01","xxx","ooo"]`)
	test("x03", `["x03","x02","x01","xxx","ooo"]`)
	test("x04", `["x04","x03","x02","x01","xxx"]`)
	test("xxx", `["xxx","x04","x03","x02","x01"]`)
}

func TestMergeWithRecentApps(t *testing.T) {
	dir, err := ioutil.TempDir("", "leancloud-test")
	if err != nil {
		t.Error(err)
	}
	defer os.RemoveAll(dir)

	if err = os.Mkdir(filepath.Join(dir, ".leancloud"), 0766); err != nil {
		t.Fatal(err)
	}

	test := func(data []byte, appIDs []string, expected []string) {
		if err = ioutil.WriteFile(filepath.Join(dir, ".leancloud", "recent_linked_apps"), data, 0644); err != nil {
			t.Fatal(err)
		}

		var apps []*api.GetAppListResult
		for _, appID := range appIDs {
			apps = append(apps, &api.GetAppListResult{
				AppID: appID,
			})
		}
		result, err := MergeWithRecentApps(dir, apps)
		if err != nil {
			t.Fatal(err)
		}

		if len(result) != len(expected) {
			t.Fatalf("invalid appIDs count, expected: %d, actual: %d", len(expected), len(result))
		}

		for i := range result {
			if result[i].AppID != expected[i] {
				t.Fatalf("invalid appIDs, expected: %v, actual: %v", expected, result)
			}
		}

	}

	test([]byte(`[]`), []string{"a1", "a2"}, []string{"a1", "a2"})
	test([]byte(`["a1", "a2", "a3"]`), []string{"a1", "a2", "a3"}, []string{"a1", "a2", "a3"})
	test([]byte(`["a1", "a2", "a3"]`), []string{"a2", "a3", "a1"}, []string{"a1", "a2", "a3"})
	test([]byte(`["a1", "a2"]`), []string{"a3", "a2", "a1"}, []string{"a1", "a2", "a3"})
	test([]byte(`["a1", "a2", "invalid"]`), []string{"a3", "a2", "a1"}, []string{"a1", "a2", "a3"})
}
