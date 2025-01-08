// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package nfs

import (
	"bufio"
	"fmt"
	"io"
	"strings"

	"github.com/prometheus/procfs/nfs"
)

// parseServerRPCStats returns stats read from /proc/net/rpc/nfsd.
func parseServerRPCStats(r io.Reader) (*nfs.ServerRPCStats, error) {
	stats := &nfs.ServerRPCStats{}

	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.Fields(scanner.Text())
		// require at least <key> <value>
		if len(parts) < 2 {
			return nil, fmt.Errorf("invalid NFSd metric line %q", line)
		}
		label := parts[0]

		var values []uint64
		var err error
		if label == "th" {
			if len(parts) < 3 {
				return nil, fmt.Errorf("invalid NFSd th metric line %q", line)
			}
			values, err = parseUint64s(parts[1:3])
		} else {
			values, err = parseUint64s(parts[1:])
		}
		if err != nil {
			return nil, fmt.Errorf("error parsing NFSd metric line: %w", err)
		}

		switch metricLine := parts[0]; metricLine {
		case "rc":
			stats.ReplyCache, err = parseReplyCache(values)
		case "fh":
			stats.FileHandles, err = parseFileHandles(values)
		case "io":
			stats.InputOutput, err = parseInputOutput(values)
		case "th":
			stats.Threads, err = parseThreads(values)
		case "ra":
			stats.ReadAheadCache, err = parseReadAheadCache(values)
		case "net":
			stats.Network, err = parseNetwork(values)
		case "rpc":
			stats.ServerRPC, err = parseServerRPC(values)
		case "proc2":
			stats.V2Stats, err = parseV2Stats(values)
		case "proc3":
			stats.V3Stats, err = parseV3Stats(values)
		case "proc4":
			stats.ServerV4Stats, err = parseServerV4Stats(values)
		case "proc4ops":
			stats.V4Ops, err = parseV4Ops(values)
		default:
			// 如果是未知的度量行，不返回错误，而是继续处理下一行
			continue
		}
		if err != nil {
			return nil, fmt.Errorf("errors parsing NFSd metric line: %w", err)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error scanning NFSd file: %w", err)
	}

	return stats, nil
}

func parseReplyCache(v []uint64) (nfs.ReplyCache, error) {
	if len(v) != 3 {
		return nfs.ReplyCache{}, fmt.Errorf("invalid ReplyCache line %q", v)
	}

	return nfs.ReplyCache{
		Hits:    v[0],
		Misses:  v[1],
		NoCache: v[2],
	}, nil
}

func parseFileHandles(v []uint64) (nfs.FileHandles, error) {
	if len(v) != 5 {
		return nfs.FileHandles{}, fmt.Errorf("invalid FileHandles, line %q", v)
	}

	return nfs.FileHandles{
		Stale:        v[0],
		TotalLookups: v[1],
		AnonLookups:  v[2],
		DirNoCache:   v[3],
		NoDirNoCache: v[4],
	}, nil
}

func parseInputOutput(v []uint64) (nfs.InputOutput, error) {
	if len(v) != 2 {
		return nfs.InputOutput{}, fmt.Errorf("invalid InputOutput line %q", v)
	}

	return nfs.InputOutput{
		Read:  v[0],
		Write: v[1],
	}, nil
}

func parseThreads(v []uint64) (nfs.Threads, error) {
	if len(v) != 2 {
		return nfs.Threads{}, fmt.Errorf("invalid Threads line %q", v)
	}

	return nfs.Threads{
		Threads: v[0],
		FullCnt: v[1],
	}, nil
}

func parseReadAheadCache(v []uint64) (nfs.ReadAheadCache, error) {
	if len(v) != 12 {
		return nfs.ReadAheadCache{}, fmt.Errorf("invalid ReadAheadCache line %q", v)
	}

	return nfs.ReadAheadCache{
		CacheSize:      v[0],
		CacheHistogram: v[1:11],
		NotFound:       v[11],
	}, nil
}

func parseNetwork(v []uint64) (nfs.Network, error) {
	if len(v) != 4 {
		return nfs.Network{}, fmt.Errorf("invalid Network line %q", v)
	}

	return nfs.Network{
		NetCount:   v[0],
		UDPCount:   v[1],
		TCPCount:   v[2],
		TCPConnect: v[3],
	}, nil
}

func parseServerRPC(v []uint64) (nfs.ServerRPC, error) {
	if len(v) != 5 {
		return nfs.ServerRPC{}, fmt.Errorf("invalid RPC line %q", v)
	}

	return nfs.ServerRPC{
		RPCCount: v[0],
		BadCnt:   v[1],
		BadFmt:   v[2],
		BadAuth:  v[3],
		BadcInt:  v[4],
	}, nil
}

func parseV2Stats(v []uint64) (nfs.V2Stats, error) {
	values := int(v[0])
	if len(v[1:]) != values || values < 18 {
		return nfs.V2Stats{}, fmt.Errorf("invalid V2Stats line %q", v)
	}

	return nfs.V2Stats{
		Null:     v[1],
		GetAttr:  v[2],
		SetAttr:  v[3],
		Root:     v[4],
		Lookup:   v[5],
		ReadLink: v[6],
		Read:     v[7],
		WrCache:  v[8],
		Write:    v[9],
		Create:   v[10],
		Remove:   v[11],
		Rename:   v[12],
		Link:     v[13],
		SymLink:  v[14],
		MkDir:    v[15],
		RmDir:    v[16],
		ReadDir:  v[17],
		FsStat:   v[18],
	}, nil
}

func parseV3Stats(v []uint64) (nfs.V3Stats, error) {
	values := int(v[0])
	if len(v[1:]) != values || values < 22 {
		return nfs.V3Stats{}, fmt.Errorf("invalid V3Stats line %q", v)
	}

	return nfs.V3Stats{
		Null:        v[1],
		GetAttr:     v[2],
		SetAttr:     v[3],
		Lookup:      v[4],
		Access:      v[5],
		ReadLink:    v[6],
		Read:        v[7],
		Write:       v[8],
		Create:      v[9],
		MkDir:       v[10],
		SymLink:     v[11],
		MkNod:       v[12],
		Remove:      v[13],
		RmDir:       v[14],
		Rename:      v[15],
		Link:        v[16],
		ReadDir:     v[17],
		ReadDirPlus: v[18],
		FsStat:      v[19],
		FsInfo:      v[20],
		PathConf:    v[21],
		Commit:      v[22],
	}, nil
}

func parseServerV4Stats(v []uint64) (nfs.ServerV4Stats, error) {
	values := int(v[0])
	if len(v[1:]) != values || values != 2 {
		return nfs.ServerV4Stats{}, fmt.Errorf("invalid V4Stats line %q", v)
	}

	return nfs.ServerV4Stats{
		Null:     v[1],
		Compound: v[2],
	}, nil
}

func parseV4Ops(v []uint64) (nfs.V4Ops, error) {
	values := int(v[0])
	if len(v[1:]) != values || values < 39 {
		return nfs.V4Ops{}, fmt.Errorf("invalid V4Ops line %q", v)
	}

	stats := nfs.V4Ops{
		Op0Unused:    v[1],
		Op1Unused:    v[2],
		Op2Future:    v[3],
		Access:       v[4],
		Close:        v[5],
		Commit:       v[6],
		Create:       v[7],
		DelegPurge:   v[8],
		DelegReturn:  v[9],
		GetAttr:      v[10],
		GetFH:        v[11],
		Link:         v[12],
		Lock:         v[13],
		Lockt:        v[14],
		Locku:        v[15],
		Lookup:       v[16],
		LookupRoot:   v[17],
		Nverify:      v[18],
		Open:         v[19],
		OpenAttr:     v[20],
		OpenConfirm:  v[21],
		OpenDgrd:     v[22],
		PutFH:        v[23],
		PutPubFH:     v[24],
		PutRootFH:    v[25],
		Read:         v[26],
		ReadDir:      v[27],
		ReadLink:     v[28],
		Remove:       v[29],
		Rename:       v[30],
		Renew:        v[31],
		RestoreFH:    v[32],
		SaveFH:       v[33],
		SecInfo:      v[34],
		SetAttr:      v[35],
		Verify:       v[36],
		Write:        v[37],
		RelLockOwner: v[38],
	}

	return stats, nil
}
