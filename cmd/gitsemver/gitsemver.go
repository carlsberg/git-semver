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
		vPrefix := getVPrefixOrFail(cmd)
		skipTag := getSkipTagOrFail(cmd)
		username := getUsernameOrFail(cmd)
		password := getPasswordOrFail(cmd)
		latest := getLatestVersionOrFail(project)
		next := getNextVersionOrFail(project, vPrefix)

		var auth gitsemver.AuthMethod = nil

		if username != "" && password != "" {
			auth = &gitsemver.BasicAuth{
				Username: username,
				Password: password,
			}
		}

		if err := project.Bump(versionFilenamesAndKeys, auth, vPrefix, skipTag); err != nil {
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
		vPrefix := getVPrefixOrFail(cmd)
		next := getNextVersionOrFail(project, vPrefix)

		fmt.Println(next.Original())
	},
}

var latestCmd = &cobra.Command{
	Use:   "latest",
	Short: "Outputs the latest released version",
	Run: func(cmd *cobra.Command, args []string) {
		project := newProjectOrPanic(cmd)
		latest := getLatestVersionOrFail(project)

		fmt.Println(latest.Original())
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
	bumpCmd.Flags().StringP("username", "u", "", "Username to use in HTTP basic authentication")
	bumpCmd.Flags().StringP("password", "P", "", "Password to use in HTTP basic authentication")
	bumpCmd.Flags().Bool("v-prefix", false, "Prefix the version with a `v`")
	bumpCmd.Flags().Bool("skip-tag", false, "Don't create a new tag automatically")

	nextCmd.Flags().Bool("v-prefix", false, "Prefix the version with a `v`")
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

func getNextVersionOrFail(project *gitsemver.Project, vPrefix bool) *semver.Version {
	next, err := project.NextVersion(vPrefix)
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

func getUsernameOrFail(cmd *cobra.Command) string {
	username, err := cmd.Flags().GetString("username")
	if err != nil {
		log.Fatalln(err)
	}

	return username
}

func getPasswordOrFail(cmd *cobra.Command) string {
	password, err := cmd.Flags().GetString("password")
	if err != nil {
		log.Fatalln(err)
	}

	return password
}

func getVPrefixOrFail(cmd *cobra.Command) bool {
	vPrefix, err := cmd.Flags().GetBool("v-prefix")
	if err != nil {
		log.Fatalln(err)
	}

	return vPrefix
}

func getSkipTagOrFail(cmd *cobra.Command) bool {
	skipTag, err := cmd.Flags().GetBool("skip-tag")
	if err != nil {
		log.Fatalln(err)
	}

	return skipTag
}
