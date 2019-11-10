package fsmonitor

import (
	"github.com/fsnotify/fsnotify"
	"os"
	"path/filepath"
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
	event := make(chan fsnotify.Event)
	watcherError := make(chan error)
	monitorWatcher := &Watcher{
		Event:       event,
		Error:       watcherError,
		watcher:     watcher,
		SkipFolders: skipFolders,
	}
	go func() {
		for {
			select {
			case ev, ok := <-watcher.Events:
				if !ok {
					return
				}
				if ev.Op&fsnotify.Write == fsnotify.Write ||
					ev.Op&fsnotify.Create == fsnotify.Create ||
					ev.Op&fsnotify.Remove == fsnotify.Remove {
					event <- ev
				}
				if ev.Op&fsnotify.Create == fsnotify.Create {
					go func() {
						if f, err := os.Stat(ev.Name); err == nil {
							if f.IsDir() {
								monitorWatcher.watchAllFolders(ev.Name)
							}
						}

					}()
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				watcherError <- err
			}
		}
	}()
	return monitorWatcher
}

type Watcher struct {
	Event       chan fsnotify.Event
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
	return self.watcher.Add(path)
}
