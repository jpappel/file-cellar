.PHONY: all build run tidy clean

DBS := testing.db

all: build

./file-cellar: build

build: 
	go build .

run: ./file-cellar

clean: tidy
	rm -rf ./file-cellar $(DBS)
	
tidy:
	go mod tidy
