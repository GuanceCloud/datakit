// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2026-present Guance, Inc.

package plval

import (
	"sync"

	"github.com/GuanceCloud/pipeline-go/ptinput/funcs"
	"github.com/GuanceCloud/platypus/pkg/ast"
	plruntime "github.com/GuanceCloud/platypus/pkg/engine/runtime"
	"github.com/GuanceCloud/platypus/pkg/errchain"
)

var installDisableGrokFastPathOnce sync.Once

func installDisableGrokFastPath() {
	installDisableGrokFastPathOnce.Do(func() {
		grokChecking := funcs.FuncsCheckMap["grok"]
		if grokChecking == nil {
			grokChecking = funcs.GrokChecking
		}

		funcs.FuncsCheckMap["grok"] = func(ctx *plruntime.Task, funcExpr *ast.CallExpr) *errchain.PlError {
			if err := grokChecking(ctx, funcExpr); err != nil {
				return err
			}
			if funcExpr != nil && funcExpr.Grok != nil && !isGrokFastPathEnabled() {
				funcExpr.Grok.DisableFastPath()
			}
			return nil
		}
	})
}
