package harborMonitor

import (
	"crypto/tls"
	"io/ioutil"
	"net/http"
	"strings"
)

func Run(method, path, body string, headers map[string]string) (int, string) {
	if headers == nil {
		headers = map[string]string{}
	}
	headers["Content-Type"] = "application/json"

	method = strings.ToUpper(method)
	if method != "GET" && method != "POST" && method != "PUT" && method != "DELETE" {
		panic("Unsupported HTTP Method: " + method)
	}

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	httpClient := &http.Client{Transport: tr}

	req, err := http.NewRequest(method, path, strings.NewReader(body))
	if err != nil {
		panic(err)
	}

	for k, v := range headers {
		req.Header.Set(k, v)
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		panic(err)
	}

	defer resp.Body.Close()

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}

	respStatusCode := resp.StatusCode
	respData := string(respBody)

	return respStatusCode, respData
}

func Get(path string) (int, string) {
	return Run("GET", path, "", nil)
}
