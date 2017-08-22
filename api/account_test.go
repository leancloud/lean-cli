package api

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/leancloud/lean-cli/api/regions"
)

func fakeHome(t *testing.T) func() {
	originHome := os.Getenv("HOME")
	tmpHome, err := ioutil.TempDir("", "lea-cli-test-home-")
	if err != nil {
		t.Fatal(err)
	}
	os.Setenv("HOME", tmpHome)
	return func() {
		os.Setenv("HOME", originHome)
	}
}

func TestLogin(t *testing.T) {
	defer fakeHome(t)()

	_, err := Login("hife@amail.club", "A12345678", regions.CN)
	if err != nil {
		t.Error(err)
	}
	_, err = Login("hife@amail.club", "A12345678", regions.TAB)
	if err != nil {
		t.Error(err)
	}
	err = LoginUSRegion()
	if err != nil {
		t.Error(err)
	}

	loginedRegions, err := GetLoginedRegion()
	if err != nil {
		t.Fatal(err)
	}
	if len(loginedRegions) != 3 {
		t.Error()
	}
}
