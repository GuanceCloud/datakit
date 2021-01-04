package process

import (
	"fmt"
	"strconv"
	"strings"
)

func cast(v interface{}, tInfo string) interface{} {
	switch strings.ToLower(tInfo) {
	case "bool":
		switch vt := v.(type) {
		case int64, int32, int16, int8, uint64, uint32, uint16, uint8, int, uint:
			return vt != 0
		case string:
			lv := strings.ToLower(vt)
			if lv == "true" || lv == "t" {
				return true
			} else {
				return false
			}
		}

	case "int":
		switch vt := v.(type) {
		case int64, int32, int16, int8, uint64, uint32, uint16, uint8, int, uint:
			return vt
		case string:
			cv, _ := strconv.ParseUint(vt, 64, 64)
			return cv
		case float64:
			return int64(vt)
		case float32:
			return int64(vt)
		}
	case "float":
		switch vt := v.(type) {
		case int64, int32, int16, int8, uint64, uint32, uint16, uint8, int, uint:
			cv, _ := strconv.ParseFloat(fmt.Sprintf("%d", vt), 64)
			return cv
		case string:
			cv, _ := strconv.ParseFloat(vt, 64)
			return cv
		case float64, float32:
			return vt
		}

	case "str":
		return fmt.Sprintf("%v", v)
	}

	return nil
}
