package gitsemver

import (
	"fmt"
	"log"
	"os"

	"github.com/Masterminds/semver"
	"github.com/carlsberg/git-semver/pkg/gitsemver"
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

		if err := project.Bump(versionFilenamesAndKeys); err != nil {
			log.Fatalln(err)
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

func newProjectOrPanic(cmd *cobra.Command) *gitsemver.Project {
	cwd, err := os.Getwd()
	if err != nil {
		log.Fatalln(err)
	}

	dir, err := cmd.Flags().GetString("project")
	if err != nil {
		log.Fatalln(err)
	}

	project, err := gitsemver.NewProject(cwd, dir)
	if err != nil {
		log.Fatalln(err)
	}

	return project
}

func getLatestVersionOrFail(project *gitsemver.Project) *semver.Version {
	latest, err := project.LatestVersion()
	if err != nil {
		log.Fatalln(err)
	}

	return latest
}

func getNextVersionOrFail(project *gitsemver.Project) *semver.Version {
	next, err := project.NextVersion()
	if err != nil {
		log.Fatalln(err)
	}

	return next
}

func getVersionFilenamesAndKeysOrFail(cmd *cobra.Command) []string {
	versionFilenamesAndKeys, err := cmd.Flags().GetStringArray("version-file")
	if err != nil {
		log.Fatalln(err)
	}

	return versionFilenamesAndKeys
}
