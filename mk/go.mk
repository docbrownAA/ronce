# Build arguments for the docker image.
has_image = true

ldflags = -X 'ronce/src/go/app.Version=${version}' -X 'ronce/src/go/app.Name=${pkg}'
ifdef static
ldflags += -extldflags=-static
endif

build:: ## build the package
	$(info building ${pkg})
	go build -o ./bin/${pkg} -ldflags="${ldflags}" ./src/${pkg}

test:: ## test the package
	$(info testing ${pkg})
	go test ./src/${pkg} ${test_args}

