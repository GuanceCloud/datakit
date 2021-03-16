package tailf

import (
	"sync"

	"github.com/fsnotify/fsnotify"
)

type notifyType int

const (
	renameNotify notifyType = iota + 1
)

type Watcher struct {
	watcher *fsnotify.Watcher
	list    sync.Map
}

func NewWatcher() (*Watcher, error) {
	var err error
	var f = &Watcher{}

	f.watcher, err = fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	return f, nil
}

func (w *Watcher) Close() error {
	return w.watcher.Close()
}

func (w *Watcher) Add(file string, notifyCh chan notifyType) error {
	// FIXME:
	// if w.watcher == nil {
	// 	return fmt.Errorf("invalid Watcher instance, should use NewWatcher()")
	// }

	if ok := w.IsExist(file); ok {
		return nil
	}

	err := w.watcher.Add(file)
	if err != nil {
		return err
	}

	w.list.Store(file, notifyCh)
	return nil
}

func (w *Watcher) Remove(file string) error {
	err := w.watcher.Remove(file)
	if err != nil {
		return err
	}

	w.list.Delete(file)
	return nil
}

func (w *Watcher) IsExist(file string) bool {
	_, ok := w.list.Load(file)
	return ok
}

func (w *Watcher) Watching(done <-chan interface{}) {
	for {
		select {
		case <-done:
			return

		case event, ok := <-w.watcher.Events:
			if !ok {
				continue
			}

			if event.Op&fsnotify.Rename == fsnotify.Rename {
				notifyCh, ok := w.list.Load(event.Name)
				if !ok {
					continue
				}
				notifyCh.(chan notifyType) <- renameNotify
			}

		case _, ok := <-w.watcher.Errors:
			if !ok {
				continue
			}
		}
	}
}
