package api

import (
	"testing"

	"github.com/leancloud/lean-cli/api/regions"
)

func TestClient(t *testing.T) {
	f := func(r regions.Region) {
		client := NewClient(r)
		resp, err := client.get("/1.1/date", nil)
		if err != nil {
			t.FailNow()
		}
		if resp.StatusCode != 200 {
			t.FailNow()
		}
	}
	f(regions.CN)
	f(regions.US)
	f(regions.TAB)
}
