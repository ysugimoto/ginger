.PHONY: all assets

all: assets
	go build -ldflags "-X github.com/ysugimoto/ginger/request.debug=enable -X github.com/ysugimoto/ginger/command.debug=enable" -o dist/ginger

assets:
	go-bindata -o assets/assets.go --pkg assets --prefix misc ./misc
