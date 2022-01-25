# Demo code for Zartis webinar on Temporal

## Required software

In order to run this monorepo you'll need the following software to be installed in your machine:

* Make: Not mandatory but recommended for fiddling in the command line
* Golang 1.16

> Everything has been tested under Ubuntu 20.1

## Quickstart

Use the following commands to perform common actions:

* `make build`: Builds the binaries
* `docker-compose -f docker-devenv/docker-compose.yaml up -d`: Run the required docker containers for everything to work
* `./bin/worker`: Run workflow worker
* `./bin/dummyserver`: Run the dummy server

Web interface is available at: [http://localhost:8099](http://localhost:8099)
