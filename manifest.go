package main

import (
	"encoding/json"
	"os"

	"github.com/Masterminds/semver"
	log "github.com/sirupsen/logrus"
)

// ManifestRepositoryInfo -
type ManifestRepositoryInfo struct {
	SrcType  string
	SrcValue string

	Values map[string]interface{}
}

// Manifest - decribes repositories and packages we want to use.
type Manifest struct {
	Wants        map[string]string
	Repositories map[string]ManifestRepositoryInfo

	Packages    map[string]map[string]map[string]interface{}
	OutPackages map[string]map[string]interface{}
}

// ParseManifest - parses a manifest file
func ParseManifest(filename string) Manifest {
	log.WithFields(log.Fields{"From": filename}).Info("Loading manifest")

	file, err := os.Open(filename)
	defer file.Close()
	if err != nil {
		log.Panic(err)
	}

	var data map[string]interface{}

	jsonParser := json.NewDecoder(file)
	err = jsonParser.Decode(&data)
	if err != nil {
		log.Panic(err)
	}

	mWants := map[string]string{}
	mRepositories := map[string]ManifestRepositoryInfo{}

	if wants, ok := data["wants"]; ok {
		wants := wants.(map[string]interface{})
		for w, v := range wants {
			mWants[w] = v.(string)
		}
	}

	if repositories, ok := data["repositories"]; ok {
		repositories := repositories.(map[string]interface{})
		for reponame, repo := range repositories {

			repo := repo.([]interface{})
			repopath := repo[0].(map[string]interface{})
			repovals := repo[1].(map[string]interface{})

			var r = ManifestRepositoryInfo{
				Values: repovals,
			}

			if _, ok := repopath["file"]; ok {
				r.SrcType = "file"
				r.SrcValue = repopath["file"].(string)
			}

			for k, v := range repovals {
				r.Values[k] = v.(string)
			}

			mRepositories[reponame] = r
		}
	}

	manifest := Manifest{
		Wants:        mWants,
		Repositories: mRepositories,

		Packages:    map[string]map[string]map[string]interface{}{},
		OutPackages: map[string]map[string]interface{}{},
	}

	for rn, rv := range manifest.Repositories {
		log.WithFields(log.Fields{"Name": rn}).Info("Loading repository")
		r := RepositoryFromInfo(rv)
		for pn, pf := range r.PackageFiles {
			for pv, p := range pf.Packages {
				if l, ok := manifest.Packages[pn]; ok {
					l[pv] = p
					manifest.Packages[pn] = l
				} else {
					manifest.Packages[pn] = map[string]map[string]interface{}{pv: p}
				}
			}
		}
	}

	for pn, pc := range manifest.Wants {
		manifest.Filter(pn, pc)
	}

	for pn := range manifest.Wants {
		manifest.Resolve(pn, []string{})
	}

	for pn := range manifest.Wants {
		manifest.Merge(pn, []string{})
	}

	manifest.Dump()

	return manifest
}

func strSliceHas(sl []string, s string) bool {
	for _, ss := range sl {
		if ss == s {
			return true
		}
	}
	return false
}

// Latest - gets latest version of a package
func (m *Manifest) Latest(pn string) *map[string]interface{} {
	var v *semver.Version
	var ap *map[string]interface{}

	for pv, p := range m.Packages[pn] {
		nv, err := semver.NewVersion(pv)
		if err != nil {
			log.Panic(err)
		}
		if v == nil || nv.GreaterThan(nv) {
			v = nv
			ap = &p
		}
	}

	return ap
}

// Filter - filters packages by version
func (m *Manifest) Filter(pn string, vc string) {
	c, err := semver.NewConstraint(vc)
	if err != nil {
		log.Panic(err)
	}

	del := []string{}
	for pv := range m.Packages[pn] {
		v, err := semver.NewVersion(pv)
		if err != nil {
			log.Panic(err)
		}
		if !c.Check(v) {
			del = append(del, pv)
		}
	}
	for _, pv := range del {
		delete(m.Packages[pn], pv)
	}

	if len(m.Packages[pn]) == 0 {
		log.WithFields(log.Fields{"Package": pn}).Panic("No versions of package left")
	}
}

// Resolve - resolves packages' dependencies
func (m *Manifest) Resolve(pn string, parents []string) {
	if strSliceHas(parents, pn) {
		return
	}
	p := m.Latest(pn)
	if wants, ok := (*p)["wants"]; ok {
		wants := wants.(map[string]interface{})
		for dn, dv := range wants {
			dv := dv.(string)
			m.Filter(dn, dv)
		}
		for dn := range wants {
			m.Resolve(dn, append(parents, pn))
		}
	}
}

// Merge - merges package and their deps into manifest out
func (m *Manifest) Merge(pn string, parents []string) {
	if strSliceHas(parents, pn) {
		return
	}
	p := m.Latest(pn)
	m.OutPackages[pn] = *p
	if wants, ok := (*p)["wants"]; ok {
		wants := wants.(map[string]interface{})
		for dn := range wants {
			m.Merge(dn, append(parents, pn))
		}
	}
}

// Dump - dumps the manifest to a file
func (m *Manifest) Dump() {
	out := map[string]interface{}{}

	wants := []string{}
	for pn := range m.Wants {
		wants = append(wants, pn)
	}
	out["wants"] = wants

	packages := map[string]interface{}{}
	for pn, p := range m.OutPackages {
		if wants, ok := p["wants"]; ok {
			wants := wants.(map[string]interface{})
			nw := []string{}
			for dn := range wants {
				nw = append(nw, dn)
			}
			p["wants"] = nw
		}
		packages[pn] = p
	}
	out["packages"] = packages

	file, err := os.Create("./airbuild.json")
	defer file.Close()
	if err != nil {
		log.Panic(err)
	}

	e := json.NewEncoder(file)
	e.SetIndent("", "  ")
	err = e.Encode(out)
	if err != nil {
		log.Panic(err)
	}

	log.Info("Generated airbuild.json")
}
