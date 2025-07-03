# Installation

Choose the installation method that best fits your environment.

## Go Install (recommended)

```
go install github.com/kontrolplane/kue@latest
```

This will compile Kue and place the binary in your `$GOPATH/bin` (usually `~/go/bin`). Make sure this directory is on your `PATH`.

## Pre-built Binaries

_Pre-built binaries will be provided in future releases._ **TODO**

## Docker

```
docker run --rm -it ghcr.io/kontrolplane/kue:latest
```

## From Source

Clone the repository and build locally:

```bash
git clone https://github.com/kontrolplane/kue.git
cd kue
go build -o kue
./kue
```
