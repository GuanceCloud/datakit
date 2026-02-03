// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

//go:build !windows
// +build !windows

package hostchange

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestChangeAdd(t *testing.T) {
	oldChange := NewChange[string, string]()
	oldChange.Add("127.0.0.1", "localhost")
	oldChange.Add("127.0.0.1", "dev.dca.com")
	oldChange.Add("127.0.0.2", "delete.com")

	newChange := NewChange[string, string]()
	newChange.Add("127.0.0.1", "localhost")
	newChange.Add("127.0.0.1", "dev.dca.com")
	newChange.Add("127.0.0.1", "prod.dca.com")
	newChange.Add("127.0.0.3", "new.com")

	changeEvents, err := newChange.GetChangeEvent(oldChange, &ChangeItemConfig[string, string]{
		Add: func(key string, newValues []string, oldValues []string, parentChanges ...*Change[string, string]) *ChangeItem {
			assert.Equal(t, 1, len(newValues))
			assert.Equal(t, "127.0.0.3", key)
			assert.ElementsMatch(t, newValues, []string{"new.com"})
			return &ChangeItem{
				ChangeID:             "add",
				ChangeTimestampMicro: 123456,
				Title:                "add title",
				Message:              "add message",
			}
		},
		Delete: func(key string, newValues []string, oldValues []string, parentChanges ...*Change[string, string]) *ChangeItem {
			assert.Equal(t, 1, len(oldValues))
			assert.Equal(t, "127.0.0.2", key)
			assert.ElementsMatch(t, oldValues, []string{"delete.com"})
			return &ChangeItem{
				ChangeID:             "delete",
				ChangeTimestampMicro: 123456,
				Title:                "delete title",
				Message:              "delete message",
			}
		},
		Modify: func(key string, newValues []string, oldValues []string, parentChanges ...*Change[string, string]) *ChangeItem {
			assert.Equal(t, "127.0.0.1", key)
			assert.ElementsMatch(t, newValues, []string{"localhost", "dev.dca.com", "prod.dca.com"})
			assert.ElementsMatch(t, oldValues, []string{"localhost", "dev.dca.com"})
			return &ChangeItem{
				ChangeID:             "modify",
				ChangeTimestampMicro: 123456,
				Title:                "modify title",
				Message:              "modify message",
			}
		},
	})

	assert.NoError(t, err)
	assert.Equal(t, 3, len(changeEvents))
}

type ChildItem struct {
	ParentKey string
	Data      string
}

func TestChangeAddChild(t *testing.T) {
	parentKey := "/etc/crontab"
	parentDeleteKey := "/etc/crontab_delete"

	oldChange := NewChange[string, ChildItem]()
	oldChildChange := NewChange[string, ChildItem]()
	oldChildChange.Add("1", ChildItem{ParentKey: parentKey, Data: "localhost"})
	oldChildChange.Add("2", ChildItem{ParentKey: parentKey, Data: "localhost"})
	oldChange.SetChild(parentKey, oldChildChange)

	oldDeleteChange := NewChange[string, ChildItem]()
	oldChildDeleteChange := NewChange[string, ChildItem]()
	oldChildDeleteChange.Add("1", ChildItem{ParentKey: parentDeleteKey, Data: "localhost"})
	oldChildDeleteChange.Add("2", ChildItem{ParentKey: parentDeleteKey, Data: "localhost"})
	oldDeleteChange.SetChild(parentDeleteKey, oldChildDeleteChange)

	newChange := NewChange[string, ChildItem]()
	newChildChange := NewChange[string, ChildItem]()
	newChildChange.Add("2", ChildItem{ParentKey: parentKey, Data: "edit"})
	newChildChange.Add("3", ChildItem{ParentKey: parentKey, Data: "add"})
	newChange.SetChild(parentKey, newChildChange)

	changeEvents, err := newChange.GetChangeEvent(oldChange, &ChangeItemConfig[string, ChildItem]{
		Add: func(key string, newValues []ChildItem, oldValues []ChildItem, parentChanges ...*Change[string, ChildItem]) *ChangeItem {
			assert.Equal(t, "3", key)
			return &ChangeItem{
				ChangeID:             "add",
				ChangeTimestampMicro: 123456,
				Title:                "add title",
				Message:              "add message",
			}
		},
		Delete: func(key string, newValues []ChildItem, oldValues []ChildItem, parentChanges ...*Change[string, ChildItem]) *ChangeItem {
			assert.Equal(t, "1", key)
			return &ChangeItem{
				ChangeID:             "delete",
				ChangeTimestampMicro: 123456,
				Title:                "delete title",
				Message:              "delete message",
			}
		},
		Modify: func(key string, newValues []ChildItem, oldValues []ChildItem, parentChanges ...*Change[string, ChildItem]) *ChangeItem {
			assert.Equal(t, "2", key)
			return &ChangeItem{
				ChangeID:             "modify",
				ChangeTimestampMicro: 123456,
				Title:                "modify title",
				Message:              "modify message",
			}
		},
	})

	assert.NoError(t, err)
	assert.Equal(t, 3, len(changeEvents))
}
