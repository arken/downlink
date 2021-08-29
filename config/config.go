package config

import (
	"os"
	"os/user"
	"path/filepath"
	"reflect"
	"strings"
)

var (
	Global  config
	Version string = "develop"
)

type config struct {
	Database database
	Ipfs     ipfs
	Manifest manifest
	Web      web
}

type database struct {
	Path string
}

type manifest struct {
	Url  string
	Path string
}

type web struct {
	Addr string
}

type ipfs struct {
	Path       string
	PrivateKey string
	PeerID     string
	Addr       string
}

func Init() error {
	// Check who's the current user to find their home directory.
	user, err := user.Current()
	if err != nil {
		return err
	}

	// Generate Default Config
	Global = config{
		Database: database{
			Path: "downlink.db",
		},
		Manifest: manifest{
			Url:  "https://github.com/arken/core-manifest.git",
			Path: filepath.Join(user.HomeDir, ".downlink", "manifest"),
		},
		Web: web{
			Addr: ":8080",
		},
		Ipfs: ipfs{
			Path:       filepath.Join(user.HomeDir, ".downlink", "ipfs"),
			PeerID:     "",
			PrivateKey: "",
			Addr:       "",
		},
	}
	err = parseConfigEnv(&Global)
	if err != nil {
		return err
	}
	return err
}

func parseConfigEnv(input *config) error {
	numSubStructs := reflect.ValueOf(input).Elem().NumField()
	for i := 0; i < numSubStructs; i++ {
		iter := reflect.ValueOf(input).Elem().Field(i)
		subStruct := strings.ToUpper(iter.Type().Name())
		structType := iter.Type()
		for j := 0; j < iter.NumField(); j++ {
			fieldVal := iter.Field(j).String()
			fieldName := structType.Field(j).Name
			evName := "DOWNLINK" + "_" + subStruct + "_" + strings.ToUpper(fieldName)
			evVal, evExists := os.LookupEnv(evName)
			if evExists && evVal != fieldVal {
				iter.FieldByName(fieldName).SetString(evVal)
			}
		}
	}
	return nil
}
