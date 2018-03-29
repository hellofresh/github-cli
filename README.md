# Github CLI

[![Build Status](https://travis-ci.org/hellofresh/github-cli.svg?branch=master)](https://travis-ci.org/hellofresh/github-cli)
[![Go Report Card](https://goreportcard.com/badge/github.com/hellofresh/github-cli)](https://goreportcard.com/report/github.com/hellofresh/github-cli)

> A CLI Tool to automate the creation of github repositories

This is a simple, CLI tool that helps you to create github repositories. 
It adds all required integrations, teams, webhooks, etc.. all based on a configuration file that you define.

## Installation

You can get the binary and play with it in your own environment (or even deploy it wherever you like it).
Just go the [releases](https://github.com/hellofresh/github-cli/releases) and download the latest one for your platform.

Just place the binary in your $PATH and you are good to go.

## Getting Started

After you have *github-cli* up and running we can create our first repository.
First of all we have to create a configuration file that will customize how our repositories will be created. You can have a look on our [example](.github.sample.toml).
This file can be placed on the same folder as your binary is, or in your home directory and it should be named `.github.toml` (You can also use it as `yaml` or `json`).

## Usage

```
github-cli [command] [--flags]
``` 

### Commands

| Command                  | Description                          |
|--------------------------|--------------------------------------|
| `github-cli repo create [--flags]`   | Creates a new github repository      |
| `github-cli repo delete [--flags]`   | Deletes a github repository          |
| `github-cli hiring send [--flags]`   | Creates a new hellofresh hiring test |
| `github-cli hiring unseat [--flags]` | Removes external collaborators from repositories |
| `github-cli update `                 | Check for new versions of github-cli |
| `github-cli version`                 | Prints the version information |

## Contributing

To start contributing, please check [CONTRIBUTING](CONTRIBUTING.md).

## Documentation

* Phanes Docs: https://godoc.org/github.com/hellofresh/github-cli
