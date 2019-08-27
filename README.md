# go-miab

go-miab is a simple command-line tool, designed to manage a [Mail-in-a-Box](https://mailinabox.email/), written in [Go](https://golang.org/).

[![Go Report Card](https://goreportcard.com/badge/github.com/rverst/go-miab)](https://goreportcard.com/report/github.com/rverst/go-miab)

## Overview

go-miab contains the package `miab` that wraps the API of a Mail-in-a-Box instance. Additionally there is an command-line interface (CLI) to access the Mail-in-a-Box API directly. Most of the endpoints, that the Mail-in-a-Box API provides should be covered (Mail-in-a-Box v0.42b).

* Query custom DNS records
* Add or set (overwrite) custom DNS records
* Delete custom DNS records
* Query e-mail users
* Create e-mail users and mail-domains
* Delete e-mail users
* Set e-mail users privileges
* Query e-mail aliases
* Create e-mail aliases
* Delete e-mail aliases

## Installation

### Binary Release

Download a binary release from the [release page](github.com/rverst/go-miab/releases).

### Go

```bash
go get github.com/rverst/go-miab

// to install the cli
cd $GOPATH/src/github.com/rverst/go-miab
go install ./cmd/cli/miab.go
```

## Usage

First you have to provide the credentials and endpoint for your Mail-in-a-Box instance.
There are several ways to do so:

1. Command-line

    ```bash
    miab --user admin@example.org --password secretpassword --endpoint https://box.example.org
    or
    miab -u admin@example.org -p secretpassword -e https://box.example.org
    ```

2. Environment variables

    ```bash
    MIAB_USER=admin@example.org \
    export MIAB_USER \
    MIAB_PASSWORD=secretpassword \
    export MIAB_PASSWORD \
    MIAB_ENDPOINT=https://box.example.org \
    export MIAB_ENDPOINT
    ```

3. Config file

    ```yaml
    user: admin@example.org
    password: supersectet
    endpoint: https://box.example.org
    ```

    The default location for the config file is `$HOME/.config/go-miab/miab.yaml`.
    The location can be specified via the `config` flag.  
  
**Run `miab help` for available commands.**

## Dependencies

go-miab uses and relies on the following, awesome libraries (in lexical order):

| Dependency | License |
| :------------- | :------------- |
| [github.com/mitchellh/go-homedir](https://github.com/mitchellh/go-homedir) | [MIT License](https://github.com/mitchellh/go-homedir/blob/master/LICENSE) |
| [github.com/spf13/cobra](https://github.com/spf13/cobra) | [Apache License 2.0](https://github.com/spf13/cobra/blob/master/LICENSE.txt) |
| [github.com/spf13/pflag](https://github.com/spf13/pflag) | [BSD 3-Clause "New" or "Revised" License](https://github.com/spf13/pflag/blob/master/LICENSE) |
| [github.com/spf13/viper](https://github.com/spf13/viper) | [MIT License](https://github.com/spf13/viper/blob/master/LICENSE) |
| [gopkg.in/yaml.v2](https://gopkg.in/yaml.v2) | [Apache License 2.0](https://github.com/go-yaml/yaml/blob/v2/LICENSE) |
