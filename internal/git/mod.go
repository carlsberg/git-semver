package git

import (
	"fmt"
	"os"
	"regexp"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/storer"
)

type Commit = object.Commit
type Hash = plumbing.Hash
type Repository = git.Repository

func OpenRepository(path string) (*Repository, error) {
	return git.PlainOpen(path)
}

func CreateCommit(repo *Repository, message string) (Hash, error) {
	tree, err := repo.Worktree()
	if err != nil {
		return plumbing.Hash{}, err
	}

	commitOpts := &git.CommitOptions{
		All: true,
	}

	return tree.Commit(message, commitOpts)
}

func ListAllCommits(repo *Repository) ([]*Commit, error) {
	iter, err := repo.CommitObjects()
	if err != nil {
		return nil, err
	}

	defer iter.Close()

	var commits []*Commit

	err = iter.ForEach(func(c *Commit) error {
		commits = append(commits, c)
		return nil
	})
	if err != nil {
		return nil, err
	}

	return commits, nil
}

func ListCommitsFromTagToHead(repo *Repository, tag string, pathPrefix string) ([]*Commit, error) {
	start, err := repo.ResolveRevision(plumbing.Revision(fmt.Sprintf("refs/tags/%s", tag)))
	if err != nil {
		return nil, err
	}

	end, err := repo.ResolveRevision("HEAD")
	if err != nil {
		return nil, err
	}

	iter, err := repo.Log(&git.LogOptions{From: *end})
	if err != nil {
		return nil, err
	}

	var commits []*Commit

	err = iter.ForEach(func(c *object.Commit) error {
		if c.Hash == *start {
			return storer.ErrStop
		}

		commits = append(commits, c)

		return nil
	})
	if err != nil {
		return nil, err
	}

	return commits, nil
}

func FindTags(repo *Repository, reg *regexp.Regexp) ([]string, error) {
	var matchTags []string

	tags, err := repo.Tags()
	if err != nil {
		return nil, err
	}

	defer tags.Close()

	err = tags.ForEach(func(ref *plumbing.Reference) error {
		if reg.MatchString(ref.Name().Short()) {
			matchTags = append(matchTags, ref.Name().Short())
		}

		return nil
	})

	return matchTags, nil
}

func CreateTag(repo *Repository, name, message string) error {
	ref, err := repo.Head()
	if err != nil {
		return err
	}

	createTagOpts := &git.CreateTagOptions{
		Message: message,
	}

	_, err = repo.CreateTag(name, ref.Hash(), createTagOpts)

	return err
}

func PushTagsToOrigin(repo *Repository) error {
	// skip pushing if remote doesn't exist
	if _, err := repo.Remote("origin"); err != nil {
		return nil
	}

	pushOpts := &git.PushOptions{
		RemoteName: "origin",
		Progress:   os.Stdout,
		RefSpecs:   []config.RefSpec{config.RefSpec("refs/tags/*:refs/tags/*")},
	}

	if err := repo.Push(pushOpts); err != nil {
		return err
	}

	return nil
}
