# goquery

## Introduction

goquery is a remote investigation tool that uses your existing osquery deployment to provide remote shell level functionality with fewer risks than SSH.

Using osquery's distributed API, hosts can be targeted for single queries to pull back specific information without having to modify the osquery schedule. goquery uses this API and abstracts it into a shell like experience that investigators are more used to while retaining all the power of osquery tables, extensions, and features like Auto Table Construction. Hosts can be connected to by their UUID and goquery creates an interactive session with the host over the asynchronous distributed API. The concept of node keys, check in times, JSON escaping, discovery queries, and the osquery schedule is abstracted away (or not used) to provide a clean, efficient way to remotely interrogate hosts for abuse, compromise investigation, or fleet management.

## Commands

### .connect <UUID>
This opens a session with a remote host. It will ask the backend if a host with that UUID is registered and if not return to the user saying it doesn't exist. If the backend returns that the host exists then a session is opened and that machine is set as the active host. All future commands will interact with this host until it's disconnected from or the user changes to another host.

### .schedule <query>
Run a query asyncronously on the remote host. The query will be tracked in the session for that host so results can be fetched at any point in time, but this allows the investigator to kick off a bunch of things without waiting for each one to complete first.

### .resume <queryName>
This will either wait for a query to complete or fetch the results and display them if the query has already posted results. This is used in conjunction with .schedule to pull the results of queries that are running asynchronously. This can also be used to display the results of any previously run query.

### .query <query>
Like .schedule and .resume together. Runs a query on a remote host and waits for the result before returning control to the REPL.

### .exit
Exit goquery. Shell state will not be saved but command history is.


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

## ATC API


## Building

[Install go via asdf install .toolversions](https://asdf-vm.com/#/core-manage-asdf-vm)

Add go plugin: `asdf plugin-add golang https://github.com/kennyp/asdf-golang.git`

and finally `asdf install`
