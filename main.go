package main

import (
	git_semver "github.com/crqra/git-semver/cmd/git-semver"
	"github.com/spf13/cobra"
)

func main() {
	cobra.CheckErr(git_semver.Execute())
}
