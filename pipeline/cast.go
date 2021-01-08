package pipeline

import (
	"strings"
)

func cast(result interface{}, tInfo string) interface{} {
	switch strings.ToLower(tInfo) {
	case "bool":
		switch v := result.(type) {
		case bool:
			return v
		case int,int32,int64,int8,int16:
			return v !=0

		}

	case "int":
		return 1

	case "float":
		return 1.1

	case "str":
		return "result.String()"
	}

	return nil
}
