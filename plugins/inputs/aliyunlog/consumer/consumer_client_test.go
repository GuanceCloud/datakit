package consumerLibrary

import (
	"github.com/aliyun/aliyun-log-go-sdk"
	"testing"
)

func InitOption() LogHubConfig {
	return LogHubConfig{
		Endpoint:                  "",
		AccessKeyID:               "",
		AccessKeySecret:           "",
		Project:                   "",
		Logstore:                  "",
		ConsumerGroupName:         "",
		ConsumerName:              "",
		CursorPosition:            "",
		HeartbeatIntervalInSecond: 5,
	}
}

func client() *sls.Client {
	option := InitOption()
	return &sls.Client{
		Endpoint:        option.Endpoint,
		AccessKeyID:     option.AccessKeyID,
		AccessKeySecret: option.AccessKeySecret,
	}
}

func consumerGroup() sls.ConsumerGroup {
	return sls.ConsumerGroup{
		ConsumerGroupName: InitOption().ConsumerGroupName,
		Timeout:           InitOption().HeartbeatIntervalInSecond * 2,
	}
}

func TestConsumerClient_createConsumerGroup(t *testing.T) {

	type fields struct {
		option        LogHubConfig
		client        *sls.Client
		consumerGroup sls.ConsumerGroup
	}
	tests := []struct {
		name   string
		fields fields
	}{
		{"TestConsumerClient_createConsumerGroup", fields{InitOption(), client(), consumerGroup()}},
	}
	for _, tt := range tests {
		consumer := &ConsumerClient{
			option:        tt.fields.option,
			client:        tt.fields.client,
			consumerGroup: tt.fields.consumerGroup,
		}
		consumer.createConsumerGroup()
	}
}
