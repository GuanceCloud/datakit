package prom

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
)

type GetReq func(map[string]string, string) (*http.Request, error)

var AuthMaps = map[string]GetReq{
	"bearer_token": BearerToken,
}

func BearerToken(auth map[string]string, url string) (*http.Request, error) {
	token, ok := auth["token"]
	if !ok {
		tokenFile, ok := auth["token_file"]
		if !ok {
			return nil, fmt.Errorf("invalid token")
		}
		tokenBytes, err := ioutil.ReadFile(tokenFile)
		if err != nil {
			return nil, fmt.Errorf("invalid token file")
		}
		token = string(tokenBytes)
		token = strings.Replace(token, "\n", "", -1)
	}
	req, err := http.NewRequest("GET", url, nil)
	if err == nil {
		req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", token))
	}
	return req, err
}
