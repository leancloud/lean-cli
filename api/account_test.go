package api

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/leancloud/lean-cli/api/regions"
	"github.com/leancloud/lean-cli/apps"
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

	loginedRegions := apps.GetLoginedRegions()
	if len(loginedRegions) != 3 {
		t.Error()
	}
}
