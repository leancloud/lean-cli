package boilerplate

import (
	"testing"
)

func TestGetBoilerplateList(t *testing.T) {
	_, err := GetBoilerplateList()
	if err != nil {
		t.Error(err)
	}
}
