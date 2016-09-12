package stats

import (
	"testing"
)

func TestGetDeviceID(t *testing.T) {
	deviceID, err := GetDeviceID()
	if err != nil {
		t.Error(err)
	}

	if deviceID == "" {
		t.Error("blank device ID")
	}

	anotherDeviceID, err := GetDeviceID()
	if err != nil {
		t.Error(err)
	}

	if deviceID != anotherDeviceID {
		t.Error("device ID not the same")
	}
}
