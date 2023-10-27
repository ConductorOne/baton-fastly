[Baton Logo](./docs/images/baton-logo.png)

# `baton-fastly` [![Go Reference](https://pkg.go.dev/badge/github.com/conductorone/baton-fastly.svg)](https://pkg.go.dev/github.com/conductorone/baton-fastly) ![main ci](https://github.com/conductorone/baton-fastly/actions/workflows/main.yaml/badge.svg)

`baton-fastly` is a connector for Baton built using the [Baton SDK](https://github.com/conductorone/baton-sdk). It works with Fastly API.

Check out [Baton](https://github.com/conductorone/baton) to learn more about the project in general.

# Prerequisites

Connector requires automation token that is used throughout the communication with API. To obtain this token, you have to create one in Fastly. More in information about how to generate token [here](https://developer.fastly.com/reference/api/auth-tokens/automation)). 

After you have obtained access token, you can use it with connector. You can do this by setting `BATON_ACCESS_TOKEN` or by passing `--access-token`.

# Getting Started

## brew

```
brew install conductorone/baton/baton conductorone/baton/baton-fastly
BATON_ACCESS_TOKEN=token baton-fastly
baton resources
```

## docker

```
docker run --rm -v $(pwd):/out -e BATON_ACCESS_TOKEN=token ghcr.io/conductorone/baton-fastly:latest -f "/out/sync.c1z"
docker run --rm -v $(pwd):/out ghcr.io/conductorone/baton:latest -f "/out/sync.c1z" resources
```

## source

```
go install github.com/conductorone/baton/cmd/baton@main
go install github.com/conductorone/baton-fastly/cmd/baton-fastly@main
BATON_ACCESS_TOKEN=token baton-fastly
baton resources
```

# Data Model

`baton-fastly` will fetch information about the following Baton resources:

- Users
- Roles
- Services

# Contributing, Support and Issues

We started Baton because we were tired of taking screenshots and manually building spreadsheets. We welcome contributions, and ideas, no matter how small -- our goal is to make identity and permissions sprawl less painful for everyone. If you have questions, problems, or ideas: Please open a Github Issue!

See [CONTRIBUTING.md](https://github.com/ConductorOne/baton/blob/main/CONTRIBUTING.md) for more details.

# `baton-fastly` Command Line Usage

```
baton-fastly

Usage:
  baton-fastly [flags]
  baton-fastly [command]

Available Commands:
  capabilities       Get connector capabilities
  completion         Generate the autocompletion script for the specified shell
  help               Help about any command

Flags:
      --access-token string    Fastly API token
      --client-id string       The client ID used to authenticate with ConductorOne ($BATON_CLIENT_ID)
      --client-secret string   The client secret used to authenticate with ConductorOne ($BATON_CLIENT_SECRET)
  -f, --file string            The path to the c1z file to sync with ($BATON_FILE) (default "sync.c1z")
  -h, --help                   help for baton-fastly
      --log-format string      The output format for logs: json, console ($BATON_LOG_FORMAT) (default "json")
      --log-level string       The log level: debug, info, warn, error ($BATON_LOG_LEVEL) (default "info")
  -p, --provisioning           This must be set in order for provisioning actions to be enabled. ($BATON_PROVISIONING)
  -v, --version                version for baton-fastly

Use "baton-fastly [command] --help" for more information about a command.
```