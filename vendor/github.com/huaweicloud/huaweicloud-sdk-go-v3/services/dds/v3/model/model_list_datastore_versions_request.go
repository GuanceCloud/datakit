/*
 * DDS
 *
 * API v3
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
type ListDatastoreVersionsRequest struct {
	DatastoreName ListDatastoreVersionsRequestDatastoreName `json:"datastore_name"`
}

func (o ListDatastoreVersionsRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ListDatastoreVersionsRequest struct{}"
	}

	return strings.Join([]string{"ListDatastoreVersionsRequest", string(data)}, " ")
}

type ListDatastoreVersionsRequestDatastoreName struct {
	value string
}

type ListDatastoreVersionsRequestDatastoreNameEnum struct {
	DDS_COMMUNITY ListDatastoreVersionsRequestDatastoreName
}

func GetListDatastoreVersionsRequestDatastoreNameEnum() ListDatastoreVersionsRequestDatastoreNameEnum {
	return ListDatastoreVersionsRequestDatastoreNameEnum{
		DDS_COMMUNITY: ListDatastoreVersionsRequestDatastoreName{
			value: "DDS-Community",
		},
	}
}

func (c ListDatastoreVersionsRequestDatastoreName) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *ListDatastoreVersionsRequestDatastoreName) UnmarshalJSON(b []byte) error {
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
