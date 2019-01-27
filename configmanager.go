package configmanager

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"os"
	"reflect"

	"github.com/fsnotify/fsnotify"

	"github.com/Eria-Project/logger"
	"github.com/ghodss/yaml"
)

// ConfigManager struct
type ConfigManager struct {
	filepath string
	watcher  *fsnotify.Watcher
}

// Init config manager
func Init(filepath string) (*ConfigManager, error) {
	logger.WithField("file", filepath).Debug("Init config")
	configManager := &ConfigManager{
		filepath: filepath,
	}
	if _, err := os.Stat(filepath); os.IsNotExist(err) {
		return configManager, errors.New("Config file missing")
	}

	return configManager, nil
}

// Load config from file
func (config *ConfigManager) Load(s interface{}) error {
	logger.Debug("Loading config")
	bytes, err := ioutil.ReadFile(config.filepath)
	if err != nil {
		// TODO What to do if file doesn't exists
		return err
	}

	if err := json.Unmarshal(bytes, s); err != nil {
		// TODO What to do if not json file
		return err
	}

	if err := processTags(s); err != nil {
		return err
	}
	logger.Tracef("%+v", s)
	return nil
}

// Save config to file
func (config *ConfigManager) Save(s interface{}) error {
	logger.WithField("file", config.filepath).Debug("Saving config")

	bytes, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}

	return ioutil.WriteFile(config.filepath, bytes, 0644)
}

// SaveAndClose ...
func (config *ConfigManager) SaveAndClose(s interface{}) error {
	//	if config.watcher != nil {
	//	config.watcher.Close()
	//	}
	return config.Save(s)
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
