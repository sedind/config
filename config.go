package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"

	"github.com/fatih/camelcase"
)

var errWrongConfigurationType = errors.New("Configuration type must be a pointer to a struct")

// LoadConfig reads configuration from path and stores it to obj interface
// The format is deduced from the file extension
//	* .json    - is decoded as json
//	* .yml     - is decoded as yaml
func LoadConfig(path string, obj interface{}) error {
	err := checkConfigObj(obj)
	if err != nil {
		return err
	}

	_, err = os.Stat(path)
	if err != nil {
		return err
	}

	data, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}

	switch filepath.Ext(path) {
	case ".json":
		err := json.Unmarshal(data, obj)
		if err != nil {
			return err
		}
	}

	return nil
}

// SyncEnv overrides obj field's values that are set in the environment.
//
// The environment variable names are derived from config fields by underscoring, and uppercasing
// the name. E.g. AppName will have a corresponding environment variable APP_NAME
//
// NOTE only int, string and bool fields are supported and the corresponding values are set.
// when the field value is not supported it is ignored.
func SyncEnv(obj interface{}) error {
	err := checkConfigObj(obj)
	if err != nil {
		return err
	}

	cfg := reflect.ValueOf(obj).Elem()
	cfgType := cfg.Type()

	for k := range make([]struct{}, cfgType.NumField()) {
		field := cfgType.Field(k)

		cm := getEnvName(field.Name)
		env := os.Getenv(cm)
		if env == "" {
			continue
		}

		switch field.Type.Kind() {
		case reflect.String:
			cfg.FieldByName(field.Name).SetString(env)
		case reflect.Int:
			v, err := strconv.Atoi(env)
			if err != nil {
				return fmt.Errorf(" Error loading config field %s %v", field.Name, err)
			}
			cfg.FieldByName(field.Name).Set(reflect.ValueOf(v))
		case reflect.Bool:
			b, err := strconv.ParseBool(env)
			if err != nil {
				return fmt.Errorf(" Error loading config field %s %v", field.Name, err)
			}
			cfg.FieldByName(field.Name).SetBool(b)
		}
	}
	return nil
}

// LoadAndSync reads configuration from path and stores it to obj interface
// and syncs config with environment variables
func LoadAndSync(path string, obj interface{}) error {
	err := LoadConfig(path, obj)
	if err != nil {
		return err
	}

	err = SyncEnv(obj)
	if err != nil {
		return err
	}

	return nil
}

// getEnvName returns all upper case and underscore separated string, from field.
// field is a camel case string.
//
// example
//	AppName will change to APP_NAME
func getEnvName(field string) string {
	camSplit := camelcase.Split(field)
	var rst string
	for k, v := range camSplit {
		if k == 0 {
			rst = strings.ToUpper(v)
			continue
		}
		rst = rst + "_" + strings.ToUpper(v)
	}
	return rst
}

func checkConfigObj(obj interface{}) error {
	// check if type is a pointer
	objVal := reflect.ValueOf(obj)
	if objVal.Kind() != reflect.Ptr || objVal.IsNil() {
		return errWrongConfigurationType
	}

	// get and configrm struct value
	objVal = objVal.Elem()
	if objVal.Kind() != reflect.Struct {
		return errWrongConfigurationType
	}
	return nil
}
