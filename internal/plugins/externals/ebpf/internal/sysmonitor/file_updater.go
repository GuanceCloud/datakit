//go:build linux
// +build linux

package sysmonitor

import (
	"os"
	"sync"
	"time"
)

type probeInfo struct {
	modTime time.Time
	inj     *ProcInjectC
	refCnt  int
}

type PassiveFileUpdater struct {
	fileRecords   map[uint64]*probeInfo
	binPathHashes map[uint64]string
	mu            sync.RWMutex
}

func NewPassiveFileUpdater() *PassiveFileUpdater {
	return &PassiveFileUpdater{
		fileRecords:   make(map[uint64]*probeInfo),
		binPathHashes: make(map[uint64]string),
	}
}

func (p *PassiveFileUpdater) Check(binpath string, absPath uint64) (*probeInfo, bool, error) {
	fileInfo, err := os.Stat(binpath)
	if err != nil {
		return nil, false, err
	}
	currentModTime := fileInfo.ModTime()

	p.mu.RLock()
	lastRec, exists := p.fileRecords[absPath]
	p.mu.RUnlock()

	if !exists {
		p.mu.Lock()
		v := &probeInfo{
			modTime: currentModTime,
			refCnt:  0,
		}
		p.fileRecords[absPath] = v
		p.binPathHashes[absPath] = binpath
		p.mu.Unlock()
		return v, true, nil
	}

	if !currentModTime.Equal(lastRec.modTime) {
		p.mu.Lock()
		v := &probeInfo{
			modTime: currentModTime,
			inj:     lastRec.inj,
			refCnt:  lastRec.refCnt,
		}
		p.fileRecords[absPath] = v
		p.binPathHashes[absPath] = binpath
		p.mu.Unlock()
		return v, true, nil
	}

	return lastRec, false, nil
}

func (p *PassiveFileUpdater) Inject(absPath uint64, injO *ProcInjectC) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if rec, ok := p.fileRecords[absPath]; ok {
		rec.inj = injO
	}
}

func (p *PassiveFileUpdater) AddRef(absPath uint64) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if rec, ok := p.fileRecords[absPath]; ok {
		rec.refCnt++
	}
}

func (p *PassiveFileUpdater) Forget(absPath uint64) (string, bool) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if rec, ok := p.fileRecords[absPath]; ok {
		rec.refCnt--
		if rec.refCnt <= 0 {
			binPath := p.binPathHashes[absPath]
			delete(p.fileRecords, absPath)
			delete(p.binPathHashes, absPath)
			return binPath, true
		}
		return p.binPathHashes[absPath], false
	}
	return "", false
}

func (p *PassiveFileUpdater) ForgetAll() {
	p.mu.Lock()
	p.fileRecords = make(map[uint64]*probeInfo)
	p.binPathHashes = make(map[uint64]string)
	p.mu.Unlock()
}
