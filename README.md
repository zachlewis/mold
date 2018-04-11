# mold [![Build Status](https://travis-ci.org/d3sw/mold.svg?branch=master)](https://travis-ci.org/d3sw/mold)
Mold is a tool to help test, build, package and publish your application completely within a containerized environment.
It automates the process from installing dependencies and testing to packaging and publishing your image to a registry.

Mold starts by creating an isolated network to run your build, followed by installing your dependencies, running
unit tests and building any needed binaries in a container.  These binaries are in turn used to package the
image and publish to a registry.

Mold also helps manage versioning by leveraging git and using tags as points of reference to automate
version computation and appropriately tagging images.

## Installation
[Download](https://github.com/d3sw/mold/releases) the binary based on your OS.  Once uncompressed copy it into your system PATH.

## Usage
To use mold you can simply issue the `mold` command in the root of your git
repository.  By default the command looks for a `.mold.yml` [configuration](#Configuration) file at
the root of your project.  All available options can be seen by issuing:
```
mold --help
```

## Windows Usage
On Windows 10, the following needs to be performed in order for mold to function properly

1. Must specify `-uri tcp://127.0.0.1:2375` options.
2. Set the home environment variable to `HOME=C:/Users/{username}`
3. Make sure `$HOME/.docker/config.json` exists.  You can run `docker login` to create one or simply create any empty file.
4. Make sure C: is shared: Docker settings > Shared Drives > Select the local drives you want to be available to your containers
5. Configure Docker by checking `Expose daemon on tcp://localhost:2375`

It is NOT recommended to run mold on Windows 7 since Docker Engine cannot run natively on it. Just when there is no other options, Docker toolbox could be installed and Docker Engine so as mold can run inside a Linux VM hosted on the Windows 7 system.

(Ref: https://docs.docker.com/toolbox/toolbox_install_windows/)

1. Check the OS version - "Run a tool like the Microsoft® Hardware-Assisted Virtualization Detection Tool or Speccy, and follow the on-screen instructions."
2. Install Remote Server Administration Tools (RSAT) (of Windows)
3. Install Hyper-V tools (of Windows)
4. Install Docker toolbox
5. Disable TLS in Docker profile
  a. Start "Docker Quickstart Terminal
  b. Start ssh: docker-machine ssh
  c. Edit /var/lib/boot2docker/profile so DOCKER_HOST='-H tcp://{ip}:2375'and DOCKER_TLS=no
  d. Exit ssh: exit
  e. Restart Docker machine: docker-machine restart
6. Mount mold realeased for Linux to the VM
7. Mount project files to the VM
8. Start `docker-machine ssh` and run mold and Docker in it

## Configuration
The mold process is controlled by a single configuration file which by default is `.mold.yml`.  For detailed
information on configuration options please visit the [configuration](docs/Configuration.md) page.

## Cleanup
As you perform builds, there will be a build of containers and images left behind that may no
longer be needed.  You can pick and choose which ones to keep.  A helper script has been provided
which removes all containers that have exited, intermediate images as well as dangling volumes.

DO NOT USE this script if any of the exited containers, images or volumes are of any value that
you would like to save.

The script can be found in [scripts/drclean](scripts/drclean).  Please read the comments if you would like to know
what it exactly does.

## FAQ

### 1. Why not use Docker Compose to test, build, package, and publish our applications?

[Docker Compose](https://docs.docker.com/compose/overview/) is a tool to define and run multi-container applications, optionally
assembling needed images before running the application stack.  Mold is used manage your CI pipeline in Docker i.e test, build,
package and publish.

One still may wonder if mold is really needed and if the same could be acheived via docker-compose. Based on our tests, it does
seem viable to use docker-compose as a CI solution.

Docker compose controls the order of service startup but does not provide way to manage the order of image builds. Below shows the
docker-compose file and the Dockerfile we used to mimic the build process and test if the dependency conditions would delay the
image build from the Dockerfile until the application is built.

#### docker-compose.yml

```
version: '2.1'
services:
  build_img:
    build: .
    depends_on:
      build_app:
        condition: service_healthy
  build_app:
    image: alpine
    volumes:
      - ./:/app/
    command: /bin/sh -ec "sleep 5s; head -c 10 /dev/urandom > /app/myApp"
    healthcheck:
        test: ["CMD", "/bin/sh", "-f", "/app/fileExist.sh", "/app/myApp"]
        interval: 30s
        timeout: 10s
        retries: 5
```

#### Dockerfile

```
FROM alpine
ADD . /app
CMD ["/bin/sh", "-f", "/app/fileExist.sh", "/app/myApp"]
```

The result below shows the dependency condition only affects the order of the service startup and not that of an image build:

```
Building build_img
...
Successfully built 2946ba878f8c
...
Creating composetest_build_app_1
Creating composetest_build_img_1
...
build_img_1  | /app/myApp does NOT exist
composetest_build_app_1 exited with code 0
composetest_build_img_1 exited with code 1
```

Another aspect that mold handles is separating the build images from the deployment images i.e builds happen in 1 container
and the resulting artifacts are then used to build the final image which will be published to a registry.

### 2. What are the system requirements to run mold?

Mold is [released for Linux, Mac, and Windows](https://github.com/d3sw/mold/releases). It however requires Docker installed on the system. This also means for Windows system, it requires 64bit Windows 10 Pro, Enterprise and Education (1511 November update, Build 10586 or later) and Microsoft Hyper-V. Please see the details from the [Docker site](https://docs.docker.com/docker-for-windows/).

### 3. Where should I run mold? Should it be triggered on the CI server or should I run it locally?

You can run mold locally or be incorporate it into your CI pipeline and run on you CI server.

## Roadmap
- 0.2.6
    - Build image optimization and caching.
- 0.2.5
    - First official open-source release.

## Known Issues

### Output Delay

This issue appears due to Docker API limits.
At times the command runs and completes execution before the output and status can even be obtained.
This is particularly the case when a command is not found or a single command execution completes quickly enough. i.e. before the call to get the status and output is made.

To avoid this a sleep statement can be added as the first command in the build process. Example:

    commands:
    - sleep 1
    - mvn test

In the case where mvn exists the sleep is not required.
