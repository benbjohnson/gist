default: check

check: errcheck lint vet

errcheck:
	@errcheck ./...

lint:
	@golint .

vet:
	@go vet ./...

assets.go: assets/*
	@go-bindata -pkg gist -o assets.go -nocompress assets
	@gofmt -w -r "assets_favicon_ico -> favicon" assets.go 
	@gofmt -w -r "assets_logo_png -> logo" assets.go 
