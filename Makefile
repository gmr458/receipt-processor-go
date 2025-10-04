run:
	go run ./cmd/webservice

git_description = $(shell git describe --always --dirty --tags --long)
linker_flags = '-s -w -X main.version=${git_description}'

build:
	rm -rf ./bin
	mkdir -p bin
	go build -ldflags=${linker_flags} -o ./bin/webservice ./cmd/webservice

start:
	./bin/webservice
