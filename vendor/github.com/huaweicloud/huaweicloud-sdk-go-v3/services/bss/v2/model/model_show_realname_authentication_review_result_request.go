/*
 * BSS
 *
 * Business Support System API
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

// Request Object
type ShowRealnameAuthenticationReviewResultRequest struct {
	CustomerId string `json:"customer_id"`
}

func (o ShowRealnameAuthenticationReviewResultRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ShowRealnameAuthenticationReviewResultRequest struct{}"
	}

	return strings.Join([]string{"ShowRealnameAuthenticationReviewResultRequest", string(data)}, " ")
}
