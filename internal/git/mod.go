package git

import (
	"fmt"
	"log"
	"regexp"

	git "github.com/libgit2/git2go/v31"
)

type Commit = git.Commit
type Repository = git.Repository

func OpenRepository(path string) (*git.Repository, error) {
	return git.OpenRepository(path)
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

	_, err = repo.CreateCommit(head.Name(), sig, sig, message, tree, commitTarget)
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

func FindTags(repo *git.Repository, reg *regexp.Regexp) ([]string, error) {
	var matchTags []string

	tags, err := repo.Tags.List()
	if err != nil {
		return make([]string, 0), nil
	}

	for _, tag := range tags {
		if reg.MatchString(tag) {
			matchTags = append(matchTags, tag)
		}
	}

	return matchTags, nil
}

func CreateTag(repo *git.Repository, tagName, message string) error {
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

	repo.Tags.Create(tagName, latestCommit, sig, message)

	return nil
}

func PushTagToRemotes(repo *git.Repository, tagName string) error {
	remotes, err := repo.Remotes.List()
	if err != nil {
		return err
	}

	if len(remotes) == 0 {
		log.Printf("No remotes found, skipping pushing tag %s\n", tagName)
		return nil
	}

	for _, remote := range remotes {
		repo.Remotes.AddPush(remote, tagName)
	}

	return nil
}
