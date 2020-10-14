package dataclean

import (
	"encoding/json"
	"fmt"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
)

const (
	objectCallbackFnName   = "handle"
	objectCallbackTypeName = "object"
)

type objectData struct {
	name     string
	category string
	data     interface{}
}

func NewObjectData(name string, category string, data []byte) (*objectData, error) {
	if name == "" {
		return nil, fmt.Errorf("invalid name, name is empty")
	}

	var v interface{}
	err := json.Unmarshal(data, &v)
	if err != nil {
		return nil, err
	}

	var j = objectData{
		name:     name,
		category: category,
		data:     v,
	}

	return &j, nil
}

func (j *objectData) Name() string {
	return j.name
}

func (j *objectData) DataToLua() interface{} {
	return j.data
}

func (*objectData) CallbackFnName() string {
	return objectCallbackFnName
}

func (*objectData) CallbackTypeName() string {
	return objectCallbackTypeName
}

func (j *objectData) Handle(value string, err error) {
	if err != nil {
		fmt.Printf("receive error: %v\n", err)
		return
	}

	err = io.Feed([]byte(value), j.category)
	if err != nil {
		l.Error(err)
	}
}
