// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package point

import "errors"

var (
	ErrNoFields            = errors.New("no fields")
	ErrInvalidLineProtocol = errors.New("invalid lineprotocol")
)

// Point warnnings.
const (
	WarnMaxTags               = "exceed_max_tags"
	WarnMaxFields             = "exceed_max_fields"
	WarnNoField               = "no_field"
	WarnMaxTagKeyLen          = "exceed_max_tag_key_len"
	WarnMaxTagValueLen        = "exceed_max_tag_value_len"
	WarnMaxFieldKeyLen        = "exceed_max_field_key_len"
	WarnMaxFieldValueLen      = "exceed_max_field_value_len"
	WarnMaxFieldValueInt      = "exceed_max_field_value_int"
	WarnMaxTagKeyValueCompose = "exceed_max_tag_key_value_compose"

	WarnInvalidTagKey         = "invalid_tag_key"
	WarnInvalidTagValue       = "invalid_tag_value"
	WarnInvalidMeasurement    = "invalid_measurement"
	WarnInvalidFieldValueType = "invalid_field_value_type"
	WarnAddRequiredKV         = "add_required_kv"

	WarnFieldDisabled = "field_disabled"
	WarnTagDisabled   = "tag_disabled"

	WarnSameTagFieldKey = "same_tag_field_key"
	WarnDotInkey        = "dot_in_key"
	WarnFieldB64Encoded = "field_base64_encoded"
	WarnNilField        = "nil_field"
)
