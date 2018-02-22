package upload

import (
	"bytes"
	"testing"
)

const (
	testAppID     = "pgk9e8orv8l9coak1rjht1avt2f4o9kptb0au0by5vbk9upb"
	testAppKey    = "hi4jsm62kok2qz2w2qphzryo564rzsrucl2czb0hn6ogwwnd"
	testAPIServer = "https://pgk9e8or.api.lncld.net"
)

func TestUpload(t *testing.T) {
	reader := bytes.NewReader([]byte("foobarbaz"))
	opts := &Options{
		AppID:     testAppID,
		AppKey:    testAppKey,
		APIServer: testAPIServer,
	}
	file, err := Upload("xxxooo.txt", "text/plain", reader, opts)
	if err != nil {
		t.Error(err)
		return
	}
	if file.ObjectID == "" {
		t.Error("invalid object id")
		return
	}
	if file.URL == "" {
		t.Error("invalid url")
		return
	}
}
