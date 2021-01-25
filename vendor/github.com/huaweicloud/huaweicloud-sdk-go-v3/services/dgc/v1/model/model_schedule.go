/*
 * DGC
 *
 * 数据湖治理中心DGC是具有数据全生命周期管理、智能数据管理能力的一站式治理运营平台，支持行业知识库智能化建设，支持大数据存储、大数据计算分析引擎等数据底座，帮助企业快速构建从数据接入到数据分析的端到端智能数据系统，消除数据孤岛，统一数据标准，加快数据变现，实现数字化转型
 *
 */

package model

import (
	"encoding/json"
	"errors"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/converter"
	"strings"
)

type Schedule struct {
	ScheType *ScheduleScheType `json:"scheType,omitempty"`
	Cron     *Cron             `json:"cron,omitempty"`
	Event    *Event            `json:"event,omitempty"`
}

func (o Schedule) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "Schedule struct{}"
	}

	return strings.Join([]string{"Schedule", string(data)}, " ")
}

type ScheduleScheType struct {
	value string
}

type ScheduleScheTypeEnum struct {
	EXECUTE_ONCE ScheduleScheType
	CRON         ScheduleScheType
	EVENT        ScheduleScheType
}

func GetScheduleScheTypeEnum() ScheduleScheTypeEnum {
	return ScheduleScheTypeEnum{
		EXECUTE_ONCE: ScheduleScheType{
			value: "EXECUTE_ONCE",
		},
		CRON: ScheduleScheType{
			value: "CRON",
		},
		EVENT: ScheduleScheType{
			value: "EVENT",
		},
	}
}

func (c ScheduleScheType) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *ScheduleScheType) UnmarshalJSON(b []byte) error {
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
