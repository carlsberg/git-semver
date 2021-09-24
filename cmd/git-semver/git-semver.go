package git_semver

import (
	"fmt"
	"log"

	"github.com/Masterminds/semver"
	git_semver "github.com/crqra/git-semver/internal/git-semver"
	git "github.com/libgit2/git2go/v31"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "git-semver",
	Short: "Easily manage your project's versions",
}

var bumpCmd = &cobra.Command{
	Use:   "bump",
	Short: "Bumps the latest version to the next version and tags it",
	Run: func(cmd *cobra.Command, args []string) {
		repo, err := git_semver.OpenRepository("./")
		if err != nil {
			log.Fatal(err)
		}

		versions, err := git_semver.ListVersions(repo)
		if err != nil {
			log.Fatal(err)
		}

		versionsLen := len(versions) - 1

		var commits []*git.Commit
		var latestVersion *semver.Version

		if versionsLen >= 0 {
			latestVersion = versions[versionsLen]

			commits, err = git_semver.ListCommitsInRange(repo, latestVersion.String(), "HEAD")
			if err != nil {
				log.Fatal(err)
			}
		} else {
			latestVersion, err = semver.NewVersion("0.0.0")
			if err != nil {
				log.Fatal(err)
			}

			commits, err = git_semver.ListCommits(repo)
			if err != nil {
				log.Fatal(err)
			}
		}

		increment := git_semver.DetectIncrement(commits)

		if increment == git_semver.None {
			log.Fatal("No increment detected to bump the version")
		}

		nextVersion := git_semver.BumpVersion(*latestVersion, increment)

		if err := git_semver.TagVersion(repo, nextVersion); err != nil {
			log.Fatal(err)
		}

		if err := git_semver.PushVersionTagToRemotes(repo, nextVersion); err != nil {
			log.Fatal(err)
		}
	},
}

var nextCmd = &cobra.Command{
	Use:   "next",
	Short: "Outputs the next unreleased version",
	Run: func(cmd *cobra.Command, args []string) {
		repo, err := git_semver.OpenRepository("./")
		if err != nil {
			log.Fatal(err)
		}

		versions, err := git_semver.ListVersions(repo)
		if err != nil {
			log.Fatal(err)
		}

		versionsLen := len(versions) - 1

		var commits []*git.Commit
		var latestVersion *semver.Version

		if versionsLen >= 0 {
			latestVersion = versions[versionsLen]

			commits, err = git_semver.ListCommitsInRange(repo, latestVersion.String(), "HEAD")
			if err != nil {
				log.Fatal(err)
			}
		} else {
			latestVersion, err = semver.NewVersion("0.0.0")
			if err != nil {
				log.Fatal(err)
			}

			commits, err = git_semver.ListCommits(repo)
			if err != nil {
				log.Fatal(err)
			}
		}

		increment := git_semver.DetectIncrement(commits)
		nextVersion := git_semver.BumpVersion(*latestVersion, increment)

		fmt.Println(nextVersion.String())
	},
}

var latestCmd = &cobra.Command{
	Use:   "latest",
	Short: "Outputs the latest released version",
	Run: func(cmd *cobra.Command, args []string) {
		repo, err := git_semver.OpenRepository("./")
		if err != nil {
			log.Fatal(err)
		}

		versions, err := git_semver.ListVersions(repo)
		if err != nil {
			log.Fatal(err)
		}

		versionsLen := len(versions) - 1

		if versionsLen < 0 {
			log.Fatal("No released versions found")
		}

		fmt.Println(versions[versionsLen].String())
	},
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.AddCommand(bumpCmd)
	rootCmd.AddCommand(nextCmd)
	rootCmd.AddCommand(latestCmd)
}
