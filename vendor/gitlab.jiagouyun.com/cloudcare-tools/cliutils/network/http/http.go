package http

import (
	"io/ioutil"
	"net/http"
)

// ReadBody will automatically unzip body
func ReadBody(req *http.Request) ([]byte, error) {
	buf, err := ioutil.ReadAll(req.Body)
	if err != nil {
		return nil, err
	}

	// as HTTP server, we do not need to close body
	switch req.Header.Get("Content-Encoding") {
	case "gzip":
		return Unzip(buf)
	default:
		return buf, err
	}
}
