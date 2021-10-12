package git_semver

import (
	"fmt"
	"io/fs"
	"io/ioutil"
	"log"
	"regexp"
	"sort"
	"strings"

	"github.com/crqra/git-semver/internal/version"
	git "github.com/libgit2/git2go/v31"
)

type Increment int64
type Change string
type VersionFile struct {
	Filename string
	Key      string
}

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

func ListVersions(repo *git.Repository) ([]*version.Version, error) {
	tags, err := repo.Tags.List()
	if err != nil {
		return make([]*version.Version, 0), err
	}

	var versions []*version.Version

	for _, tag := range tags {
		version, err := version.NewVersionFromTag(tag)
		if err != nil {
			continue
		}

		versions = append(versions, version)
	}

	sort.Sort(version.Collection(versions))

	return versions, nil
}

func CreateCommit(repo *git.Repository, message string) error {
	sig, err := repo.DefaultSignature()
	if err != nil {
		return err
	}

	head, err := repo.Head()
	if err != nil {
		return err
	}

	idx, err := repo.Index()
	if err != nil {
		return err
	}

	idx.AddAll(make([]string, 0), git.IndexAddDefault, func(_, _ string) int {
		return 0
	})

	treeId, err := idx.WriteTree()
	if err != nil {
		return err
	}

	err = idx.Write()
	if err != nil {
		return err
	}

	tree, err := repo.LookupTree(treeId)
	if err != nil {
		return err
	}

	commitTarget, err := repo.LookupCommit(head.Target())
	if err != nil {
		return err
	}

	_, err = repo.CreateCommit("refs/heads/main", sig, sig, message, tree, commitTarget)
	if err != nil {
		return err
	}

	return nil
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

func BumpVersion(version *version.Version, increment Increment) *version.Version {
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

func TagVersion(repo *git.Repository, version *version.Version, versionPrefix string) error {
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

	tagName := fmt.Sprintf("%s%s", versionPrefix, version.String())

	repo.Tags.Create(tagName, latestCommit, sig, fmt.Sprintf("Release %s", version.String()))

	return nil
}

func PushVersionTagToRemotes(repo *git.Repository, version *version.Version) error {
	remotes, err := repo.Remotes.List()
	if err != nil {
		return err
	}

	if len(remotes) == 0 {
		log.Printf("No remotes found, skipping pushing tag for version %s\n", version.String())
		return nil
	}

	for _, remote := range remotes {
		repo.Remotes.AddPush(remote, version.String())
	}

	return nil
}

func UpdateVersionFiles(repo *git.Repository, versionFiles []VersionFile, currentVersion *version.Version, nextVersion *version.Version) error {
	if len(versionFiles) == 0 {
		return nil
	}

	for _, versionFile := range versionFiles {
		err := updateVersionFileVersion(versionFile, currentVersion, nextVersion)
		if err != nil {
			return err
		}
	}

	err := CreateCommit(repo, fmt.Sprintf("bump: %s -> %s", currentVersion.String(), nextVersion.String()))
	if err != nil {
		return err
	}

	return nil
}

func regexForVersionFileKey(key string, currentVersion *version.Version) (*regexp.Regexp, error) {
	return regexp.Compile(fmt.Sprintf("%s(.{1,})?%s", key, currentVersion.String()))
}

func updateVersionFileVersion(versionFile VersionFile, currentVersion *version.Version, nextVersion *version.Version) error {
	r, err := regexForVersionFileKey(versionFile.Key, currentVersion)
	if err != nil {
		return err
	}

	contents, err := ioutil.ReadFile(versionFile.Filename)
	if err != nil {
		return err
	}

	match := string(r.Find(contents))
	newVersionString := strings.Replace(match, currentVersion.String(), nextVersion.String(), 1)
	newContents := strings.Replace(string(contents), match, newVersionString, 1)

	err = ioutil.WriteFile(versionFile.Filename, []byte(newContents), fs.ModePerm)
	if err != nil {
		return err
	}

	return nil
}
