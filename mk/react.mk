# Build arguments for the docker image.
has_image = true

env ?= local

deps:: ## install the dependencies
	npm install

build:: deps ## build the package
	$(info building ${pkg})
	rm -rf ./bin/${pkg}
	npm exec -- tsc -p ./src/${pkg}/tsconfig.json --noEmit
	npm exec -- vite build ./src/${pkg}/ --outDir ../../bin/${pkg}

serve:: deps ## live-serve the package
	$(info serving ${pkg})
	npm exec -- vite ./src/${pkg}/ --port 8000

test:: deps ## test the package
	$(info testing ${pkg})
	npm exec -- jest --passWithNoTests --testPathPattern './src/${pkg}/.*'
