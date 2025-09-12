//go:build linux
// +build linux

package sysmonitor

import (
	"os"
	"sync"
	"time"
)

type probeInfo struct {
	absPath string
	modTime time.Time
	inj     *ProcInjectC
}
type PassiveFileUpdater struct {
	fileRecords map[string]*probeInfo
	mu          sync.RWMutex
}

func NewPassiveFileUpdater() *PassiveFileUpdater {
	return &PassiveFileUpdater{
		fileRecords: make(map[string]*probeInfo),
	}
}

func (p *PassiveFileUpdater) Check(absPath string) (*probeInfo, bool, error) {
	fileInfo, err := os.Stat(absPath)
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
			absPath: absPath,
			modTime: currentModTime,
		}
		p.fileRecords[absPath] = v
		p.mu.Unlock()
		return v, true, nil
	}

	if !currentModTime.Equal(lastRec.modTime) {
		// 更新记录的修改时间
		p.mu.Lock()
		v := &probeInfo{
			absPath: absPath,
			modTime: currentModTime,
			inj:     lastRec.inj,
		}
		p.fileRecords[absPath] = v

		p.mu.Unlock()
		return v, true, nil
	}

	// 未更新
	return lastRec, false, nil
}

func (p *PassiveFileUpdater) Inject(absPath string, injO *ProcInjectC) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if rec, ok := p.fileRecords[absPath]; ok {
		rec.inj = injO
	}
}

func (p *PassiveFileUpdater) Forget(absPath string) error {
	p.mu.Lock()
	delete(p.fileRecords, absPath)
	p.mu.Unlock()
	return nil
}

func (p *PassiveFileUpdater) ForgetAll() {
	p.mu.Lock()
	p.fileRecords = make(map[string]*probeInfo)
	p.mu.Unlock()
}
