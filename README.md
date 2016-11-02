# mold
Test, Build, Package and Publish your application completely using docker.

## Installation
[Download](https://github.com/d3sw/mold/releases) the binary based on your OS.  Once uncompressed copy it into your system PATH.

## Usage
To use mold you can simply issue the `mold` command in the working directory or your git
repository.  By default the command looks for a `.mold.yml` at the root of your project.  To
specify an alternate file you can use the `-f` flag followed by the path to your build config.

    Usage of mold:

      -f string
            Build config file (default ".mold.yml")
      -n    Enable notifications (default "false")
      -t string
            Build a specific target only [build|artifact|publish]
      -uri string
            Docker URI (default "unix:///var/run/docker.sock")
      -version
            Show version

In most cases you will simply issue the `mold` command.

## Config
The build configuration is broken up into the following sections:

- Services
- Build
- Artifacts/Publish

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
Services is a list of containers that need to be started prior to the build.  These are spun up
before the the code build actually starts.  They services can then be accesses via their image name
from the build container.

#### image
The image name of the service to start.

#### commands
These are the arguments passed to the service container.

## Build
Build contains a list of builds to perform. This is used to perform testing and/or building binaries.  
Each build will run its set of provided commands in the specified container.  Any failed
command will cause the build to fail.

#### image
Image name used to build/test code.  These are disposable and not used to generate the final
artifact.

#### commands
These are the commands that will be run in the container to do testing and building.

## Artifacts
Artifacts are the images that will generated as part of this build.  These are docker images
that would then get published to a registry.  

#### registry
This option sets the default registry for all images in the case where it is not supplied
as for of the image config.  This defaults to docker hub if not specified.

#### publish
This option specifies which branches will trigger a push to the registry.  `*` triggers
a push on all branches/tags.

#### images
A list of images to build.

- **name**: Specifies the name of the image
- **dockerfile**: Relative path to the Dockerfile.
- **registry**: Registry to push to.  If not specified the default one is used.
