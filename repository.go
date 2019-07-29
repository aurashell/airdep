package main

import (
	"encoding/json"
	"os"

	log "github.com/sirupsen/logrus"
)

// Repository - describe a repository with all its packages
type Repository struct {
	Variables    []string
	PackageFiles map[string]PackageFile
}

// RepositoryFromInfo - creates a new repository from manifest info
func RepositoryFromInfo(i ManifestRepositoryInfo) Repository {
	variables := []string{}
	packageFiles := map[string]PackageFile{}

	var data map[string]interface{}

	if i.SrcType == "file" {
		file, err := os.Open(i.SrcValue)
		defer file.Close()
		if err != nil {
			log.Panic(err)
		}

		jsonParser := json.NewDecoder(file)
		err = jsonParser.Decode(&data)
		if err != nil {
			log.Panic(err)
		}
	} else {
		log.WithFields(log.Fields{"src": i.SrcType}).Panic("Invalid source type!")
		os.Exit(1)
	}

	if pfm, ok := data["package-files"]; ok {
		pfm := pfm.(map[string]interface{})
		for k, v := range pfm {
			v := v.(string)
			pf := LoadPackageFile(v)
			packageFiles[k] = pf
		}
	} else {
		log.Warning("No package-files!")
	}

	return Repository{
		Variables:    variables,
		PackageFiles: packageFiles,
	}
}
