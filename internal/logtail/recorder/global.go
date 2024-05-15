// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package recorder

import (
	"fmt"
	"sync"
)

var (
	globalRecorder Recorder
	initOnce       sync.Once
)

func Init(file string) error {
	var err error
	initOnce.Do(func() {
		r, _err := NewRecorder(file)
		if _err == nil {
			globalRecorder = r
			globalRecorder.Clean()
		}
		err = _err
	})
	return err
}

func Set(key string, value *MetaData) error {
	if globalRecorder == nil {
		return fmt.Errorf("invalid recorder")
	}
	return globalRecorder.Set(key, value)
}

func Get(key string) *MetaData {
	if globalRecorder == nil {
		return nil
	}
	return globalRecorder.Get(key)
}

func SetAndFlush(key string, value *MetaData) error {
	if globalRecorder == nil {
		return fmt.Errorf("invalid recorder")
	}
	if err := globalRecorder.Set(key, value); err != nil {
		return err
	}
	return globalRecorder.Flush()
}
