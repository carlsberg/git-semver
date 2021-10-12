package git_semver

import (
	"errors"
	"fmt"
	"log"
	"regexp"
	"sort"
	"strings"

	"github.com/Masterminds/semver"
	"github.com/carlsberg/git-semver/internal/git"
)

type Increment int64
type Change string

const (
	Major Increment = 4
	Minor Increment = 3
	Patch Increment = 1
	None  Increment = 0
)

const (
	Feature        Change = "feat"
	Fix            Change = "fix"
	Refactor       Change = "refactor"
	BreakingChange Change = "BREAKING CHANGE"
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

func (p Project) NextVersion() (*semver.Version, error) {
	versions, err := p.Versions()
	if err != nil {
		log.Fatal(err)
	}

	var latestVersion *semver.Version

	versionsLen := len(versions) - 1

	if versionsLen >= 0 {
		latestVersion = versions[versionsLen]
	} else {
		latestVersion, err = semver.NewVersion("0.0.0")
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
		return &semver.Version{}, errors.New("no released versions found")
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
		commits, err = git.ListCommitsInRange(p.repo, TagNameFromProjectAndVersion(p, versions[versionsLen]), "HEAD")
		if err != nil {
			return None, err
		}
	} else {
		commits, err = git.ListCommits(p.repo)
		if err != nil {
			return None, err
		}
	}

	var increment = None

	for _, commit := range commits {
		var commitMessageArr = strings.Split(commit.Message(), ":")

		if len(commitMessageArr) == 0 {
			continue
		}

		var commitIncrement = None
		var commitChange = Change(commitMessageArr[0])

		if commitChange == Feature {
			commitIncrement = Minor
		}

		if commitChange == Fix {
			commitIncrement = Patch
		}

		if commitChange == Refactor || commitChange == BreakingChange {
			commitIncrement = Major
		}

		if commitIncrement > increment {
			increment = commitIncrement
		}
	}

	return increment, nil
}

func TagNameFromProjectAndVersion(p *Project, v *semver.Version) string {
	if p.IsSubProject() {
		return fmt.Sprintf("%s/%s", p.dir, v.Original())
	}

	return v.Original()
}
