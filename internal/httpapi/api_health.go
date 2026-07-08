// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package httpapi

import (
	"net/http"

	"github.com/gin-gonic/gin"
	componenthealth "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/health"
)

func healthHandler(registry *componenthealth.Registry) gin.HandlerFunc {
	return func(c *gin.Context) {
		snapshot := registry.Snapshot()
		statusCode := http.StatusOK
		if !snapshot.Live {
			statusCode = http.StatusServiceUnavailable
		}
		c.JSON(statusCode, snapshot)
	}
}
