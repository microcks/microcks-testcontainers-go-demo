# Step 1: Getting Started

Before getting started, let's make sure you have everything you need for running this demo.

## Prerequisites

### Install Go 1.23 or newer

You'll need Go 1.23 or newer for this workshop.

### Install Docker

You need to have a [Docker](https://docs.docker.com/get-docker/) or [Podman](https://podman.io/) environment to use Testcontainers.

```shell
$ docker version

Client:
 Cloud integration: v1.0.35+desktop.10
 Version:           25.0.3
 API version:       1.44
 Go version:        go1.21.6
 Git commit:        4debf41
 Built:             Tue Feb  6 21:13:26 2024
 OS/Arch:           darwin/arm64
 Context:           desktop-linux

Server: Docker Desktop 4.27.2 (137060)
 Engine:
  Version:          25.0.3
  API version:      1.44 (minimum version 1.24)
  Go version:       go1.21.6
  Git commit:       f417435e5f6216828dec57958c490c4f8bae4f98
  Built:            Wed Feb  7 00:39:16 2024
  OS/Arch:          linux/arm64
  Experimental:     false
```

## Download the project

Clone the [microcks-testcontainers-go-demo](https://github.com/microcks/microcks-testcontainers-go-demo) repository from GitHub to your computer:

```shell
git clone https://github.com/microcks/microcks-testcontainers-go-demo.git
```

## Compile the project to download the dependencies

With the Makefile:

```shell
make build
```

### 

[Next](step-2-exploring-the-app.md)