.PHONY: build clean run

build:
	go build -o ./build/ server.go

run:
	./build/server

clean:
	rm -rf ./build/