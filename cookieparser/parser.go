package cookieparser

import (
	"net/http"
)

// Parse raw cookie string
func Parse(raw string) []*http.Cookie {
	header := http.Header{}
    header.Add("Cookie", raw)
    request := http.Request{Header: header}
	return request.Cookies()
}