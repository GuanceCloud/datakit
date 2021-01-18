package pipeline

import (
	"strings"

	conv "github.com/spf13/cast"
)

func cast(result interface{}, tInfo string) interface{} {
	switch strings.ToLower(tInfo) {
	case "bool":
		return conv.ToBool(result)

	case "int":
		return conv.ToInt64(conv.ToFloat64(result))

	case "float":
		return conv.ToFloat64(result)

	case "str":
		return conv.ToString(result)
	}

	return nil
}
