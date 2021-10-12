package git_semver

import (
	"fmt"
	"log"
	"strings"

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

		versionFiles := make([]git_semver.VersionFile, 0)
		versionFilenamesAndKeys, err := cmd.Flags().GetStringArray("version-file")
		if err != nil {
			log.Fatal(err)
		}

		for _, filenameAndKey := range versionFilenamesAndKeys {
			slice := strings.Split(filenameAndKey, ":")

			if len(slice) != 2 {
				log.Fatalf("%s is not correctly formatted. Should be `filename:key`", filenameAndKey)
			}

			versionFiles = append(versionFiles, git_semver.VersionFile{Filename: slice[0], Key: slice[1]})
		}

		if err := git_semver.UpdateVersionFiles(repo, versionFiles, *latestVersion, nextVersion); err != nil {
			log.Fatal(err)
		}

		versionPrefix, err := cmd.Flags().GetString("version-prefix")
		if err != nil {
			log.Fatal(err)
		}

		if err := git_semver.TagVersion(repo, nextVersion, versionPrefix); err != nil {
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

	bumpCmd.Flags().StringArrayP("version-file", "f", make([]string, 0), "Specify version files to be updated with the new version in the format `filename:key` (i.e. `package.json:\"version\"`)")
	bumpCmd.Flags().StringP("version-prefix", "p", "", "A prefix for the version's tag name")
}
