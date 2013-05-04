Fsmonit
=======
Simple wrapper around [fsnotify](https://github.com/howeyc/fsnotify) to monitor changes in files and folders (including subdirectories created after monitoring has started).

Example:

```go
package main

import (
	"github.com/romanoff/fsmonitor"
	"fmt"
	"log"
	)
func main() {
	watcher, err := fsmonitor.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	err = watcher.Watch("/tmp/foo")
	if err != nil {
		log.Fatal(err)
	}
	for {
		select {
		case ev := <-watcher.Event:
			fmt.Println("event:", ev)
		case err := <-watcher.Error:
			fmt.Println("error:", err)
		}
	}
}
```
