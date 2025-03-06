.PHONY: build run clean

build:
	go build -o bin/lmdiff .

run: build
	./bin/lmdiff --branch main

clean:
	rm -f bin/lmdiff
