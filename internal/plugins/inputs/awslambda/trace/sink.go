// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2024-present Guance, Inc.

package lambdatrace

import "context"

type Sink interface {
	Consume(ctx context.Context, spans []Span) error
}
