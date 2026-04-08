.PHONY: build run test clean docker-up docker-down

build:
	go build -o bin/chargeops ./cmd/server

run: build
	./bin/chargeops

test:
	go test ./...

clean:
	rm -rf bin/

docker-up:
	docker-compose up --build -d

docker-down:
	docker-compose down

vet:
	go vet ./...

fmt:
	gofmt -w .
