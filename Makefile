.PHONY: all assets docs

VERSION=$(or ${CIRCLE_TAG}, ${BUILD_VERSION}, dev)

build: assets
	go build \
		-ldflags "-X github.com/ysugimoto/ginger/request.debug=enable -X github.com/ysugimoto/ginger/command.debug=enable" \
		-o dist/ginger

build-release:
	GOOS=darwin GOARCH=amd64 go build \
			 -ldflags "-X github.com/ysugimoto/ginger/command.version=$(VERSION)" \
			 -o dist/ginger-${CIRCLE_TAG}-osx
	GOOS=linux GOARCH=amd64 go build \
			 -ldflags "-X github.com/ysugimoto/ginger/command.version=$(VERSION)" \
			 -o dist/ginger-${CIRCLE_TAG}-linux

publish: build-release
	sh ./_tools/github-release.sh

assets:
	go-bindata -o assets/assets.go --pkg assets --prefix misc ./misc

docs:
	go run ./_tools/generate-doc.go
