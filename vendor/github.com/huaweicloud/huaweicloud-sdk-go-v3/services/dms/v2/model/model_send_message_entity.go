/*
 * DMS
 *
 * DMS Document API
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

type SendMessageEntity struct {
	// 消息正文。
	Body *interface{} `json:"body"`
	// 属性列表，包含属性名称和属性值。  同一条消息的属性名称不可重复，否则属性值将被覆盖。
	Attributes *interface{} `json:"attributes,omitempty"`
	// 消息标签，即Label，是通过对消息增加Label来区分队列中的消息分类，DMS允许消费者按照Label对消息进行过滤，确保消费者最终只消费到他关心的消息类型。  消息标签只能包含a~z，A~Z，0-9，-，_，长度是[1，64]。  最多可添加3个标签。
	Tags *interface{} `json:"tags,omitempty"`
	// 延时消息的延时时长。  延时消息是指消息发送到DMS服务后，并不期望这条消息立即被消费，而是延迟一段时间后才能被消费。  取值范围：0~604800000  单位：毫秒  不配置该参数或者配置为0，表示无延时。  配置为浮点数时，自动取小数点前面的整数值，比如配置为6000.9，则自动取值为6000。  仅NORMAL队列和FIFO队列可以设置延时消息，Kafka队列不支持延时消息的功能，如果向Kafka队列生产延时消息，提示{\"code\":10540010, \"message\":\"Invalid request format: kafka queue message could not have delayTime.\"}。
	DelayTime *interface{} `json:"delay_time,omitempty"`
}

func (o SendMessageEntity) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "SendMessageEntity struct{}"
	}

	return strings.Join([]string{"SendMessageEntity", string(data)}, " ")
}
