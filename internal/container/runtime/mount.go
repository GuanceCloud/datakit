// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package runtime

import (
	"path/filepath"
	"strings"
)

// 说明：
// 本文件提供与容器挂载（Mount）相关的路径解析能力，用于在容器与宿主机之间定位文件：
// 1) FindBestMount(path): 在给定 mounts 中，寻找与 path 形成“最长路径段前缀匹配”的挂载，返回匹配到的 Destination 与 Source；
// 2) ResolveToSourcePath(destination, source, path): 将 path 中与 destination 的相同前缀去除，把余下部分追加到 source 后，得到宿主机侧路径；
// 3) ReverseResolveFromSourcePath(mapped, destination, source): 将宿主机侧 mapped 路径逆向还原为以 destination 为前缀的容器内路径。
// 三者配合实现：容器内路径 -> 宿主机路径 的正向解析，以及宿主机路径 -> 容器内路径 的逆向还原。

type Mount struct {
	Type        string
	Destination string
	Source      string
}

type Mounts []Mount

// FindBestMount 在 mounts 中查找能与给定 path 形成最长路径段前缀匹配的挂载，
// 返回该挂载的 Destination 与 Source。若无匹配，返回 ok=false。
func (mounts Mounts) FindBestMount(path string) (destination, source string, ok bool) {
	if len(mounts) == 0 || path == "" {
		return "", "", false
	}

	cleanPath := filepath.Clean(path)
	var (
		bestLen  int
		bestDest string
		bestSrc  string
	)

	for _, m := range mounts {
		if m.Destination == "" {
			continue
		}
		dest := filepath.Clean(m.Destination)
		if dest == cleanPath || strings.HasPrefix(cleanPath, dest+string(filepath.Separator)) {
			if l := len(dest); l > bestLen {
				bestLen = l
				bestDest = dest
				bestSrc = filepath.Clean(m.Source)
			}
		}
	}

	if bestLen == 0 {
		return "", "", false
	}
	return bestDest, bestSrc, true
}

// ResolveToSourcePath 将给定 path 中与 destination 的最长前缀去除，
// 将剩余部分追加到 source 后，得到基于 source 的路径。
// 若 path 不是以路径段意义上的 destination 为前缀（或等于），返回 ok=false。
func ResolveToSourcePath(destination, source, path string) (mapped string, ok bool) {
	if destination == "" || source == "" || path == "" {
		return "", false
	}
	dest := filepath.Clean(destination)
	src := filepath.Clean(source)
	cleanPath := filepath.Clean(path)

	if dest == cleanPath {
		return src, true
	}
	if !strings.HasPrefix(cleanPath, dest+string(filepath.Separator)) {
		return "", false
	}
	// 计算与 destination 匹配后的相对后缀。
	rel := strings.TrimPrefix(cleanPath, dest)
	rel = strings.TrimPrefix(rel, string(filepath.Separator))
	if rel == "" {
		return src, true
	}
	return filepath.Join(src, rel), true
}

// ReverseResolveFromSourcePath 将位于 Source 下的路径还原为位于对应 Destination
// 下的原始路径，是 ResolveToSourcePath 的逆操作。
// 示例：destination=/tmp/opt/abc，source=/mount/a，mapped=/mount/a/123.go，
// 返回 /tmp/opt/abc/123.go。
// 如果 mapped 不是以路径段意义上的 source 为前缀，则返回 ok=false。
func ReverseResolveFromSourcePath(mapped, destination, source string) (original string, ok bool) {
	if mapped == "" || source == "" {
		return "", false
	}

	cleanMapped := filepath.Clean(mapped)
	destWasEmpty := destination == ""
	dest := destination
	if !destWasEmpty {
		dest = filepath.Clean(destination)
	}
	src := filepath.Clean(source)

	// helper: join with dest allowing empty destination to mean "/" prefix
	joinWithDest := func(rel string) (string, bool) {
		if rel == "" {
			if destWasEmpty {
				return string(filepath.Separator), true
			}
			return dest, true
		}
		if destWasEmpty {
			return string(filepath.Separator) + rel, true
		}
		return filepath.Join(dest, rel), true
	}

	// Source 为根目录的特殊处理。
	if src == string(filepath.Separator) {
		rel := strings.TrimPrefix(cleanMapped, src)
		rel = strings.TrimPrefix(rel, string(filepath.Separator))
		return joinWithDest(rel)
	}

	if src == cleanMapped {
		return joinWithDest("")
	}
	if strings.HasPrefix(cleanMapped, src+string(filepath.Separator)) {
		rel := strings.TrimPrefix(cleanMapped, src)
		rel = strings.TrimPrefix(rel, string(filepath.Separator))
		return joinWithDest(rel)
	}
	return "", false
}
