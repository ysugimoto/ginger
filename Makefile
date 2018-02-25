.PHONY: all assets

all: assets
	go build \
		-ldflags "-X github.com/ysugimoto/ginger/request.debug=enable -X github.com/ysugimoto/ginger/command.debug=enable" \
		-o dist/ginger

pubish:
	GOOS=darwin GOARCH=amd64 go build -o dist/ginger-${CIRCLE_TAG}-osx
	GOOS=darwin GOARCH=amd64 go build -o dist/ginger-${CIRCLE_TAG}-linux
	sh ./_tools/github-release.sh

assets:
	go-bindata -o assets/assets.go --pkg assets --prefix misc ./misc
