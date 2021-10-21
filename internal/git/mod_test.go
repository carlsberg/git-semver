package git

import (
	"os"
	"path"
	"regexp"
	"testing"

	"github.com/go-git/go-git/v5"
	"github.com/stretchr/testify/assert"
)

func TestCreateCommit(t *testing.T) {
	dir, repo := initTempRepo(t)

	// modify file
	if err := os.WriteFile(path.Join(dir, "README.md"), []byte("test"), os.ModeAppend); err != nil {
		t.Fatal(err)
	}

	hash, err := CreateCommit(repo, "modify README.md")
	if err != nil {
		t.Fatal(err)
	}

	commit, err := repo.CommitObject(hash)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, hash, commit.Hash)
	assert.Equal(t, "modify README.md", commit.Message)

	if _, err := commit.File("README.md"); err != nil {
		t.Fatal(err)
	}
}

func TestListAllCommits(t *testing.T) {
	_, repo := initTempRepo(t)

	commits, err := ListAllCommits(repo)
	if err != nil {
		t.Fatal(err)
	}

	// initTempRepo creates an initial commit
	assert.Equal(t, 1, len(commits))
	assert.Equal(t, "initial commit", commits[0].Message)
}

func TestListCommitsFromTagToHead(t *testing.T) {
	dir, repo := initTempRepo(t)

	if err := os.WriteFile(path.Join(dir, "README.md"), []byte("test"), os.ModeAppend); err != nil {
		t.Fatal(err)
	}

	_, err := CreateCommit(repo, "commit before tagging")
	if err != nil {
		t.Fatal(err)
	}

	if err := CreateTag(repo, "0.0.1", "Release 0.0.1"); err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(path.Join(dir, "README.md"), []byte("test"), os.ModeAppend); err != nil {
		t.Fatal(err)
	}

	_, err = CreateCommit(repo, "modify README.md")
	if err != nil {
		t.Fatal(err)
	}

	commits, err := ListCommitsFromTagToHead(repo, "0.0.1", ".")
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, 1, len(commits))
	assert.Equal(t, "modify README.md", commits[0].Message)
}

func TestCreateTag(t *testing.T) {
	_, repo := initTempRepo(t)

	if err := CreateTag(repo, "0.0.1", "Release 0.0.1"); err != nil {
		t.Fatal(err)
	}

	if _, err := repo.Tag("0.0.1"); err != nil {
		t.Fatal(err)
	}
}

func TestFindTags(t *testing.T) {
	_, repo := initTempRepo(t)

	if err := CreateTag(repo, "0.0.1", "Release 0.0.1"); err != nil {
		t.Fatal(err)
	}

	tags, err := FindTags(repo, regexp.MustCompile("0.0.1"))
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, 1, len(tags))
	assert.Equal(t, "0.0.1", tags[0])
}

func initTempRepo(t *testing.T) (string, *Repository) {
	dir := t.TempDir()

	// initialize repository
	repo, err := git.PlainInit(dir, false)
	if err != nil {
		t.Fatal(err)
	}

	// create README.md file
	file, err := os.Create(path.Join(dir, "README.md"))
	if err != nil {
		t.Fatal(err)
	}

	if _, err := file.WriteString("# Git SemVer Test"); err != nil {
		t.Fatal(err)
	}

	// create initial commit
	tree, err := repo.Worktree()
	if err != nil {
		t.Fatal(err)
	}

	if err := tree.AddGlob("."); err != nil {
		t.Fatal(err)
	}

	if _, err := tree.Commit("initial commit", &git.CommitOptions{All: true}); err != nil {
		t.Fatal(err)
	}

	return dir, repo
}
