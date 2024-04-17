package point

import types "github.com/gogo/protobuf/types"

func setVal(v isField_Val, raw any) isField_Val {
	switch f := v.(type) {
	case *Field_I:
		switch x := raw.(type) {
		case int8:
			f.I = int64(x)
		case int16:
			f.I = int64(x)
		case int32:
			f.I = int64(x)
		case int64:
			f.I = x
		case int:
			f.I = int64(x)
		}

	case *Field_U:
		switch x := raw.(type) {
		case uint8:
			f.U = uint64(x)
		case uint16:
			f.U = uint64(x)
		case uint32:
			f.U = uint64(x)
		case uint64:
			f.U = x
		case uint:
			f.U = uint64(x)
		}

	case *Field_F:
		switch x := raw.(type) {
		case float64:
			f.F = x
		case float32:
			f.F = float64(x)
		}

	case *Field_S:
		if x, ok := raw.(string); ok {
			f.S = x
		}

	case *Field_D:
		if x, ok := raw.([]byte); ok {
			f.D = f.D[:0] // reset
			f.D = append(f.D, x...)
		}

	case *Field_B:
		if x, ok := raw.(bool); ok {
			f.B = x
		}

	case *Field_A:
		var (
			arr *types.Any
			err error
		)

		switch x := raw.(type) {
		case []int8:
			arr, err = NewIntArray(x...)
		case []int16:
			arr, err = NewIntArray(x...)
		case []int32:
			arr, err = NewIntArray(x...)
		case []int64:
			arr, err = NewIntArray(x...)

		case []uint16:
			arr, err = NewUintArray(x...)
		case []uint32:
			arr, err = NewUintArray(x...)
		case []uint64:
			arr, err = NewUintArray(x...)

		case []string:
			arr, err = NewStringArray(x...)

		case []bool:
			arr, err = NewBoolArray(x...)

		case [][]byte:
			arr, err = NewBytesArray(x...)

		case *types.Any:
			arr = x
		default:
			// do nothing
		}

		if err == nil {
			f.A = arr
		}
	}

	// NOTE: if raw is nil or other types, do nothing
	return v
}

func newVal(v any) isField_Val {
	switch x := v.(type) {
	case int8:
		return &Field_I{int64(x)}
	case []int8:
		iarr, err := NewIntArray(x...)
		if err != nil {
			return nil
		} else {
			return &Field_A{iarr}
		}

	case uint8:
		return &Field_U{uint64(x)}
		// case []uint8 is []byte, skip it.

	case int16:
		return &Field_I{int64(x)}

	case []int16:
		iarr, err := NewIntArray(x...)
		if err != nil {
			return nil
		} else {
			return &Field_A{iarr}
		}

	case uint16:
		return &Field_U{uint64(x)}

	case []uint16:
		iarr, err := NewUintArray(x...)
		if err != nil {
			return nil
		} else {
			return &Field_A{iarr}
		}

	case int32:
		return &Field_I{int64(x)}

	case []int32:
		iarr, err := NewIntArray(x...)
		if err != nil {
			return nil
		} else {
			return &Field_A{iarr}
		}

	case uint32:
		return &Field_U{uint64(x)}

	case []uint32:
		iarr, err := NewUintArray(x...)
		if err != nil {
			return nil
		} else {
			return &Field_A{iarr}
		}

	case int:
		return &Field_I{int64(x)}

	case []int:
		iarr, err := NewIntArray(x...)
		if err != nil {
			return nil
		} else {
			return &Field_A{iarr}
		}

	case uint:
		return &Field_U{uint64(x)}

	case []uint:
		iarr, err := NewUintArray(x...)
		if err != nil {
			return nil
		} else {
			return &Field_A{iarr}
		}

	case int64:
		return &Field_I{x}

	case []int64:
		iarr, err := NewIntArray(x...)
		if err != nil {
			return nil
		} else {
			return &Field_A{iarr}
		}

	case uint64:
		return &Field_U{x}

	case []uint64:
		iarr, err := NewUintArray(x...)
		if err != nil {
			return nil
		} else {
			return &Field_A{iarr}
		}

	case float64:
		return &Field_F{x}

	case []float64:
		farr, err := NewFloatArray(x...)
		if err != nil {
			return nil
		} else {
			return &Field_A{farr}
		}

	case float32:
		return &Field_F{float64(x)}

	case []float32:
		farr, err := NewFloatArray(x...)
		if err != nil {
			return nil
		} else {
			return &Field_A{farr}
		}

	case string:
		return &Field_S{x}

	case []string:
		sarr, err := NewStringArray(x...)
		if err != nil {
			return nil
		} else {
			return &Field_A{sarr}
		}

	case []byte:
		return &Field_D{x}

	case bool:
		return &Field_B{x}

	case []bool:
		barr, err := NewBoolArray(x...)
		if err != nil {
			return nil
		} else {
			return &Field_A{barr}
		}

	case *types.Any:
		return &Field_A{x}

	case nil: // pass
		return nil

	default: // value ignored
		return nil
	}
}
