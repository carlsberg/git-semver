package version

import (
	"regexp"

	"github.com/Masterminds/semver"
)

type Version struct {
	version *semver.Version
	tag     string
}

func NewVersion(v string) (*Version, error) {
	sv, err := semver.NewVersion(v)
	if err != nil {
		return &Version{}, nil
	}

	return &Version{version: sv, tag: v}, nil
}

func NewVersionFromTag(tag string) (*Version, error) {
	reg, err := regexp.Compile(semver.SemVerRegex)
	if err != nil {
		return &Version{}, err
	}

	if !reg.MatchString(tag) {
		return &Version{}, err
	}

	version, err := semver.NewVersion(string(reg.Find([]byte(tag))))
	if err != nil {
		return &Version{}, err
	}

	return &Version{version: version, tag: tag}, nil
}

func (v *Version) Tag() string {
	return v.tag
}

func (v *Version) String() string {
	return v.version.String()
}

func (v *Version) IncMajor() *Version {
	nextVersion := v.version.IncMajor()
	v.version = &nextVersion

	return v
}

func (v *Version) IncMinor() *Version {
	nextVersion := v.version.IncMinor()
	v.version = &nextVersion

	return v
}

func (v *Version) IncPatch() *Version {
	nextVersion := v.version.IncPatch()
	v.version = &nextVersion

	return v
}
