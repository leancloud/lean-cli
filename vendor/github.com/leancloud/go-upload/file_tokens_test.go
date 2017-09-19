package upload

import (
	"strings"
	"testing"
)

func TestGetFileKey(t *testing.T) {
	key := getFileKey("xxx")
	if len(key) != 40 {
		t.Error("invalid key length")
	}

	key = getFileKey("xxx.txt")
	if !strings.HasSuffix(key, ".txt") {
		t.Error("invalid key ext")
	}
}

func TestGetFileTokens(t *testing.T) {
	_, err := getFileTokens("what", "text/plain", 12345, &Options{
		AppID:  testAppID,
		AppKey: testAppKey,
	})
	if err != nil {
		t.Error(err)
	}
}
