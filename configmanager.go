package configmanager

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"sync"

	"github.com/tidwall/gjson"

	"github.com/fsnotify/fsnotify"

	"github.com/ghodss/yaml"
	"github.com/project-eria/logger"
)

// ConfigManager struct
type ConfigManager struct {
	filepath string
	json     []byte
	watcher  *fsnotify.Watcher
	s        interface{}
	sync.Mutex
}

// Init config manager with filename, and a struct
func Init(fileName string, s interface{}) (*ConfigManager, error) {
	logger.Module("configmanager").WithField("filename", fileName).Debug("Init config")
	path := os.Getenv("ERIA_CONF_PATH")
	if path == "" {
		return nil, errors.New("env ERIA_CONF_PATH not set")
	}
	filePath := filepath.Join(path, fileName)
	configManager := &ConfigManager{
		filepath: filePath,
	}
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return nil, fmt.Errorf("Config file '%s' missing", filePath)
	}

	configManager.s = s

	configManager.initWatcher(filePath)

	return configManager, nil
}

// Load config from file, based on the configmanger parameters
func (c *ConfigManager) Load() error {
	c.Lock()
	logger.Module("configmanager").Debug("Loading config")
	bytes, err := ioutil.ReadFile(c.filepath)
	if err != nil {
		// TODO What to do if file doesn't exists
		return err
	}

	if !gjson.ValidBytes(bytes) {
		return errors.New("Not a valid JSON file")
	}

	// Save as json string
	c.json = bytes

	if err := json.Unmarshal(bytes, c.s); err != nil {
		// TODO What to do if not json file
		return err
	}

	if err := processTags(c.s); err != nil {
		return err
	}
	logger.Module("configmanager").Tracef("%+v", c.s)
	c.Unlock()

	return nil
}

// Save config to file, based on the configmanger parameters
func (c *ConfigManager) Save() error {
	c.Lock()

	logger.Module("configmanager").WithField("filename", c.filepath).Debug("Saving config")

	bytes, err := json.MarshalIndent(c.s, "", "  ")
	if err != nil {
		return err
	}

	c.Unlock()
	return ioutil.WriteFile(c.filepath, bytes, 0644)
}

// SaveAndClose closes the watcher, and save the config to file
func (c *ConfigManager) SaveAndClose() error {
	closeWatcher()
	return c.Save()
}

// Close closes the watcher
func (c *ConfigManager) Close() {
	closeWatcher()
}

// Watch a specific path, for value changes
func (c *ConfigManager) Watch(path string) *Watcher {
	return c.newWatcher(path)
}

func processTags(config interface{}) error {
	configValue := reflect.Indirect(reflect.ValueOf(config))
	if configValue.Kind() != reflect.Struct {
		return errors.New("invalid config, should be struct")
	}

	configType := configValue.Type()
	for i := 0; i < configType.NumField(); i++ {
		var (
			fieldStruct = configType.Field(i)
			field       = configValue.Field(i)
		)

		if !field.CanAddr() || !field.CanInterface() {
			continue
		}

		if isBlank := reflect.DeepEqual(field.Interface(), reflect.Zero(field.Type()).Interface()); isBlank {
			// Set default configuration if blank
			if value := fieldStruct.Tag.Get("default"); value != "" {
				if err := yaml.Unmarshal([]byte(value), field.Addr().Interface()); err != nil {
					return err
				}
			} else if fieldStruct.Tag.Get("required") == "true" {
				// return error if it is required but blank
				return errors.New(fieldStruct.Name + " is required, but blank")
			}
		}

		for field.Kind() == reflect.Ptr {
			field = field.Elem()
		}

		if field.Kind() == reflect.Struct {
			if err := processTags(field.Addr().Interface()); err != nil {
				return err
			}
		}

		if field.Kind() == reflect.Slice {
			for i := 0; i < field.Len(); i++ {
				if reflect.Indirect(field.Index(i)).Kind() == reflect.Struct {
					if err := processTags(field.Index(i).Addr().Interface()); err != nil {
						return err
					}
				}
			}
		}
	}
	return nil
}
