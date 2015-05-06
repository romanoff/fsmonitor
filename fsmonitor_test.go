package fsmonitor

import (
	"io/ioutil"
	"os"
	"path/filepath"

	"gopkg.in/fsnotify.v1"

	. "gopkg.in/check.v1"
)

const (
	// default permissions
	defaultFilePerms = 0600
	defaultDirPerms  = 0700
)

type WatcherTests struct {
	dir     string
	watcher *Watcher
}

var _ = Suite(&WatcherTests{})

func (t *WatcherTests) SetUpTest(c *C) {
	t.dir = c.MkDir()
	watcher, err := NewWatcher()
	c.Assert(err, IsNil)
	t.watcher = watcher
}

func (t *WatcherTests) TearDownTest(c *C) {
	err := t.watcher.Close()
	c.Assert(err, IsNil)
}

func (t *WatcherTests) TestFileCreation(c *C) {
	err := t.watcher.Watch(t.dir)
	c.Assert(err, IsNil)

	name := filepath.Join(t.dir, "test.txt")
	err = ioutil.WriteFile(name, []byte("asdf"), defaultFilePerms)
	c.Assert(err, IsNil)

	event := <-t.watcher.Events
	c.Assert(event.Name, Equals, name)
	c.Assert(event.Op, Equals, fsnotify.Create)
}

func (t *WatcherTests) TestFileDeletion(c *C) {
	name := filepath.Join(t.dir, "test.txt")
	err := ioutil.WriteFile(name, []byte("asdf"), defaultFilePerms)
	c.Assert(err, IsNil)

	err = t.watcher.Watch(t.dir)
	c.Assert(err, IsNil)

	err = os.Remove(name)
	c.Assert(err, IsNil)

	event := <-t.watcher.Events
	c.Assert(event.Name, Equals, name)
	c.Assert(event.Op, Equals, fsnotify.Remove)
}

func (t *WatcherTests) TestFileModification(c *C) {
	name := filepath.Join(t.dir, "test.txt")
	err := ioutil.WriteFile(name, []byte("asdf"), defaultFilePerms)
	c.Assert(err, IsNil)

	err = t.watcher.Watch(t.dir)
	c.Assert(err, IsNil)

	err = ioutil.WriteFile(name, []byte("asdfefadsjklvafh7öafdsü+38"), defaultFilePerms)
	c.Assert(err, IsNil)

	event := <-t.watcher.Events
	c.Assert(event.Name, Equals, name)
	c.Assert(event.Op, Equals, fsnotify.Write)
}

func (t *WatcherTests) TestFileRename(c *C) {
	name := filepath.Join(t.dir, "test.txt")
	err := ioutil.WriteFile(name, []byte("asdf"), defaultFilePerms)
	c.Assert(err, IsNil)

	err = t.watcher.Watch(t.dir)
	c.Assert(err, IsNil)

	newName := filepath.Join(t.dir, "test2.txt")
	err = os.Rename(name, newName)
	c.Assert(err, IsNil)

	event := <-t.watcher.Events
	c.Assert(event.Name, Equals, name)
	c.Assert(event.Op, Equals, fsnotify.Rename)
}

func (t *WatcherTests) TestCreateEventInDir(c *C) {
	dirName := filepath.Join(t.dir, "test")
	err := os.Mkdir(dirName, defaultDirPerms)
	c.Assert(err, IsNil)

	err = t.watcher.Watch(t.dir)
	c.Assert(err, IsNil)

	name := filepath.Join(dirName, "test.txt")
	err = ioutil.WriteFile(name, []byte("asdf"), defaultFilePerms)
	c.Assert(err, IsNil)

	event := <-t.watcher.Events
	c.Assert(event.Name, Equals, name)
	c.Assert(event.Op, Equals, fsnotify.Create)
}

func (t *WatcherTests) TestCreateEventInCreatedDir(c *C) {
	err := t.watcher.Watch(t.dir)
	c.Assert(err, IsNil)

	dirName := filepath.Join(t.dir, "test")
	err = os.Mkdir(dirName, defaultDirPerms)
	c.Assert(err, IsNil)

	event := <-t.watcher.Events
	c.Assert(event.Name, Equals, dirName)
	c.Assert(event.Op, Equals, fsnotify.Create)

	name := filepath.Join(dirName, "test.txt")
	err = ioutil.WriteFile(name, []byte("asdf"), defaultFilePerms)
	c.Assert(err, IsNil)

	event = <-t.watcher.Events
	c.Assert(event.Name, Equals, name)
	c.Assert(event.Op, Equals, fsnotify.Create)
}

func (t *WatcherTests) TestDeleteInDir(c *C) {
	dirName := filepath.Join(t.dir, "test")
	err := os.Mkdir(dirName, defaultDirPerms)
	c.Assert(err, IsNil)

	name := filepath.Join(dirName, "test.txt")
	err = ioutil.WriteFile(name, []byte("asdf"), defaultFilePerms)
	c.Assert(err, IsNil)

	err = t.watcher.Watch(t.dir)
	c.Assert(err, IsNil)

	err = os.Remove(name)
	c.Assert(err, IsNil)

	event := <-t.watcher.Events
	c.Assert(event.Name, Equals, name)
	c.Assert(event.Op, Equals, fsnotify.Remove)
}

func (t *WatcherTests) TestDeleteInCreatedDir(c *C) {
	err := t.watcher.Watch(t.dir)
	c.Assert(err, IsNil)

	dirName := filepath.Join(t.dir, "test")
	err = os.Mkdir(dirName, defaultDirPerms)
	c.Assert(err, IsNil)

	event := <-t.watcher.Events
	c.Assert(event.Name, Equals, dirName)
	c.Assert(event.Op, Equals, fsnotify.Create)

	name := filepath.Join(dirName, "test.txt")
	err = ioutil.WriteFile(name, []byte("asdf"), defaultFilePerms)
	c.Assert(err, IsNil)

	event = <-t.watcher.Events
	c.Assert(event.Name, Equals, name)
	c.Assert(event.Op, Equals, fsnotify.Create)
	event = <-t.watcher.Events
	c.Assert(event.Name, Equals, name)
	c.Assert(event.Op, Equals, fsnotify.Write)

	err = os.Remove(name)
	c.Assert(err, IsNil)

	event = <-t.watcher.Events
	c.Assert(event.Name, Equals, name)
	c.Assert(event.Op, Equals, fsnotify.Remove)
}

func (t *WatcherTests) TestClose(c *C) {
	err := t.watcher.Watch(t.dir)
	c.Assert(err, IsNil)

    c.Assert(t.watcher.IsClosed(), Equals, false)
	err = t.watcher.Close()
	c.Assert(err, IsNil)
    c.Assert(t.watcher.IsClosed(), Equals, true)

	watcher, err := NewWatcher()
	c.Assert(err, IsNil)
	t.watcher = watcher
}
