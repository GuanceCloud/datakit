// Package conntrack wrap host network collection stats
package conntrack

type Info struct {
	Current       int64 `json:"entries"`
	Limit         int64 `json:"entries_limit"`
	Found         int64 `json:"stat_found"`
	Invalid       int64 `json:"stat_invalid"`
	Ignore        int64 `json:"stat_ignore"`
	Insert        int64 `json:"stat_insert"`
	InsertFailed  int64 `json:"stat_insert_failed"`
	Drop          int64 `json:"stat_drop"`
	EarlyDrop     int64 `json:"stat_early_drop"`
	SearchRestart int64 `json:"stat_search_restart"`
}

func GetConntrackInfo() *Info {
	info := Collect()
	return info
}
