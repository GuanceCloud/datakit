//go:build linux
// +build linux

package protodec

import "github.com/GuanceCloud/cliutils/point"

func appendKV(kvs point.KVs, key string, val any) point.KVs {
	return append(kvs, point.NewKV(key, val))
}
