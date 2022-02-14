package gitsemver

import (
	"fmt"
	"log"
	"regexp"
	"sort"

	"github.com/Masterminds/semver"
	"github.com/carlsberg/git-semver/internal/git"
)

type Increment int64
type Change string
type AuthMethod = git.AuthMethod
type BasicAuth = git.BasicAuth

const (
	Major Increment = 4
	Minor Increment = 3
	Patch Increment = 1
	None  Increment = 0
)

var (
	breaking     = regexp.MustCompile("(?im)^breaking change:.*")
	breakingBang = regexp.MustCompile(`(?i)^(\w+)(\(.*\))?!:.*`)
	feature      = regexp.MustCompile(`(?i)^feat(\(.*\))?:.*`)
	patch        = regexp.MustCompile(`(?i)^fix(\(.*\))?:.*`)
)

type Project struct {
	dir  string
	repo *git.Repository
}

func NewProject(root string, dir string) (*Project, error) {
	repo, err := git.OpenRepository(root)
	if err != nil {
		return &Project{}, err
	}

	return &Project{
		dir:  dir,
		repo: repo,
	}, nil
}

func (p Project) IsSubProject() bool {
	return p.dir != "" && p.dir != "/"
}

func (p Project) Dir() string {
	return p.dir
}

func (p Project) Repo() *git.Repository {
	return p.repo
}

func (p Project) Tags() ([]string, error) {
	var reg *regexp.Regexp
	var err error

	if p.IsSubProject() {
		reg, err = regexp.Compile(fmt.Sprintf("^%s/%s$", p.dir, semver.SemVerRegex))
	} else {
		reg, err = regexp.Compile(fmt.Sprintf("^%s$", semver.SemVerRegex))
	}

	if err != nil {
		return make([]string, 0), err
	}

	tags, err := git.FindTags(p.repo, reg)
	if err != nil {
		return make([]string, 0), err
	}

	return tags, nil
}

func (p Project) Versions() ([]*semver.Version, error) {
	tags, err := p.Tags()
	if err != nil {
		return make([]*semver.Version, 0), err
	}

	var versions []*semver.Version

	reg, err := regexp.Compile(semver.SemVerRegex)
	if err != nil {
		return make([]*semver.Version, 0), err
	}

	for _, tag := range tags {
		version, err := semver.NewVersion(reg.FindString(tag))
		if err != nil {
			return make([]*semver.Version, 0), err
		}

		versions = append(versions, version)
	}

	sort.Sort(semver.Collection(versions))

	return versions, nil
}

func (p Project) NextVersion(vPrefix bool) (*semver.Version, error) {
	versions, err := p.Versions()
	if err != nil {
		log.Fatal(err)
	}

	var latestVersion *semver.Version

	versionsLen := len(versions) - 1

	if versionsLen >= 0 {
		latestVersion = versions[versionsLen]
	} else {
		latestVersionStr := "0.0.0"

		if vPrefix {
			latestVersionStr = fmt.Sprintf("v%s", latestVersionStr)
		}

		latestVersion, err = semver.NewVersion(latestVersionStr)
		if err != nil {
			return &semver.Version{}, err
		}
	}

	increment, err := p.NextVersionIncrement()
	if err != nil {
		return &semver.Version{}, err
	}

	var nextVersion = *latestVersion

	switch increment {
	case Major:
		nextVersion = latestVersion.IncMajor()

	case Minor:
		nextVersion = latestVersion.IncMinor()

	case Patch:
		nextVersion = latestVersion.IncPatch()
	}

	return &nextVersion, nil
}

func (p Project) LatestVersion() (*semver.Version, error) {
	versions, err := p.Versions()
	if err != nil {
		return &semver.Version{}, err
	}

	versionsLen := len(versions) - 1

	if versionsLen < 0 {
		return semver.NewVersion("0.0.0")
	}

	return versions[versionsLen], nil
}

func (p *Project) NextVersionIncrement() (Increment, error) {
	versions, err := p.Versions()
	if err != nil {
		log.Fatal(err)
	}

	versionsLen := len(versions) - 1

	var commits []*git.Commit

	if versionsLen >= 0 {
		commits, err = git.ListCommitsFromTagToHead(p.repo, TagNameFromProjectAndVersion(p, versions[versionsLen]), ".")
		if err != nil {
			return None, err
		}
	} else {
		commits, err = git.ListAllCommits(p.repo)
		if err != nil {
			return None, err
		}
	}

	increment, err := resolveIncrement(commits)
	if err != nil {
		return None, err
	}

	return increment, nil
}

func (p *Project) Bump(versionFilenamesAndKeys []string, auth AuthMethod, vPrefix, skipTag bool) error {
	latest, err := p.LatestVersion()
	if err != nil {
		return err
	}

	next, err := p.NextVersion(vPrefix)
	if err != nil {
		return err
	}

	if len(versionFilenamesAndKeys) > 0 {
		for _, filenameAndKey := range versionFilenamesAndKeys {
			vf, err := NewVersionFile("./", filenameAndKey)
			if err != nil {
				return err
			}

			vf.UpdateVersion(latest, next)
			if err != nil {
				return err
			}
		}

		if _, err := git.CreateCommit(p.Repo(), fmt.Sprintf("bump: %s -> %s", latest.String(), next.String())); err != nil {
			return err
		}
	}

	if !skipTag {
		tagName := TagNameFromProjectAndVersion(p, next)
		tagMessage := fmt.Sprintf("Release %s", tagName)

		if err := git.CreateTag(p.Repo(), tagName, tagMessage); err != nil {
			return err
		}
	}

	if !skipTag || len(versionFilenamesAndKeys) > 0 {
		if err := git.PushToOrigin(p.Repo(), auth); err != nil {
			return err
		}
	}

	return nil
}

func TagNameFromProjectAndVersion(p *Project, v *semver.Version) string {
	if p.IsSubProject() {
		return fmt.Sprintf("%s/%s", p.dir, v.Original())
	}

	return v.Original()
}

func resolveIncrement(commits []*git.Commit) (Increment, error) {
	var increment = None

	for _, commit := range commits {
		commitIncrement := None

		if breaking.MatchString(commit.Message) || breakingBang.MatchString(commit.Message) {
			commitIncrement = Major
		}

		if feature.MatchString(commit.Message) {
			commitIncrement = Minor
		}

		if patch.MatchString(commit.Message) {
			commitIncrement = Patch
		}

		if commitIncrement > increment {
			increment = commitIncrement
		}
	}

	return increment, nil
}
