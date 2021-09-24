package git_semver

import (
	"fmt"
	"log"
	"regexp"
	"sort"
	"strings"

	"github.com/Masterminds/semver"
	git "github.com/libgit2/git2go/v31"
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

func OpenRepository(path string) (*git.Repository, error) {
	return git.OpenRepository(path)
}

func ListVersions(repo *git.Repository) ([]*semver.Version, error) {
	tags, err := repo.Tags.List()
	if err != nil {
		return make([]*semver.Version, 0), err
	}

	var versions []*semver.Version

	reg, err := regexp.Compile(semver.SemVerRegex)
	if err != nil {
		return make([]*semver.Version, 0), err
	}

	for _, tag := range tags {
		if !reg.MatchString(tag) {
			continue
		}

		version, err := semver.NewVersion(tag)
		if err != nil {
			return make([]*semver.Version, 0), err
		}

		versions = append(versions, version)
	}

	sort.Sort(semver.Collection(versions))

	return versions, nil
}

func ListCommits(repo *git.Repository) ([]*git.Commit, error) {
	revwalk, err := repo.Walk()
	if err != nil {
		return make([]*git.Commit, 0), err
	}

	if err := revwalk.PushHead(); err != nil {
		return make([]*git.Commit, 0), err
	}

	var commits []*git.Commit

	err = revwalk.Iterate(func(commit *git.Commit) bool {
		commits = append(commits, commit)
		return true
	})
	if err != nil {
		return make([]*git.Commit, 0), err
	}

	return commits, nil
}

func ListCommitsInRange(repo *git.Repository, lRange string, rRange string) ([]*git.Commit, error) {
	revwalk, err := repo.Walk()
	if err != nil {
		return make([]*git.Commit, 0), err
	}

	if err := revwalk.PushRange(fmt.Sprintf("%s..%s", lRange, rRange)); err != nil {
		return make([]*git.Commit, 0), err
	}

	var commits []*git.Commit

	err = revwalk.Iterate(func(commit *git.Commit) bool {
		commits = append(commits, commit)
		return true
	})
	if err != nil {
		return make([]*git.Commit, 0), err
	}

	return commits, nil
}

func DetectIncrement(commits []*git.Commit) Increment {
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

	return increment
}

func BumpVersion(version semver.Version, increment Increment) semver.Version {
	switch increment {
	case Major:
		return version.IncMajor()

	case Minor:
		return version.IncMinor()

	case Patch:
		return version.IncPatch()
	}

	return version
}

func TagVersion(repo *git.Repository, version semver.Version) error {
	sig, err := repo.DefaultSignature()
	if err != nil {
		return err
	}

	latestCommitObject, err := repo.RevparseSingle("HEAD")
	if err != nil {
		return err
	}

	latestCommit, err := latestCommitObject.AsCommit()
	if err != nil {
		return err
	}

	repo.Tags.Create(version.String(), latestCommit, sig, fmt.Sprintf("Release %s", version.String()))

	return nil
}

func PushVersionTagToRemotes(repo *git.Repository, version semver.Version) error {
	remotes, err := repo.Remotes.List()
	if err != nil {
		return err
	}

	if len(remotes) == 0 {
		log.Printf("No remotes found, skipping pushing tag %s\n", version.String())
		return nil
	}

	for _, remote := range remotes {
		repo.Remotes.AddPush(remote, version.String())
	}

	return nil
}
