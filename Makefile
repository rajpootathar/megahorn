.PHONY: build run test clean

build:
	go build -o megahorn .

run:
	go run .

test:
	go test ./... -v

clean:
	rm -f megahorn
