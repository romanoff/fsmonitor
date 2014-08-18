package fsmonitor

import (
	"os"
	"path/filepath"

	"gopkg.in/fsnotify.v0"
)

func NewWatcher() (*Watcher, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}
	monitorWatcher := initWatcher(watcher, []string{})
	return monitorWatcher, nil
}

func NewWatcherWithSkipFolders(skipFolders []string) (*Watcher, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}
	monitorWatcher := initWatcher(watcher, skipFolders)
	return monitorWatcher, nil
}

func initWatcher(watcher *fsnotify.Watcher, skipFolders []string) *Watcher {
	event := make(chan *fsnotify.FileEvent)
	watcherError := make(chan error)
	monitorWatcher := &Watcher{Event: event, Error: watcherError, watcher: watcher, SkipFolders: skipFolders}
	go func() {
		for {
			select {
			case ev := <-watcher.Event:
				event <- ev
				if ev.IsCreate() {
					go func() {
						if f, err := os.Stat(ev.Name); err == nil {
							if f.IsDir() {
								monitorWatcher.watchAllFolders(ev.Name)
							}
						}

					}()
				}
				if ev.IsDelete() {
					go func() {
						watcher.RemoveWatch(ev.Name)
					}()
				}
			case e := <-watcher.Error:
				watcherError <- e
			}
		}
	}()
	return monitorWatcher
}

type Watcher struct {
	Event       chan *fsnotify.FileEvent
	Error       chan error
	SkipFolders []string
	watcher     *fsnotify.Watcher
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
		if f != nil && f.IsDir() {
			filename := f.Name()
			for _, skipFolder := range self.SkipFolders {
				match, err := filepath.Match(skipFolder, filename)
				if err != nil {
					return err
				}
				if match {
					return filepath.SkipDir
				}
			}
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
