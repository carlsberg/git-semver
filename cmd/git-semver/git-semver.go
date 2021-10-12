package git_semver

import (
	"fmt"
	"log"
	"os"

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
		cwd, err := os.Getwd()
		if err != nil {
			log.Println(err)
		}

		dir, err := cmd.Flags().GetString("project")
		if err != nil {
			log.Fatal(err)
		}

		p, err := git_semver.NewProject(cwd, dir)
		if err != nil {
			log.Fatal(err)
		}

		versionFilenamesAndKeys, err := cmd.Flags().GetStringArray("version-file")
		if err != nil {
			log.Fatal(err)
		}

		latest, err := p.LatestVersion()
		if err != nil {
			log.Fatal(err)
		}

		next, err := p.NextVersion()
		if err != nil {
			log.Fatal(err)
		}

		if len(versionFilenamesAndKeys) > 0 {
			for _, filenameAndKey := range versionFilenamesAndKeys {
				vf, err := git_semver.NewVersionFile("./", filenameAndKey)
				if err != nil {
					log.Fatal(err)
				}

				vf.UpdateVersion(latest, next)
				if err != nil {
					log.Fatal(err)
				}
			}

			err := git.CreateCommit(p.Repo(), fmt.Sprintf("bump: %s -> %s", latest.String(), next.String()))
			if err != nil {
				log.Fatal(err)
			}
		}

		tagName := git_semver.TagNameFromProjectAndVersion(p, next)
		tagMessage := fmt.Sprintf("Release %s", tagName)

		if err := git.CreateTag(p.Repo(), tagName, tagMessage); err != nil {
			log.Fatal(err)
		}

		if err := git.PushTagToRemotes(p.Repo(), tagName); err != nil {
			log.Fatal(err)
		}

		fmt.Printf("bump %s from %s to %s\n", p.Dir(), latest, next)
	},
}

var nextCmd = &cobra.Command{
	Use:   "next",
	Short: "Outputs the next unreleased version",
	Run: func(cmd *cobra.Command, args []string) {
		cwd, err := os.Getwd()
		if err != nil {
			log.Println(err)
		}

		dir, err := cmd.Flags().GetString("project")
		if err != nil {
			log.Fatal(err)
		}

		p, err := git_semver.NewProject(cwd, dir)
		if err != nil {
			log.Fatal(err)
		}

		next, err := p.NextVersion()
		if err != nil {
			log.Fatal(err)
		}

		fmt.Println(next.String())
	},
}

var latestCmd = &cobra.Command{
	Use:   "latest",
	Short: "Outputs the latest released version",
	Run: func(cmd *cobra.Command, args []string) {
		cwd, err := os.Getwd()
		if err != nil {
			log.Println(err)
		}

		dir, err := cmd.Flags().GetString("project")
		if err != nil {
			log.Fatal(err)
		}

		p, err := git_semver.NewProject(cwd, dir)
		if err != nil {
			log.Fatal(err)
		}

		latest, err := p.LatestVersion()
		if err != nil {
			log.Fatal(err)
		}

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
