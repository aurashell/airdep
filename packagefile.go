package main

import (
	"encoding/json"
	"os"

	log "github.com/sirupsen/logrus"
)

// PackageFile - a json file describing all versions of a package
type PackageFile struct {
	Packages map[string]map[string]interface{}
}

// LoadPackageFile - loads a package file from a file
func LoadPackageFile(filename string) PackageFile {
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

	return PackageFile{
		Packages: data,
	}
}
