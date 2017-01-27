# mold
Test, Build, Package and Publish your application completely using docker.

## Installation
[Download](https://github.com/d3sw/mold/releases) the binary based on your OS.  Once uncompressed copy it into your system PATH.

## Usage
To use mold you can simply issue the `mold` command in the root of your git
repository.  By default the command looks for a `.mold.yml` at the root of your project.  To
specify an alternate file you can use the `-f` flag followed by the path to your build config.

    Usage of mold:

      -f       string  Build config file (default ".mold.yml")
      -n               Enable notifications
      -t       string  Build a specific target only [build|artifact|publish]
      -uri     string  Docker URI (default "unix:///var/run/docker.sock")
      -version         Show version
      -var     string  Show value of vairable specified in the configuration file  (default: NA)

In most cases you will simply issue the `mold` command.

## Configuration
By default mold looks for a .mold.yml configuration file at the root of your project.
This contains all the necessary information to perform your build.  A sample with comments
can be found in [testdata/mold1.yml](testdata/mold1.yml).

The build configuration is broken up into the following sections:

- Services
- Build
- Artifacts/Publish

This also is representative of the lifecycle the build follows.  Each of the above
happen in sequential order.

All sections aside from `build` are optional.

### Example:
This example contains all supported options.  The `services` and `build` definitions are
identical.  Multiple services and builds can be defined for each of these sections.

    # Launch services needed for the build
    services:
        - image: elasticsearch
        - image: progrium/consul
          commands:
              - -server
              - -bootstrap

    # Perform 1 or more builds
    build:
        - image: golang:1.7.3
          workdir: /go/src/github.com/euforia/mold
          environment:
              - TEST_ENV=test_env
          commands:
              - hostname
              - uname -a
              - make

    # Build docker images
    artifacts:
        # Only publish the image on the following branches/tags. * can be used to
        # publish on all branches/tags
        publish:
            - master
        # Default registry to use if not specified. Blank uses docker hub
        registry: test.docker.registry
        images:
            - name: euforia/mold-test
              dockerfile: testdata/Dockerfile
              registry:

## Services
Services is a list of containers that need to be started prior to the build.  These are
containers your build process needs to perform the build.  For example if you are running
tests that require elasticsearch, you would declare a elasticsearch container to run
in this section as shown in the above example.

Service containers are spun up prior to the code build.  They are accessed via their image name
from your build container.

#### image
The image name of the service to start.  A vast list of public images can be found on
[Docker Hub](https://hub.docker.com).  Private images can also be specified.

#### commands
These are the arguments passed to the service container.

## Build
Build contains a list of builds to perform. This is used to perform testing and/or building binaries.  
Each build will run its set of provided commands in the specified container.  Any failed
command will cause the build to fail.  This is the only required configuration needed
to run the build.

#### workdir
This is **path inside the container** where the project repository will be accessible (mounted).
It can be an path of your choosing.  In the above example the source repo for mold will
be available under `/go/src/github.com/euforia/mold` inside the `golang:1.7.3` container.

#### image
This is the docker image name used to build/test code.  These are disposable and not used to generate the final
artifact.  Code is built using this image and the generated binaries or files are then used to
package the image as specified in the [artifacts](#Artifacts) configuration.

#### commands
These are the commands that will be run in the container to do testing and building.

## Artifacts
Artifacts are docker images to be built using the data available from the build step.  
Using the specified Dockerfile and name, image are built which may be published to a registry
based on conditional parameters.  This is the final product destined for production.  
These images would be very trimmed down and as minimalistic as possible specifically tailored
for the application.  

#### registry
This option sets the default registry for all images in the case where it is not supplied.
It defaults to [Docker Hub](https://hub.docker.com) if not specified.

#### publish
This option specifies which branches will trigger a push to the registry.  Available options
are:

- `*` For all branches/tags
- Name of a branch/tag

#### images
A list of images to build.  Each image has the following options available:

- **name**: Specifies the name of the image (required)

- **dockerfile**: Relative path to the Dockerfile. (required) Information on how a
Dockerfile works can be found [here](https://docs.docker.com/engine/reference/builder/)

- **registry**: Registry to push to.  If not specified the default one is used.

## Cleanup
As you perform builds, there will be a build of containers and images left behind that may no
longer be needed.  You can pick and choose which ones to keep.  A helper script has been provided
which removes all containers that have exited, intermediate images as well as dangling volumes.

DO NOT USE this script if any of the exited containers, images or volumes are of any value that
you would like to save.

The script can be found in [scripts/drclean](scripts/drclean).  Please read the comments if you would like to know
what it exactly does.

## FAQ
1. What is the difference between mold and Docker Compose? Why don't we use Docker Compose instead to test, build, package, and publish our applications?
    [Docker Compose](https://docs.docker.com/compose/overview/) is a tool to define and run multi-container applications while mold is to manage the CI steps. People may still wonder if Docker Compose could be used to achieve what mold does even though it is not originally made for the purpose. It does not seem like possible based on our experiment. Docker Compose controls the order of service startup but does not provide a way to manage when a image should be built. Below shows the docker-compose file and the Dockerfile we used to mimic the build process and test if the dependency condition would also delay the image build from the Dockerfile till the application is built.

    Docker-compose.yml
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
    Dockerfile
    ```
    FROM alpine
    ADD . /app
    CMD ["/bin/sh", "-f", "/app/fileExist.sh", "/app/myApp"]
    ```
    The result however shows the dependency condition only affects the order of the service startup (as what Docker Compose is made for.)

