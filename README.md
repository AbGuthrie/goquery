# goquery

## Introduction

goquery is a remote investigation client that uses your existing osquery deployment to provide remote shell level functionality with fewer risks than SSH.

Using osquery's distributed API, hosts can be targeted for single queries to pull back specific information without having to modify the osquery schedule. goquery uses this API and abstracts it into a shell like experience that investigators are more used to while retaining all the power of osquery tables, extensions, and features like Auto Table Construction. Hosts can be connected to by their UUID and goquery creates an interactive session with the host over the asynchronous distributed API. The concept of node keys, check in times, JSON escaping, discovery queries, and the osquery schedule is abstracted away (or not used) to provide a clean, efficient way to remotely interrogate hosts for abuse, compromise investigation, or fleet management.

## Commands

### .clear
Clear the terminal screen

### .connect \<UUID\>
This opens a session with a remote host. It will ask the backend if a host with that UUID is registered and if not return to the user saying it doesn't exist. If the backend returns that the host exists then a session is opened and that machine is set as the active host. All future commands will interact with this host until it's disconnected from or the user changes to another host. Supports suggestions.

### .disconnect \<UUID\>
Close a session with a remote host. Fails if you're not connected to a host with that UUID. Supports suggestions.

### .exit
Exit goquery. Shell state will not be saved but command history is.

### .help
Show goquery help formatted with the currently selected printing mode.

### .history
Show all past queries in the current session for the current host.

### .hosts
Show all hosts you are connected to with their osquery version, hostname, UUID, and platform

### .mode \<PrintMode\>
Change the printing mode. goquery supports multiple printing modes to help you make sense of data at a glance. We currently support: Line, JSON, and Pretty (default).

### .query \<query\>
Like .schedule and .resume together. Runs a query on a remote host and waits for the result before returning control to the REPL.

### .resume \<queryName\>
This will either wait for a query to complete or fetch the results and display them if the query has already posted results. This is used in conjunction with .schedule to pull the results of queries that are running asynchronously. This can also be used to display the results of any previously run query.

### .schedule \<query\>
Run a query asyncronously on the remote host. The query will be tracked in the session for that host so results can be fetched at any point in time, but this allows the investigator to kick off a bunch of things without waiting for each one to complete first.

### .alias \<aliasName\> \<command\> \<interpolated args\>
List current aliases when called with no arguments or flags. To create a new alias, call with `--add` flag and provide arguments as follows:  `.alais --add ALIAS_NAME command_string`

Positional arguments with $# placeholders are interpolated when the command is run, for example the following alias `.all` with command `.query select * from $#` will evaluate to `.query select * from processes` when called with `.all processes`.

Command name must not contain any spaces in order to preserve the space delimmitted arguments

To remove an alias, use `.alias --remove ALIAS_NAME`

### cd \<DIR\>
Change directories on a remote host. This affects other pseudo-commands like `ls`.

### ls
List the files in the current directory. The current directory is set by using the `cd` command and starts at `/`.

## API

To support the various features of goquery, your backend will need to support various APIs used to interact with your fleet. Not all APIs are needed but there are no redundent APIs (goquery can work without all APIs but will have its functionality diminished).

## Core API

### checkHost
Verify a host exists in the fleet.

### scheduleQuery
Schedule a query on a remote machine.

### fetchResults
Pull the results of a query by the name returned from scheduleQuery

## Hunting API

Placeholder

## ATC API

Placeholder

## Config

### Structure

Goquery can be configured via a configuration json file. Debug mode, defaults, and aliases can be set in the structure of the provided `config.template.json`. Valid print modes are as follows:
- "json"
- "line"
- "pretty"


By default, goquery will check for a config file at the following path: `~/.goquery/config.json`

This can be overidden when calling the binary or running with the following flags: `--config ./path_to_file.json`

## Building

### Docker Testing Infra
Hopefully one day goquery will be plug'n'play with the most popular osquery backends, but for now it'll take a little work to integrate. To get up and running playing with goquery as quickly as possible, you can use the docker test infra.

Running `make docker` will build a set of nodes used to create a simulated osquery deployment with two Ubuntu hosts, a central osquery server, along with a SAML IdP. goquery contains its own osquery server written in Go which is designed to be lightweight and easy to understand to help you learn how to integrate goquery into your enterprise.

Deploy it locally with `make deploy` (which uses docker swarm) and then you're ready to start testing by running goquery.

[Install go via asdf install .toolversions](https://asdf-vm.com/#/core-manage-asdf-vm)

Add go plugin: `asdf plugin-add golang https://github.com/kennyp/asdf-golang.git`

and finally `asdf install`

