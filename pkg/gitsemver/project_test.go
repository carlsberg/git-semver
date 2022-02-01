package gitsemver

import (
	"testing"

	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/stretchr/testify/assert"
)

func TestIncrementMajor(t *testing.T) {
	testCommitIncrement(t, "feat!: test", Major)
	testCommitIncrement(t, "fix!: test", Major)
	testCommitIncrement(t, "any!: test", Major)
	testCommitIncrement(t, "breaking change: test", Major)
	testCommitIncrement(t, "BREAKING CHANGE: test", Major)
	testCommitIncrement(t, `some comment
breaking change: test
  `, Major)
}

func TestIncremenMinor(t *testing.T) {
	testCommitIncrement(t, "feat: test", Minor)
}

func TestIncremenPatch(t *testing.T) {
	testCommitIncrement(t, "fix: test", Patch)
}

func TestIncremenNone(t *testing.T) {
	testCommitIncrement(t, "docs: test", None)
	testCommitIncrement(t, "rand: test", None)
	testCommitIncrement(t, "rand: with feat: inside", None)
}

func TestIncremeWithScope(t *testing.T) {
	testCommitIncrement(t, "docs(README): test", None)
	testCommitIncrement(t, "feat(login)!: test", Major)
	testCommitIncrement(t, "feat(users): test", Minor)
	testCommitIncrement(t, "fix(orders): test", Patch)
}

func testCommitIncrement(t *testing.T, msg string, expected Increment) {
	commits := []*object.Commit{{Message: msg}}

	increment, err := resolveIncrement(commits)
	if err != nil {
		t.Error(err)
	}

	assert.Equal(t, expected, increment)
}
