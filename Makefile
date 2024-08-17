.PHONY: all build run test tidy clean

DBS := testing.db
PKGS := db server storage

all: build

./file-cellar: build

build: 
	go build .

run: ./file-cellar

test:
	go test ./...

clean: tidy
	rm -rf ./file-cellar $(DBS)

tidy:
	go mod tidy
