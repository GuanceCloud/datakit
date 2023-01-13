// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package register wraps history cache functions
package register

import (
	"fmt"
	"sync"
)

var (
	globalRegister Register

	initOnce sync.Once
	//nolint
	initErr error

	assertTesting = false
)

func Init(file string) error {
	initOnce.Do(func() {
		globalRegister, initErr = NewRegisterFileIfNotExist(file)
	})
	return initErr
}

func AssertTesting() {
	assertTesting = true
}

func Set(key string, value *MetaData) error {
	if assertTesting {
		return nil
	}
	if globalRegister == nil {
		return fmt.Errorf("invalid register")
	}
	return globalRegister.Set(key, value)
}

func Get(key string) *MetaData {
	if assertTesting {
		return nil
	}
	if globalRegister == nil {
		return nil
	}
	return globalRegister.Get(key)
}
