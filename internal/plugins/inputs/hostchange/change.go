// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

//go:build !windows
// +build !windows

package hostchange

import (
	"fmt"
	"time"
)

type ChangeKey string

// Change represents a configuration change entity.
type Change[K comparable, T any] struct {
	LastUpdateTime time.Time
	ContentHash    uint64 // hash of the content to quickly compare changes
	RawKey         string // original key before any processing

	isValid  bool                // skip compare if not valid
	data     map[K][]T           // change data
	children map[K]*Change[K, T] // child changes
}

type ChangeOption[K comparable, T any] func(*Change[K, T])

// NewChange create a new Change instance.
func NewChange[K comparable, T any](opts ...ChangeOption[K, T]) *Change[K, T] {
	change := &Change[K, T]{
		data:           make(map[K][]T),
		children:       make(map[K]*Change[K, T]),
		isValid:        true,
		LastUpdateTime: time.Now(),
		ContentHash:    0,
	}
	for _, opt := range opts {
		opt(change)
	}

	return change
}

func (c *Change[K, T]) IsValid() bool {
	return c.isValid
}

// SetValid set change valid status to skip or not skip compare.
func (c *Change[K, T]) SetValid(valid bool) {
	c.isValid = valid
}

// Add add value to change data.
func (c *Change[K, T]) Add(key K, value T) {
	c.data[key] = append(c.data[key], value)
}

// SetChild set child change.
func (c *Change[K, T]) SetChild(key K, child *Change[K, T]) {
	c.children[key] = child
}

// GetChild get child change by key.
func (c *Change[K, T]) GetChild(key K) *Change[K, T] {
	return c.children[key]
}

type GetChangeItemFunc[K comparable, T any] func(key K, newValues, oldValues []T, parentChanges ...*Change[K, T]) *ChangeItem

type ChangeItemConfig[K comparable, T any] struct {
	Modify,
	Delete,
	Add GetChangeItemFunc[K, T]
}

func (c *Change[K, T]) GetChangeEvent(oldChange *Change[K, T],
	changeItemConfig *ChangeItemConfig[K, T], parentChanges ...*Change[K, T],
) ([]*ChangeItem, error) {
	changeItems := make([]*ChangeItem, 0)

	if oldChange != nil && !oldChange.IsValid() {
		return nil, fmt.Errorf("oldChange is not valid")
	}

	if changeItemConfig == nil {
		return nil, fmt.Errorf("changeItemConfig is nil")
	}

	// 1. compare data
	// 1.1 compare new and modify
	for key, newValues := range c.data {
		isNew := false
		if oldChange == nil {
			isNew = true
		} else {
			if oldValues, ok := oldChange.data[key]; ok { // modify
				if changeItemConfig.Modify != nil {
					if item := changeItemConfig.Modify(key, newValues, oldValues, c); item != nil {
						changeItems = append(changeItems, item)
					}
				}
			} else { // new
				isNew = true
			}
		}

		if isNew && changeItemConfig.Add != nil {
			if item := changeItemConfig.Add(key, newValues, nil, c); item != nil {
				changeItems = append(changeItems, item)
			}
		}
	}

	// 1.2 compare delete
	if oldChange != nil {
		for key, oldValues := range oldChange.data {
			if _, ok := c.data[key]; !ok { // delete
				if changeItemConfig.Delete != nil {
					if item := changeItemConfig.Delete(key, nil, oldValues, c); item != nil {
						changeItems = append(changeItems, item)
					}
				}
			}
		}
	}

	// 2. compare children
	if oldChange != nil {
		// 2.1 compare new and modify
		for key, newChildren := range c.children {
			if oldChildren, ok := oldChange.children[key]; ok { // modify
				if newChildren.ContentHash != 0 && newChildren.ContentHash == oldChildren.ContentHash {
					continue
				}
				childEvents, err := newChildren.GetChangeEvent(oldChildren, changeItemConfig, append(parentChanges, c)...)
				if err != nil {
					return nil, err
				}
				changeItems = append(changeItems, childEvents...)
			} else { // new
				childEvents, err := newChildren.GetChangeEvent(nil, changeItemConfig, append(parentChanges, c)...)
				if err != nil {
					return nil, err
				}
				changeItems = append(changeItems, childEvents...)
			}
		}

		// 2.2 compare delete
		for key, oldChild := range oldChange.children {
			if _, ok := c.children[key]; !ok { // delete
				if changeItemConfig.Delete != nil {
					for childKey, oldChildValues := range oldChild.data {
						if item := changeItemConfig.Delete(childKey, nil, oldChildValues, append(parentChanges, oldChild)...); item != nil {
							changeItems = append(changeItems, item)
						}
					}
				}
			}
		}
	} else {
		// Handle case where oldChange is nil but there are new children
		for _, newChildren := range c.children {
			childEvents, err := newChildren.GetChangeEvent(nil, changeItemConfig, append(parentChanges, c)...)
			if err != nil {
				return nil, err
			}
			changeItems = append(changeItems, childEvents...)
		}
	}

	return changeItems, nil
}
