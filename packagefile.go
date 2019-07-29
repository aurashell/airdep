package main

import (
	"encoding/json"
	"os"
	"strings"

	log "github.com/sirupsen/logrus"
)

// PackageFile - a json file describing all versions of a package
type PackageFile struct {
	Packages map[string]map[string]interface{}
}

// LoadPackageFile - loads a package file from a file
func LoadPackageFile(filename string, replace map[string]interface{}) PackageFile {
	var data map[string]map[string]interface{}

	file, err := os.Open(filename)
	defer file.Close()
	if err != nil {
		log.Panic(err)
	}

	jsonParser := json.NewDecoder(file)
	err = jsonParser.Decode(&data)
	if err != nil {
		log.Panic(err)
	}

	data = replaceValues(data, replace).(map[string]map[string]interface{})

	return PackageFile{
		Packages: data,
	}
}

func replaceString(val string, replacements map[string]interface{}) string {
	for k, v := range replacements {
		val = strings.ReplaceAll(val, "{"+k+"}", v.(string))
	}
	return val
}

func replaceValues(val interface{}, replacements map[string]interface{}) interface{} {
	switch val.(type) {
	case string:
		val = replaceString(val.(string), replacements)
		break
	case []interface{}:
		v := val.([]interface{})
		for i := range v {
			v[i] = replaceValues(v[i], replacements)
		}
		break
	case map[string]interface{}:
		v := val.(map[string]interface{})
		for k := range v {
			v[k] = replaceValues(v[k], replacements)
		}
		break
	case map[string]map[string]interface{}:
		v := val.(map[string]map[string]interface{})
		for k := range v {
			v[k] = replaceValues(v[k], replacements).(map[string]interface{})
		}
		break
	}
	return val
}
