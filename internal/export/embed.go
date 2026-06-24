// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package export

import (
	"embed"
)

//go:embed doc/*
var AllDocs embed.FS

//go:embed measurements_meta_i18n.json
var measurementsMetaI18nJSON []byte

var AllTemplates *embed.FS = nil
