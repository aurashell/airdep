package main

import (
	"flag"
	"os"

	log "github.com/sirupsen/logrus"
	// "github.com/Masterminds/semver"
)

func init() {
	log.SetOutput(os.Stdout)
	log.SetLevel(log.DebugLevel)
}

func main() {
	flag.Parse()
	log.Info("Hello, World!")

	_ = ParseManifest("./airdep.json")
}
