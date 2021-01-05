package process

import (
	"strings"
	"github.com/tidwall/gjson"
)

func cast(result *gjson.Result, tInfo string) interface{} {
	switch strings.ToLower(tInfo) {
	case "bool":
		return result.Bool()

	case "int":
		return result.Int()

	case "float":
		return result.Float()

	case "str":
		return result.String()
	}

	return nil
}
