package configmanager

import (
	"github.com/fsnotify/fsnotify"
	"github.com/project-eria/logger"
	"github.com/tidwall/gjson"
)

var (
	_watcher *fsnotify.Watcher
	_ready   chan bool
)

// Watcher struct that defines a specific JSON path to watch for changes
type Watcher struct {
	path  string
	value interface{}
	c     *ConfigManager
}

func (c *ConfigManager) initWatcher(file string) {
	if _watcher == nil {
		var err error
		_watcher, err = fsnotify.NewWatcher()
		if err != nil {
			logger.Module("configmanager").Fatal(err)
		}
		_ready = make(chan bool)
		go func() {
			for {
				select {
				case event, ok := <-_watcher.Events:
					if !ok {
						return
					}
					if event.Op&fsnotify.Write == fsnotify.Write {
						logger.Module("configmanager").WithField("file", event.Name).Debug("Config file modified, reloading config")
						c.Load()
						_ready <- true
					}
				case err, ok := <-_watcher.Errors:
					if !ok {
						return
					}
					logger.Module("configmanager").Error(err)
				}
			}
		}()
	}
	_watcher.Add(file)
}

func (c *ConfigManager) newWatcher(path string) *Watcher {
	value := gjson.GetBytes(c.json, path)
	return &Watcher{
		path:  path,
		value: value.Value(),
		c:     c,
	}
}

// Next wait for the next file change
func (c *ConfigManager) Next() {
	<-_ready
}

// Next wait for the next value change, for the specific watcher path
func (w *Watcher) Next() interface{} {
	for {
		<-_ready
		newValue := gjson.GetBytes(w.c.json, w.path)
		if w.value == newValue.Value() {
			continue
		}
		w.value = newValue.Value()
		return w.value
	}
}

func closeWatcher() {
	_watcher.Close()
}
