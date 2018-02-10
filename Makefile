.PHONY: all assets

all: assets
	go build -ldflags "-X github.com/ysugimoto/ginger/request.debug=enable -X github.com/ysugimoto/ginger/command.debug=enable" -o dist/ginger

production: assets
	go build -o dist/ginger

assets:
	go-bindata -o internal/assets/assets.go --pkg assets --prefix misc ./misc
