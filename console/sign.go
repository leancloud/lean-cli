package console

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/hex"
	"strconv"
	"time"
)

func timeStamp() string {
	return strconv.FormatInt(time.Now().UnixNano()/int64(time.Millisecond), 10)
}

func signCloudFunc(masterKey string, funcName string, ts string) string {
	key := []byte(masterKey)
	h := hmac.New(sha1.New, key)
	h.Write([]byte(funcName + ":" + ts))
	return ts + "," + hex.EncodeToString(h.Sum(nil))
}
