# git-semver

[![Build](https://github.com/crqra/git-semver/actions/workflows/build.yml/badge.svg)](https://github.com/crqra/git-semver/actions/workflows/build.yml) [![LoC](https://tokei.rs/b1/github/crqra/git-semver)](https://github.com/crqra/git-semver).


> Git extension to easily manage your project's version based on [Semantic Versioning][semver] and [Conventional Commits][conventional-commits]

_work in progress_

## System Requirements

- `libgit2 >= 1.1`

## Commands

#### `git semver bump`

Bumps the latest version to the next version and tags it

#### `git semver latest`

Outputs the latest released version

#### `git semver next`

Outputs the next unreleased version

## License

This project is released under the [MIT License](LICENSE).

[conventional-commits]: https://www.conventionalcommits.org/en/v1.0.0/
[semver]: https://semver.org