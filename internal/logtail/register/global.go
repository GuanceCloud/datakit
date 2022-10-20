// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package register wraps history cache functions
package register

import "errors"

var (
	ErrInvalidRegister = errors.New("invalid register")
	globalRegister     Register

	assertTesting = false
)

func Init(file string) error {
	if globalRegister != nil {
		return nil
	}
	r, err := NewRegisterFileIfNotExist(file)
	if err != nil {
		return err
	}
	globalRegister = r
	return nil
}

func AssertTesting() {
	assertTesting = true
}

func Set(key string, value *MetaData) error {
	if assertTesting {
		return nil
	}
	if globalRegister == nil {
		return ErrInvalidRegister
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
