release: setup
	goreleaser

setup:
	go get github.com/goreleaser/goreleaser
