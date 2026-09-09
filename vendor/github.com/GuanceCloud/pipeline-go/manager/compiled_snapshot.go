// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package manager

import (
	"fmt"

	"github.com/GuanceCloud/pipeline-go/lang/platypus"
)

// NewCompiledSnapshot builds independent lookup indexes over existing compiled
// scripts. It neither compiles nor cleans up scripts, including on failure.
// The caller owns their lifetimes and must keep them alive for all readers of
// this snapshot. Do not call LoadScripts/LoadScriptWithCat on this manager:
// those legacy mutators own and clean up the scripts they replace. Defaults
// and relations may be configured before publishing the snapshot to readers.
// This constructor emits no publication events; unpublished candidates must
// not announce changes to live scripts. Existing manager APIs are unchanged.
func NewCompiledSnapshot(cfg ManagerCfg, scripts []*platypus.PlScript) (*Manager, error) {
	m := NewManager(cfg)
	for _, script := range scripts {
		if script == nil {
			return nil, fmt.Errorf("nil compiled script")
		}
		store, ok := m.whichStore(script.Category())
		if !ok {
			return nil, fmt.Errorf("unsupported script category %s", script.Category())
		}
		ns, name := script.NS(), script.Name()
		if store.storage.scripts[ns] == nil {
			store.storage.scripts[ns] = map[string]*platypus.PlScript{}
		}
		if _, exists := store.storage.scripts[ns][name]; exists {
			return nil, fmt.Errorf("duplicate compiled script %s/%s/%s", ns, script.Category(), name)
		}
		store.storage.scripts[ns][name] = script
		current := store.index[name]
		if current == nil || NSFindPriority(ns) >= NSFindPriority(current.NS()) {
			store.index[name] = script
		}
	}
	return m, nil
}
