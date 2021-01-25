/*
 * SWR
 *
 * SWR API
 *
 */

package model

import (
	"encoding/json"
	"errors"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/converter"
	"strings"
)

// Request Object
type DeleteImageSyncRepoRequest struct {
	ContentType DeleteImageSyncRepoRequestContentType `json:"Content-Type"`
	Namespace   string                                `json:"namespace"`
	Repository  string                                `json:"repository"`
	Body        *DeleteImageSyncRepoRequestBody       `json:"body,omitempty"`
}

func (o DeleteImageSyncRepoRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "DeleteImageSyncRepoRequest struct{}"
	}

	return strings.Join([]string{"DeleteImageSyncRepoRequest", string(data)}, " ")
}

type DeleteImageSyncRepoRequestContentType struct {
	value string
}

type DeleteImageSyncRepoRequestContentTypeEnum struct {
	APPLICATION_JSONCHARSETUTF_8 DeleteImageSyncRepoRequestContentType
	APPLICATION_JSON             DeleteImageSyncRepoRequestContentType
}

func GetDeleteImageSyncRepoRequestContentTypeEnum() DeleteImageSyncRepoRequestContentTypeEnum {
	return DeleteImageSyncRepoRequestContentTypeEnum{
		APPLICATION_JSONCHARSETUTF_8: DeleteImageSyncRepoRequestContentType{
			value: "application/json;charset=utf-8",
		},
		APPLICATION_JSON: DeleteImageSyncRepoRequestContentType{
			value: "application/json",
		},
	}
}

func (c DeleteImageSyncRepoRequestContentType) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *DeleteImageSyncRepoRequestContentType) UnmarshalJSON(b []byte) error {
	myConverter := converter.StringConverterFactory("string")
	if myConverter != nil {
		val, err := myConverter.CovertStringToInterface(strings.Trim(string(b[:]), "\""))
		if err == nil {
			c.value = val.(string)
			return nil
		}
		return err
	} else {
		return errors.New("convert enum data to string error")
	}
}
