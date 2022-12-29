// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package runtime

import (
	"encoding/json"

	"github.com/GuanceCloud/platypus/pkg/ast"
	"github.com/spf13/cast"
)

func CastToStringE(val any, dtype ast.DType) (string, error) {
	switch dtype {
	case ast.String, ast.Bool, ast.Int, ast.Float:
		return cast.ToString(val), nil
	case ast.Map, ast.List:
		ret, err := json.Marshal(val)
		return string(ret), err
	case ast.Nil, ast.Void, ast.Invalid:
		return "", nil
	default:
		return "", nil
	}
}
