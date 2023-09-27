# Define the SILENT target if the verbose flag isn't set. This avoids flooding
# the output with potentially long build commands.
ifndef verbose
.SILENT:
endif

# Target selection. pkg is required on the command line, while version is
# retrieved from either the tags matching the package name or the current
# commit version.
tag = $(shell git describe --tags --match="${pkg}-*" --dirty="*" 2>/dev/null)
version ?= $(subst ${pkg}-,,${tag})
ifeq ($(strip $(version)),)
version = $(shell git rev-parse --short HEAD)
endif

# Include the package specific rules and configuration. The ${pkg}/pkg should
# contain the type definition.
ifdef pkg
include ./src/${pkg}/pkg

ifndef type
$(error missing type definition in package ${pkg})
endif
include ./mk/${type}.mk
endif

# The target themselves.

# Default target is to build and test.
all:: build test

describe:: ## display debug information about the package
	echo "pkg=${pkg}"
	echo "type=${type}"
	echo "version=${version}"

help:: ## display this help message
	fgrep -h "##" $(MAKEFILE_LIST) | fgrep -v fgrep | sed 's/\(.*\):: *##\(.*\)/\1\t\2/g' | expand -t 20

# If the type makefile defines the has_image variable, we enable the image and
# publish rules. The type makefile has the possibility to pre-define the
# dockerfile which will be used for building (useful for the docker package
# type).
ifdef has_image
dockerfile ?= docker/${type}.dockerfile
target ?= dist
image_build_args += --build-arg PKG='${pkg}'
image_build_args += --build-arg VERSION='${version}'
image:: ## build the docker image for the package
	$(info building ${pkg} image)
	docker buildx build --ssh default=${HOME}/.ssh/id_rsa . -f ${dockerfile} --target ${target} -t docbrownaa/${pkg}:${version} ${image_build_args}

publish:: ## publish the package image to docker
	$(info publishing ${pkg}:${version} image)
	docker push docbrownaa/${pkg}:${version}
endif

bin/transpiler:: $(shell find ./src/transpiler) ## build the transpiler
	make build pkg=transpiler --no-print-directory

ifdef pkg
endif