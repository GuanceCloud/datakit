// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package mapper contains graphite mapper action
package mapper

type ActionType string

const (
	ActionTypeMap     ActionType = "map"
	ActionTypeDrop    ActionType = "drop"
	ActionTypeDefault ActionType = ""
)
