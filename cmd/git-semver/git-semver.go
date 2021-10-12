package git_semver

import (
	"fmt"
	"log"
	"os"

	"github.com/Masterminds/semver"
	"github.com/carlsberg/git-semver/internal/git"
	git_semver "github.com/carlsberg/git-semver/internal/git-semver"
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
		project := newProjectOrPanic(cmd)
		versionFilenamesAndKeys := getVersionFilenamesAndKeysOrFail(cmd)
		latest := getLatestVersionOrFail(project)
		next := getNextVersionOrFail(project)

		if len(versionFilenamesAndKeys) > 0 {
			for _, filenameAndKey := range versionFilenamesAndKeys {
				updateVersionFileOrPanic(filenameAndKey, latest, next)
			}

			if err := git.CreateCommit(project.Repo(), fmt.Sprintf("bump: %s -> %s", latest.String(), next.String())); err != nil {
				log.Fatal(err)
			}
		}

		tagName := git_semver.TagNameFromProjectAndVersion(project, next)
		tagMessage := fmt.Sprintf("Release %s", tagName)

		if err := git.CreateTag(project.Repo(), tagName, tagMessage); err != nil {
			log.Fatal(err)
		}

		if err := git.PushTagToRemotes(project.Repo(), tagName); err != nil {
			log.Fatal(err)
		}

		fmt.Printf("bump %s from %s to %s\n", project.Dir(), latest, next)
	},
}

var nextCmd = &cobra.Command{
	Use:   "next",
	Short: "Outputs the next unreleased version",
	Run: func(cmd *cobra.Command, args []string) {
		project := newProjectOrPanic(cmd)
		next := getNextVersionOrFail(project)

		fmt.Println(next.String())
	},
}

var latestCmd = &cobra.Command{
	Use:   "latest",
	Short: "Outputs the latest released version",
	Run: func(cmd *cobra.Command, args []string) {
		project := newProjectOrPanic(cmd)
		latest := getLatestVersionOrFail(project)

		fmt.Println(latest.String())
	},
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.AddCommand(bumpCmd)
	rootCmd.AddCommand(nextCmd)
	rootCmd.AddCommand(latestCmd)
	rootCmd.PersistentFlags().StringP("project", "p", "", "Project")

	bumpCmd.Flags().StringArrayP("version-file", "f", make([]string, 0), "Specify version files to be updated with the new version in the format `filename:key` (i.e. `package.json:\"version\"`)")
}

func newProjectOrPanic(cmd *cobra.Command) *git_semver.Project {
	cwd, err := os.Getwd()
	if err != nil {
		log.Println(err)
	}

	dir, err := cmd.Flags().GetString("project")
	if err != nil {
		log.Fatal(err)
	}

	project, err := git_semver.NewProject(cwd, dir)
	if err != nil {
		log.Fatal(err)
	}

	return project
}

func updateVersionFileOrPanic(filenameAndKey string, latest, next *semver.Version) {
	vf, err := git_semver.NewVersionFile("./", filenameAndKey)
	if err != nil {
		log.Fatal(err)
	}

	vf.UpdateVersion(latest, next)
	if err != nil {
		log.Fatal(err)
	}
}

func getLatestVersionOrFail(project *git_semver.Project) *semver.Version {
	latest, err := project.LatestVersion()
	if err != nil {
		log.Fatal(err)
	}

	return latest
}

func getNextVersionOrFail(project *git_semver.Project) *semver.Version {
	next, err := project.NextVersion()
	if err != nil {
		log.Fatal(err)
	}

	return next
}

func getVersionFilenamesAndKeysOrFail(cmd *cobra.Command) []string {
	versionFilenamesAndKeys, err := cmd.Flags().GetStringArray("version-file")
	if err != nil {
		log.Fatal(err)
	}

	return versionFilenamesAndKeys
}
