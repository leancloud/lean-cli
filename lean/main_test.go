package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

var testUsername, testPassword, testRegion, testGroup, testAppID, repoURL string

func TestMain(m *testing.M) {
	testUsername, testPassword, testRegion = os.Getenv("TEST_USERNAME"), os.Getenv("TEST_PASSWORD"), os.Getenv("TEST_REGION")
	repoURL, testGroup, testAppID = os.Getenv("REPO_URL"), os.Getenv("TEST_GROUP"), os.Getenv("TEST_APPID")

	dir, err := ioutil.TempDir("", "*")
	if err != nil {
		fmt.Println(err)
		panic(err)
	}

	if err := os.Chdir(dir); err != nil {
		fmt.Println(err)
		panic(err)
	}

	gitExec, err := exec.LookPath("git")
	if err != nil {
		fmt.Println("can't find git executable file")
		panic(errors.New("can't find git executable file"))
	}
	if err := exec.Command(gitExec, "clone", repoURL, "lean-cli-deployment").Run(); err != nil {
		fmt.Println(err)
		panic(err)
	}

	gitDir := filepath.Join(dir, "lean-cli-deployment")
	if err := os.Chdir(gitDir); err != nil {
		fmt.Println(err)
		panic(err)
	}

	os.Exit(m.Run())

	defer os.RemoveAll(dir)

}

func TestLogin(t *testing.T) {
	os.Args = []string{"lean", "login", "--username", testUsername, "--password", testPassword, "--region", testRegion}
	main()
}

func TestSwitch(t *testing.T) {
	os.Args = []string{"lean", "switch", "--region", testRegion, "--group", testGroup, testAppID}
	main()
}

func TestDeploy(t *testing.T) {
	os.Args = []string{"lean", "deploy", "--prod", "0"}
	main()
}

func TestPublish(t *testing.T) {
	os.Args = []string{"lean", "publish"}
	main()
}
