package utils

import (
	"testing"
)

func TestIsFileExists(t *testing.T) {
	if IsFileExists("/tmp") {
		t.Error("/tmp is a directory, should not exits")
	}

	if IsFileExists("a_invalid_file_path") {
		t.Error("a_invalid_file_path should not exits")
	}

	if !IsFileExists("file_exists_test.go") {
		t.Error("file_exists_test.go")
	}
}
