package fsmonitor

import (
	"os"
	"path/filepath"

    "gopkg.in/fsnotify.v1"
)

// NewWatcher initializes a new watcher which is able to recursively scan
// the directory structure.
func NewWatcher() (*Watcher, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}
	monitorWatcher := initWatcher(watcher, []string{})
	return monitorWatcher, nil
}

// NewWatcherWithSkipFolders initializes a new watcher capable of watching
// for file system events of a recursive directory structure.
func NewWatcherWithSkipFolders(skipFolders []string) (*Watcher, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}
	monitorWatcher := initWatcher(watcher, skipFolders)
	return monitorWatcher, nil
}

// initWatcher starts the underlying watcher interface.
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

// Watcher is the struct handling the watching of file system events over a recursive
// directory tree.
type Watcher struct {
	Events      chan *fsnotify.Event
	Error       chan error
	SkipFolders []string
	watcher     *fsnotify.Watcher
}

// Watch starts watching the given path and all it's subdirectories for file system
// changes.
func (self *Watcher) Watch(path string) error {
	return self.watchAllFolders(path)
}

// watchAllFolders starts watching all subdirectories via the underlying fsnotify watchers.
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
	return err
}

// addWatcher adds the given path to the underyling fsnotify watcher.
func (self *Watcher) addWatcher(path string) (err error) {
	return self.watcher.Add(path)
}
