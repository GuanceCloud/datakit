// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package runtime

import (
	"github.com/GuanceCloud/platypus/pkg/errchain"
	"github.com/GuanceCloud/platypus/pkg/token"
)

func NewRunError(ctx *Context, err string, pos token.LnColPos) *errchain.PlError {
	return errchain.NewErr(ctx.name, pos, err)
}
