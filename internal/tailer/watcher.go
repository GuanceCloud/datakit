package tailer

import (
	"context"
	"os"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
)

const (
	checkFileExistsInterval = time.Second * 5
)

type WatcherStop interface {
	Stop()
}

type Watcher struct {
	watcher *fsnotify.Watcher
	list    map[string]WatcherStop
	mu      sync.Mutex
}

func NewWatcher() (*Watcher, error) {
	var err error
	var w = &Watcher{}

	w.watcher, err = fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}
	w.list = make(map[string]WatcherStop)

	return w, nil
}

func (w *Watcher) Close() error {
	w.mu.Lock()
	defer w.mu.Unlock()

	for filename, in := range w.list {
		in.Stop()
		delete(w.list, filename)
	}

	return w.watcher.Close()
}

func (w *Watcher) List() []string {
	w.mu.Lock()
	defer w.mu.Unlock()

	var res []string
	for filename := range w.list {
		res = append(res, filename)
	}

	return res
}

func (w *Watcher) Add(filename string, in WatcherStop) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	if _, ok := w.list[filename]; ok {
		return nil
	}

	if err := w.watcher.Add(filename); err != nil {
		return err
	}

	w.list[filename] = in
	return nil
}

func (w *Watcher) Stop(filename string) {
	w.mu.Lock()
	defer w.mu.Unlock()
	if in, ok := w.list[filename]; ok {
		in.Stop()
	}
}

func (w *Watcher) Remove(filename string) error {
	w.mu.Lock()
	defer w.mu.Unlock()
	delete(w.list, filename)
	return w.watcher.Remove(filename)
}

func (w *Watcher) IsExist(filename string) bool {
	w.mu.Lock()
	defer w.mu.Unlock()
	_, ok := w.list[filename]
	return ok
}

func (w *Watcher) CleanExpired() {
	w.mu.Lock()
	defer w.mu.Unlock()

	for filename, in := range w.list {
		_, statErr := os.Lstat(filename)
		if os.IsNotExist(statErr) {
			in.Stop()
			delete(w.list, filename)
			w.watcher.Remove(filename)
		}
	}
}

func (w *Watcher) Watching(ctx context.Context) {
	tick := time.NewTicker(checkFileExistsInterval)
	defer tick.Stop()

	for {
		select {
		case <-ctx.Done():
			return

		case event, ok := <-w.watcher.Events:
			if !ok {
				continue
			}
			if event.Op&fsnotify.Rename == fsnotify.Rename {
				w.Stop(event.Name)
				w.Remove(event.Name)
			}

		case <-tick.C:
			// 为什么不使用 notify 的方式监控文件删除，而采用 Lstat()
			// notify 只有当文件引用计数为 0 时，才会认为此文件已经被删除，从而触发 remove event
			// 在此处，datakit 打开文件后保存句柄，即使 rm 删除文件，该文件的引用计数依旧是 1，因为 datakit 在占用
			// 从而导致，datakit 占用文件无法删除，无法删除就收不到 remove event，此 goroutine 就会长久存在
			// 且极端条件下，长时间运行，可能会导致磁盘容量不够的情况，因为占用容量的文件在此被引用，新数据无法覆盖
			// 以上结论仅限于 linux
			w.CleanExpired()

		case _, ok := <-w.watcher.Errors:
			if !ok {
				continue
			}
		}
	}
}
