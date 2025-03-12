// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package funcs

import (
	"fmt"
	"reflect"
	"time"

	"github.com/GuanceCloud/cliutils/point"
	"github.com/GuanceCloud/pipeline-go/ptinput"
	"github.com/GuanceCloud/platypus/pkg/ast"
	"github.com/GuanceCloud/platypus/pkg/engine/runtime"
	"github.com/GuanceCloud/platypus/pkg/errchain"
	"github.com/spf13/cast"
)

func CreatePointChecking(ctx *runtime.Task, funcExpr *ast.CallExpr) *errchain.PlError {
	if err := normalizeFuncArgsDeprecated(funcExpr, []string{
		"name", "tags", "fields",
		"ts", "category", "after_use",
	}, 3); err != nil {
		return runtime.NewRunError(ctx, err.Error(), funcExpr.NamePos)
	}

	if arg := funcExpr.Param[5]; arg != nil {
		switch arg.NodeType { //nolint:exhaustive
		case ast.TypeStringLiteral:
			usecall := &ast.CallExpr{
				Name:    "use",
				NamePos: arg.StartPos(),
				Param:   []*ast.Node{funcExpr.Param[5]},
			}
			funcExpr.PrivateData = usecall
			return UseChecking(ctx, usecall)
		default:
			return runtime.NewRunError(ctx, fmt.Sprintf("param after_use expects StringLiteral, got %s",
				arg.NodeType), arg.StartPos())
		}
	}

	return nil
}

func CreatePoint(ctx *runtime.Task, funcExpr *ast.CallExpr) *errchain.PlError {
	ptIn, errg := getPoint(ctx.InData())
	if errg != nil {
		return nil
	}

	var ptName string
	var ptTags map[string]string
	ptFields := map[string]any{}
	ptCat := point.Metric

	name, _, err := runtime.RunStmt(ctx, funcExpr.Param[0])
	if err != nil {
		return err
	}

	if name == nil {
		return runtime.NewRunError(ctx,
			"type of parameter expected to be string, got nil",
			funcExpr.Param[0].StartPos())
	}

	if v, ok := name.(string); ok {
		ptName = v
	} else {
		return runtime.NewRunError(ctx, fmt.Sprintf(
			"type of parameter expected to be string, got %s",
			reflect.TypeOf(v)), funcExpr.Param[0].StartPos())
	}

	if tags, _, err := runtime.RunStmt(ctx, funcExpr.Param[1]); err != nil {
		return nil
	} else if tags != nil {
		if tKV, ok := tags.(map[string]any); ok {
			if len(tKV) > 0 {
				ptTags = map[string]string{}
				for tagK, tagV := range tKV {
					if val, ok := tagV.(string); ok {
						ptTags[tagK] = val
					}
				}
			}
		} else {
			return runtime.NewRunError(ctx, fmt.Sprintf(
				"type of parameter expected to be map, got %s",
				reflect.TypeOf(tags)), funcExpr.Param[1].StartPos())
		}
	}

	if fields, _, err := runtime.RunStmt(ctx, funcExpr.Param[2]); err != nil {
		return nil
	} else if fields != nil {
		if fKV, ok := fields.(map[string]any); ok {
			if len(fKV) > 0 {
				ptFields = map[string]any{}
				for fK, fV := range fKV {
					switch fV := fV.(type) {
					case int32, int8, int16, int,
						uint, uint16, uint32, uint64, uint8:
						ptFields[fK] = cast.ToInt64(fV)
					case float32:
						ptFields[fK] = cast.ToFloat64(fV)
					case []byte:
						ptFields[fK] = string(fV)
					case string, bool, float64, int64:
						ptFields[fK] = fV
					}
				}
			}
		} else {
			return runtime.NewRunError(ctx, fmt.Sprintf(
				"type of parameter expected to be map, got %s",
				reflect.TypeOf(fields)), funcExpr.Param[2].StartPos())
		}
	}

	var ptTime time.Time
	if arg := funcExpr.Param[3]; arg != nil {
		if pTS, _, err := runtime.RunStmt(ctx, arg); err != nil {
			return nil
		} else if pTS != nil {
			if ts, ok := pTS.(int64); ok {
				if ts > 0 {
					ptTime = time.Unix(0, ts)
				}
			} else {
				return runtime.NewRunError(ctx, fmt.Sprintf(
					"type of parameter expected to be int64, got %s",
					reflect.TypeOf(pTS)), arg.StartPos())
			}
		}
	}

	if arg := funcExpr.Param[4]; arg != nil {
		if catName, _, err := runtime.RunStmt(ctx, arg); err != nil {
			return nil
		} else if catName != nil {
			if catN, ok := catName.(string); ok {
				ptCat = ptCategory(catN)
				if ptCat == point.UnknownCategory {
					return nil
				}
			} else {
				return runtime.NewRunError(ctx, fmt.Sprintf(
					"type of parameter expected to be str, got %s",
					reflect.TypeOf(catName)), arg.StartPos())
			}
		}
	}

	if ptTime.IsZero() {
		ptTime = ptIn.PtTime()
	}

	plpt := ptinput.NewPlPt(ptCat, ptName, ptTags, ptFields, ptTime)
	if arg := funcExpr.Param[5]; arg != nil {
		if refCall, ok := funcExpr.PrivateData.(*ast.CallExpr); ok {
			if srcipt, ok := refCall.PrivateData.(*runtime.Script); ok {
				if err := srcipt.Run(plpt, ctx.Signal()); err != nil {
					return err.ChainAppend(ctx.Name(), funcExpr.NamePos)
				}
			}
		}
	}

	if ptIn != nil {
		ptIn.AppendSubPoint(plpt)
	}

	return nil
}

func ptCategory(cat string) point.Category {
	switch cat {
	case point.SLogging, point.CL:
		return point.Logging
	case point.SMetric, point.CM:
		return point.Metric
	case point.STracing, point.CT:
		return point.Tracing
	case point.SRUM, point.CR:
		return point.RUM
	case point.SNetwork, point.CN:
		return point.Network
	case point.SObject, point.CO:
		return point.Object
	case point.SCustomObject, point.CCO:
		return point.CustomObject
	case point.SSecurity, point.CS:
		return point.Security
	case point.SDialTesting, point.CDT:
		return point.DialTesting
	}
	return point.UnknownCategory
}
