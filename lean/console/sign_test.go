package console

import (
	"testing"
)

func TestSignCloudFunc(t *testing.T) {
	expected := "1468317699700,79ba1b1e1f8348a657151715d5ee4d46d60d0a4b"
	if signCloudFunc("aaa", "bbb", "1468317699700") != expected {
		t.Error("invalid sign")
	}
}
