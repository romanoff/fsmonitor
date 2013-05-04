package fsmonitor

import (
	"github.com/howeyc/fsnotify"
	"path/filepath"
	"os"
)

func NewWatcher() (*Watcher, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}
	event := make(chan *fsnotify.FileEvent)
	watcherError := make(chan error)
	monitorWatcher := &Watcher{Event: event, Error: watcherError, watcher: watcher}
	go func() {
		for {
			select {
			case ev := <- watcher.Event:
				event <- ev
				if ev.IsCreate() {
					go func() {
						monitorWatcher.watchAllFolders(ev.Name)
					}()
				}
				if ev.IsDelete() {
					go func() {
						watcher.RemoveWatch(ev.Name)
					}()
				}
			case e := <- watcher.Error:
				watcherError  <- e
			}
		}
	}()
	return monitorWatcher, nil
}

type Watcher struct {
	Event chan *fsnotify.FileEvent
	Error chan error
	watcher *fsnotify.Watcher
}

func (self *Watcher) Watch(path string) error {
	err := self.watchAllFolders(path)
	if err != nil {
		return err
	}	
	return nil
}

func (self *Watcher) watchAllFolders(path string) (err error) {
	err = filepath.Walk(path, func(path string, f os.FileInfo, err error) error {
		if f.IsDir() {
			err := self.addWatcher(path)
			if err != nil {
				return err
			}		
		}
		return nil
	})
	return
}

func (self *Watcher) addWatcher(path string) (err error) {
	err = self.watcher.Watch(path)
	return
}
