// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2024-present Guance, Inc.

// Package consts define some constants needed for aws lambda.
package consts

// Extension.
const (
	ExtensionNameHeader       = "Lambda-Extension-Name"
	ExtensionIdentifierHeader = "Lambda-Extension-Identifier"
	ExtensionErrorType        = "Lambda-Extension-Function-Error-Type"
	ExtensionRoute            = "2020-01-01/extension"
)

// Telemetry.
const (
	TelemetrySubscriptionRoute = "2022-07-01/telemetry"
	TelemetryPort              = "8106"
)
