package jsontoolkit

import (
	"encoding/json"
	"fmt"
	"strconv"
)

func IFaceCast2Int64(x interface{}) (int64, error) {
	if x == nil {
		return 0, fmt.Errorf("can not convert nil value to int64")
	}

	switch xx := x.(type) {
	case string:
		n, err := strconv.ParseInt(xx, 10, 64)
		if err != nil {
			return 0, fmt.Errorf("cannot convert profile.profile_start: [%s] to int64, err: %w", xx, err)
		}
		return n, nil

	case float64:
		return int64(xx), nil
	case json.Number:
		return xx.Int64()
	case int64:
		return xx, nil
	case uint64:
		return int64(xx), nil
	case int:
		return int64(xx), nil
	case uint:
		return int64(xx), nil
	case float32:
		return int64(xx), nil
	}

	return 0, fmt.Errorf("无法把interface{}类型: %T 转换为 int64", x)
}

func IFaceCast2String(x interface{}) (string, error) {
	if x == nil {
		return "", fmt.Errorf("can not convert nil value to string")
	}

	switch xx := x.(type) {
	case string:
		return xx, nil
	case float64:
		return strconv.FormatFloat(xx, 'g', -1, 64), nil
	case json.Number:
		return xx.String(), nil
	case int64:
		return strconv.FormatInt(xx, 10), nil
	}

	return "", fmt.Errorf("无法把interface{}类型: %T 转换为 string", x)
}
