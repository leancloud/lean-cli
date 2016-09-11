package stats

import (
	"crypto/rand"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/leancloud/lean-cli/lean/utils"
)

func newDeviceID() (string, error) {
	uuid := make([]byte, 16)
	n, err := io.ReadFull(rand.Reader, uuid)
	if n != len(uuid) || err != nil {
		return "", err
	}
	uuid[8] = uuid[8]&^0xc0 | 0x80
	return fmt.Sprintf("%x-%x-%x-%x-%x", uuid[0:4], uuid[4:6], uuid[6:8], uuid[8:10], uuid[10:]), nil
}

// GetDeviceID returns the current machine's device ID
func GetDeviceID() (string, error) {
	err := os.MkdirAll(filepath.Join(utils.ConfigDir(), "leancloud"), 0755)
	if err != nil && !os.IsExist(err) {
		return "", err
	}

	var deviceID string
	path := filepath.Join(utils.ConfigDir(), "leancloud", "device_id")
	content, err := ioutil.ReadFile(path)
	if os.IsNotExist(err) {
		// generate new device id
		deviceID, err = newDeviceID()
		if err != nil {
			return "", err
		}
		err = ioutil.WriteFile(path, []byte(deviceID), 0644)
		if err != nil {
			return "", err
		}
	} else if err != nil {
		return "", err
	} else {
		deviceID = string(content)
	}
	return deviceID, nil
}
