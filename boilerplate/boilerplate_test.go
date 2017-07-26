package boilerplate

import (
	"testing"
)

func TestGetBoilerplates(t *testing.T) {
	_, err := GetBoilerplates()
	if err != nil {
		t.Error(err)
	}
}
