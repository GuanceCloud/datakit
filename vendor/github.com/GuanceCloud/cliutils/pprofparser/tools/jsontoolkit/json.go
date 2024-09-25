package jsontoolkit

import (
	"fmt"
	"net/http"

	"github.com/GuanceCloud/cliutils/pprofparser/tools/logtoolkit"
	"github.com/gin-gonic/gin"
)

type JSONResp struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

func JSONSuccess(ctx *gin.Context, data interface{}) {
	ctx.JSON(http.StatusOK, &JSONResp{
		Code:    0,
		Message: "success",
		Data:    data,
	})
}

func JSONError(ctx *gin.Context, code int, message string) {
	ctx.JSON(http.StatusOK, &JSONResp{
		Code:    code,
		Message: message,
		Data:    struct{}{},
	})
}

func JSONErrorf(ctx *gin.Context, code int, format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	logtoolkit.Error(msg)
	ctx.JSON(http.StatusOK, &JSONResp{
		Code:    code,
		Message: msg,
		Data:    struct{}{},
	})
}
