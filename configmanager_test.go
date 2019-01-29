package configmanager

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"reflect"
	"strings"
	"testing"
)

type testStruct struct {
	A string `default:"A"`
	B uint   `default:"1"`
	C bool   `default:"true"`
	D struct {
		D1 string
	}
	E []struct {
		E1 string
	}
	F string `required:"true"`
}

func generateDefaultConfig() testStruct {
	config := testStruct{
		A: "A",
		B: 1,
		C: true,
		D: struct {
			D1 string
		}{
			D1: "Y",
		},
		E: []struct {
			E1 string
		}{
			{
				E1: "Z",
			},
			{
				E1: "W",
			},
		},
		F: "V",
	}
	return config
}

func TestInit(t *testing.T) {
	currentEnv := os.Getenv("ERIA_PATH") // Save current env

	// Create dummy file for file exist check
	file, err := ioutil.TempFile("", "test.json")
	if err == nil {
		defer file.Close()
		defer os.Remove(file.Name())
		file.Write([]byte{0})
	}

	path := os.TempDir()
	fileName := strings.TrimPrefix(file.Name(), path)
	fmt.Println(fileName)
	fmt.Println(path)

	type args struct {
		fileName string
	}
	tests := []struct {
		name    string
		args    args
		want    *ConfigManager
		wantErr bool
		env     string
	}{
		{
			name:    "Existing file",
			args:    args{fileName: fileName},
			want:    &ConfigManager{filepath: file.Name()},
			wantErr: false,
			env:     path,
		},
		{
			name:    "Missing file",
			args:    args{fileName: "test.json"},
			wantErr: true,
			env:     path,
		},
		{
			name:    "Missing env",
			args:    args{fileName: "test.json"},
			wantErr: true,
			env:     "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Setenv("ERIA_PATH", tt.env)
			got, err := Init(tt.args.fileName)
			if (err != nil) != tt.wantErr {
				t.Errorf("Init() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Init() = %+v, want %+v", got, tt.want)
			}
		})
	}
	os.Setenv("ERIA_PATH", currentEnv) // Restore current env
}

func TestConfigManager_Load_ValidJson(t *testing.T) {
	config := generateDefaultConfig()
	if bytes, err := json.Marshal(config); err == nil {
		if file, err := ioutil.TempFile("", "test.json"); err == nil {
			defer file.Close()
			defer os.Remove(file.Name())
			file.Write(bytes)

			configmanager := &ConfigManager{
				filepath: file.Name(),
				watcher:  nil,
			}

			var result testStruct
			if err := configmanager.Load(&result); err != nil {
				t.Errorf("Load_ValidJson() Error: %s", err)
			}
			if !reflect.DeepEqual(result, config) {
				t.Errorf("Load_ValidJson() %+v, want %+v", result, config)
			}
		}
	} else {
		t.Errorf("Load_ValidJson() failed to marshal config")
	}
}

func TestConfigManager_Load_InvalidJson(t *testing.T) {
	if file, err := ioutil.TempFile("", "test.json"); err == nil {
		defer file.Close()
		defer os.Remove(file.Name())
		file.Write([]byte{0})

		configmanager := &ConfigManager{
			filepath: file.Name(),
			watcher:  nil,
		}

		var result testStruct
		if err := configmanager.Load(&result); err == nil {
			t.Errorf("Load_InvalidJson() should return an error")
		}
	}
}

func TestConfigManager_Load_Required(t *testing.T) {
	config := generateDefaultConfig()
	config.F = ""
	if bytes, err := json.Marshal(config); err == nil {
		if file, err := ioutil.TempFile("", "test.json"); err == nil {
			defer file.Close()
			defer os.Remove(file.Name())
			file.Write(bytes)

			configmanager := &ConfigManager{
				filepath: file.Name(),
				watcher:  nil,
			}

			var result testStruct
			if err := configmanager.Load(&result); err.Error() != "F is required, but blank" {
				t.Errorf("Load_Required() Doesn't returns the correct error: %s", err)
			}
		}
	} else {
		t.Errorf("Load_Required() failed to marshal config")
	}
}

func TestConfigManager_Load_Default(t *testing.T) {
	config := generateDefaultConfig()
	config.A = ""
	config.B = 0
	config.C = false
	if bytes, err := json.Marshal(config); err == nil {
		if file, err := ioutil.TempFile("", "test.json"); err == nil {
			defer file.Close()
			defer os.Remove(file.Name())
			file.Write(bytes)

			configmanager := &ConfigManager{
				filepath: file.Name(),
				watcher:  nil,
			}

			var result testStruct
			if err := configmanager.Load(&result); err != nil {
				t.Errorf("Load_Default() Error: %s", err)
			}
			if !reflect.DeepEqual(result, generateDefaultConfig()) {
				t.Errorf("Load_Default() %+v, want %+v", result, generateDefaultConfig())
			}
		}
	} else {
		t.Errorf("Load_Default() failed to marshal config")
	}
}

func TestConfigManager_Save(t *testing.T) {
	if file, err := ioutil.TempFile("", "test.json"); err == nil {
		defer os.Remove(file.Name())

		configmanager := &ConfigManager{
			filepath: file.Name(),
			watcher:  nil,
		}

		var result testStruct
		if err := configmanager.Save(&result); err != nil {
			t.Errorf("Load_Default() Error: %s", err)
		}
		if _, err := os.Stat(file.Name()); os.IsNotExist(err) {
			t.Errorf("Config file '%s' missing", file.Name())
		}
	}
}
