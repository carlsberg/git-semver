package main

import (
	"github.com/carlsberg/git-semver/cmd/gitsemver"
	"github.com/spf13/cobra"
)

func main() {
	cobra.CheckErr(gitsemver.Execute())
}
