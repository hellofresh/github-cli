# Github CLI

[![Build Status](https://travis-ci.org/hellofresh/github-cli.svg?branch=master)](https://travis-ci.org/hellofresh/github-cli)
[![Go Report Card](https://goreportcard.com/badge/github.com/hellofresh/github-cli)](https://goreportcard.com/report/github.com/hellofresh/github-cli)

> A CLI Tool to automate the creation of github repositories

This is a simple, CLI tool that helps you to create github repositories.
It adds all required integrations, teams, webhooks, etc.. all based on a configuration file that you define.

## Installation

You can get the binary and play with it in your own environment (or even deploy it wherever you like it).
Just go the [releases](https://github.com/hellofresh/github-cli/releases) and download the latest one for your platform.

Move the downloaded binary to any place that's in your `$PATH`, usually you would add it to the local bin:

```
$ mv ~/Downloads/github-cli /usr/local/bin/github-cli
$ chmod +x /usr/local/bin/github-cli
```

## Getting Started

After you have _github-cli_ up and running we can create our first repository.
First of all we have to create a configuration file that will customise how our repositories will be created. You can have a look at our [example](./.github.sample.toml) and copy it.

You will need to fill in the following values to be able to create a repo:

### GitHub `github`

GitHub needs a [token](https://github.com/settings/tokens/new) with repo access. You can add additional collaborators and/or teams, the teams are defined by ID and the ID can be found with this simple cURL:

```
$ curl -s -i -X GET -u TOKEN:x-oauth-basic -d '' https://api.github.com/orgs/hellofresh/teams | grep -A1 "TEAM_NAME"
```

Check out descriptions on the other config values in the [sample file](./.github.sample.toml).

### GitHub Test Org `githubtestorg`

This is used for creating GitHub tests. This just needs a GitHub token with repo access.

### Finalising the file

You can either write this file to be next to the binary or add it to `~/.github.toml`, we recommend the latter!

## Usage

```
github-cli [command] [--flags]
```

### Commands

| Command                              | Description                                      |
| ------------------------------------ | ------------------------------------------------ |
| `github-cli repo create [--flags]`   | Creates a new github repository                  |
| `github-cli repo delete [--flags]`   | Deletes a github repository                      |
| `github-cli hiring send [--flags]`   | Creates a new hellofresh hiring test             |
| `github-cli hiring unseat [--flags]` | Removes external collaborators from repositories |
| `github-cli update`                  | Check for new versions of github-cli             |
| `github-cli version`                 | Prints the version information                   |

## Contributing

To start contributing, please check [CONTRIBUTING](CONTRIBUTING.md).

## Documentation

- Phanes Docs: https://godoc.org/github.com/hellofresh/github-cli
