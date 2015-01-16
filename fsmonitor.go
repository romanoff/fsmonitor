package fsmonitor

import (
	"github.com/go-fsnotify/fsnotify"
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
	event := make(chan *fsnotify.Event)
	watcherError := make(chan error)
	monitorWatcher := &Watcher{Events: event, Error: watcherError, watcher: watcher, SkipFolders: skipFolders}
	go func() {
		for {
			select {
			case ev := <-watcher.Events:
				if ev.Op&fsnotify.Create == fsnotify.Create {
					go func() {
						if f, err := os.Stat(ev.Name); err == nil {
							if f.IsDir() {
								monitorWatcher.watchAllFolders(ev.Name)
							}
						}

					}()
				}
				if ev.Op&fsnotify.Remove == fsnotify.Remove {
					go func() {
						watcher.Remove(ev.Name)
					}()
				}
				monitorWatcher.Events <- &ev
			case e := <-watcher.Errors:
				watcherError <- e
			}
		}
	}()
	return monitorWatcher
}

type Watcher struct {
	Events      chan *fsnotify.Event
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
	err = self.watcher.Add(path)
	return
}
