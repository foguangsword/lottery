.PHONY: build run tidy clean

build:
	go build -o backend-api ./cmd/api

run:
	go run ./cmd/api

tidy:
	go mod tidy

clean:
	rm -f backend-api
