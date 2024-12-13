.PHONY: build

build:
	GOPRIVATE="github.com/opengovern" CC=/usr/bin/musl-gcc GOPRIVATE="github.com/opengovern" GOOS=linux GOARCH=amd64 go build -v -ldflags "-linkmode external -extldflags '-static' -s -w" -o ./build/steampipe-plugin-oci.plugin *.go