// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package configwatcher

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/GuanceCloud/cliutils/point"
	"github.com/google/uuid"
)

func generateTitle(source string, event *changeEvent) string {
	return fmt.Sprintf("%s %s in %s", source, event.typ.String(), filepath.Base(event.path))
}

func generateMessage(source string, event *changeEvent) string {
	diff := "(No content changes to display)"
	if event.diff != "" {
		diff = event.diff
	}

	// nolint:exhaustive
	switch event.typ {
	case created:
		return fmt.Sprintf("%s at %s was created\nchange details:\n%s", source, event.path, diff)
	case modified:
		return fmt.Sprintf("%s at %s was modified\nchange details:\n%s", source, event.path, diff)
	case deleted:
		return fmt.Sprintf("%s at %s was deleted", source, event.path)
	default:
		return "Unexpected Type"
	}
}

func buildDefaultChangeEventKVs() (kvs point.KVs) {
	const (
		defaultStatus = "info"
		defaultSource = "change"
	)

	var uid string
	if u, err := uuid.NewRandom(); err == nil {
		uid = "event-" + strings.ToLower(u.String())
	}

	kvs = kvs.AddTag("df_event_id", uid)
	kvs = kvs.AddTag("df_source", defaultSource)
	kvs = kvs.AddTag("df_status", defaultStatus)
	kvs = kvs.AddTag("df_sub_status", defaultStatus)
	return
}

func isHiddenFile(path string) bool {
	filename := filepath.Base(path)
	return strings.HasPrefix(filename, ".")
}
