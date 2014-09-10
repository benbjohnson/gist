default: assets

assets.go: assets/*
	@go-bindata -pkg gist -o assets.go -nocompress assets
	@gofmt -w -r "assets_favicon_ico -> favicon" assets.go 
