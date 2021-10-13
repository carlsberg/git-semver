package git_semver

import (
	"fmt"
	"io/fs"
	"io/ioutil"
	"regexp"
	"strings"

	"github.com/Masterminds/semver"
)

type VersionFile struct {
	Filename string
	Key      string
	// TODO: cwd     string
}

func NewVersionFile(cwd string, filenameAndKey string) (*VersionFile, error) {
	slice := strings.Split(filenameAndKey, ":")

	if len(slice) != 2 {
		return &VersionFile{}, fmt.Errorf("%s is not correctly formatted. Should be `filename:key`", filenameAndKey)
	}

	return &VersionFile{Filename: slice[0], Key: slice[1]}, nil
}

func (vf VersionFile) UpdateVersion(current *semver.Version, next *semver.Version) error {
	r, err := regexForVersionFileKey(vf.Key, current)
	if err != nil {
		return err
	}

	contents, err := ioutil.ReadFile(vf.Filename)
	if err != nil {
		return err
	}

	match := string(r.Find(contents))
	newVersionString := strings.Replace(match, current.String(), next.String(), 1)
	newContents := strings.Replace(string(contents), match, newVersionString, 1)

	err = ioutil.WriteFile(vf.Filename, []byte(newContents), fs.ModePerm)
	if err != nil {
		return err
	}

	return nil
}

func regexForVersionFileKey(key string, currentVersion *semver.Version) (*regexp.Regexp, error) {
	return regexp.Compile(fmt.Sprintf("%s(.{1,})?%s", key, currentVersion.String()))
}
