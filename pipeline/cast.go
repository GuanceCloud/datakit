package pipeline

import (
	"fmt"
	"strconv"
	"strings"
)

func cast(result interface{}, tInfo string) interface{} {
	switch strings.ToLower(tInfo) {
	case "bool":
		switch v := result.(type) {
		case bool:
			return v
		case int8, int16, int, int32, int64, uint8, uint16, uint, uint32, uint64:
			return v != 0
		case float32, float64:
			return v != 0
		case string:
			return v != "" && v != "0" && v != "false"
		default:
			return nil
		}

	case "int":
		switch v := result.(type) {
		case bool:
			if v {
				return 1
			} else {
				return 0
			}
		case int8, int16, int, int32, int64, uint8, uint16, uint, uint32, uint64:
			return v
		case float32:
			return int64(v)
		case float64:
			return int64(v)
		case string:
			if intV, err := strconv.ParseInt(v, 64, 64); err != nil {
				l.Error(err)
				return nil
			} else {
				return intV
			}
		default:
			return nil
		}

	case "float":
		switch v := result.(type) {
		case bool:
			if v {
				return float64(1)
			} else {
				return float64(0)
			}
		case string:
			n, _ := strconv.ParseFloat(v, 64)
			return n
		case float64, float32:
			return v
		case int8, int16, int, int32, int64, uint8, uint16, uint, uint32, uint64:
			return v
		}

	case "str":
		switch v := result.(type) {
		case string:
			return v
		default:
			return fmt.Sprintf("%v", v)
		}
	}

	return nil
}
