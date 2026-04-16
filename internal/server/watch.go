package server

import (
	"log"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
)

type debouncer struct {
	mu       sync.Mutex
	timer    *time.Timer
	duration time.Duration
	fn       func()
}

func newDebouncer(d time.Duration, fn func()) *debouncer {
	return &debouncer{duration: d, fn: fn}
}

func (d *debouncer) trigger() {
	d.mu.Lock()
	defer d.mu.Unlock()
	if d.timer != nil {
		d.timer.Stop()
	}
	d.timer = time.AfterFunc(d.duration, d.fn)
}

func watchFile(path string, onChange func()) (func(), error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}
	if err := watcher.Add(path); err != nil {
		watcher.Close()
		return nil, err
	}

	db := newDebouncer(100*time.Millisecond, onChange)

	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				if event.Has(fsnotify.Write) || event.Has(fsnotify.Create) {
					db.trigger()
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				log.Printf("[watch] error: %v", err)
			}
		}
	}()

	return func() { watcher.Close() }, nil
}
