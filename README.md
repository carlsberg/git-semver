# git-semver

[![Build](https://github.com/carlsberg/git-semver/actions/workflows/build.yml/badge.svg)](https://github.com/carlsberg/git-semver/actions/workflows/build.yml) [![LoC](https://tokei.rs/b1/github/crqra/git-semver)](https://github.com/crqra/git-semver)

> Git extension to easily manage your project's version based on [Semantic Versioning][semver] and [Conventional Commits][conventional-commits]

## Install

### Using Brew

```shell
$ brew install carlsberg/tap/git-semver
```

### From source

Requires `go >= 1.16`.

```shell
$ go install github.com/carlsberg/git-semver@latest
```

## Commands

#### `git semver bump`

Bumps the latest version to the next version, creates a tag and pushes it to `origin`.

**Options:**

| Name                 | Description                                                               |
| -------------------- | ------------------------------------------------------------------------- |
| `-P, --password`     | Password to use in HTTP basic authentication (useful for CI/CD pipelines) |
| `-u, --username`     | Username to use in HTTP basic authentication (useful for CI/CD pipelines) |
| `-f, --version-file` | Version files that should be updated. Format should be `filename:key`     |
| `--v-prefix`         | Prefix the version with a `v`                                             |
| `--skip-tag`         | Don't create a new tag automatically                                      |

**Example:**

```shell
$ git commit -am "fix: something important"
$ git semver bump -f package.json:version
bump from 0.0.0 to 0.0.1
```

#### `git semver latest`

Outputs the latest released version.

**Example:**

```shell
$ git semver latest
0.0.1
```

#### `git semver next`

Outputs the next unreleased version.

**Options:**

| Name                 | Description                                                               |
| -------------------- | ------------------------------------------------------------------------- |
| `--v-prefix`         | Prefix the version with a `v`                                             |

**Example:**

```shell
$ git commit -am "feat: neat feat"
$ git semver next
0.1.0
```

## Monorepos Support

Git SemVer supports monorepos with multiple projects through the `-p, --project` option. For example, imagine you've a repository with a project in `libs/xyz` that you want to version independently:

```shell
$ git commit -am "feat: awesome xyz lib"
$ git bump -p libs/xyz -f package.json:version
bump libs/xyz from 0.0.0 to 0.1.0
# creates a tag `libs/xyz/0.1.0`
```

## Usage in GitHub Actions

The following example shows a workflow using git-semver in GitHub Actions:

```yaml
name: Bump

jobs:
  bump:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v2
        with:
          fetch-depth: 0

      - uses: carlsberg/git-semver@v0.5.0
        with:
          script: |
            echo "Latest Version: $(git semver latest)"
            echo "Next Version: $(git semver next)"
            
            git semver bump -u ${{ github.actor }} -P ${{ github.token }}
```

## License

This project is released under the [MIT License](LICENSE).

[conventional-commits]: https://www.conventionalcommits.org/en/v1.0.0/
[semver]: https://semver.org
