// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package man

import (
	"embed"
)

//go:embed docs/*
var AllDocs embed.FS

//go:embed dashboards/*
var AllDashboard embed.FS
