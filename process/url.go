package process

import (
	"strings"
	"github.com/tidwall/gjson"
)

func urldecode(url string) (interface{}, error) {
	params, err := url.ParseQuery(url)
	if err != nil {
		return nil, err
	}

	return params, nil
}