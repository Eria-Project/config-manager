package configmanager

import (
	"reflect"
	"testing"

	"github.com/fsnotify/fsnotify"
)

type testStruct struct {
	A string `default:"A"`
	B uint   `default:"1"`
	C bool   //`default:"true"` TOFIX
	D struct {
		D1 string
	}
	E []struct {
		E1 string
	}
	F string `required:"true"`
}

func TestInit(t *testing.T) {
	type args struct {
		filepath string
	}
	tests := []struct {
		name    string
		args    args
		want    *ConfigManager
		wantErr bool
	}{
		{
			name:    "Existing file",
			args:    args{filepath: "./test/fileA.json"},
			want:    &ConfigManager{filepath: "./test/fileA.json"},
			wantErr: false,
		},
		{
			name:    "Missing file",
			args:    args{filepath: "dummy.json"},
			want:    &ConfigManager{filepath: "dummy.json"},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Init(tt.args.filepath)
			if (err != nil) != tt.wantErr {
				t.Errorf("Init() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Init() = %+v, want %+v", got, tt.want)
			}
		})
	}
}

func TestConfigManager_Load(t *testing.T) {
	type fields struct {
		filepath string
		watcher  *fsnotify.Watcher
	}
	type args struct {
		s interface{}
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *testStruct
		wantErr bool
	}{
		{
			name: "Valid Json file",
			fields: fields{
				filepath: "./test/fileA.json",
			},
			args: args{
				s: &testStruct{},
			},
			want: &testStruct{
				A: "X",
				B: 9,
				C: false,
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
			},
			wantErr: false,
		},
		{
			name: "Default values",
			fields: fields{
				filepath: "./test/fileB.json",
			},
			args: args{
				s: &testStruct{},
			},
			want: &testStruct{
				A: "A",
				B: 1,
				F: "V",
			},
			wantErr: false,
		},
		{
			name: "Incorrect JSON",
			fields: fields{
				filepath: "./test/fileC.json",
			},
			args: args{
				s: &testStruct{},
			},
			want:    &testStruct{},
			wantErr: true,
		},
		{
			name: "Required values",
			fields: fields{
				filepath: "./test/fileD.json",
			},
			args: args{
				s: &testStruct{},
			},
			want: &testStruct{
				A: "A",
				B: 1,
				F: "",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &ConfigManager{
				filepath: tt.fields.filepath,
				watcher:  tt.fields.watcher,
			}
			if err := config.Load(tt.args.s); (err != nil) != tt.wantErr {
				t.Errorf("ConfigManager.Load() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !reflect.DeepEqual(tt.args.s, tt.want) {
				t.Errorf("ConfigManager.Load() = %+v, want %+v", tt.args.s, tt.want)
			}
		})
	}
}
