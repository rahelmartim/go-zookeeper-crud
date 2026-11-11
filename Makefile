SITE := site
PORT ?= 8080

.PHONY: test test-zookeeper build web serve clean

## test: run the tests with the race detector
test:
	go vet ./...
	go test -race ./...

## test-zookeeper: run the tests, including the ones against ZooKeeper on localhost:2181
test-zookeeper:
	ZOOKEEPER_SERVERS=localhost:2181 go test -race -count=1 ./...

## build: build the zoocrud command into bin/
build:
	go build -o bin/ ./cmd/zoocrud

## web: build the browser demo into site/
web:
	rm -rf $(SITE) && mkdir -p $(SITE)
	cp web/* $(SITE)/
	cp "$$(go env GOROOT)/lib/wasm/wasm_exec.js" $(SITE)/
	GOOS=js GOARCH=wasm go build -trimpath -ldflags="-s -w" -o $(SITE)/zoo.wasm ./cmd/zoowasm

## serve: build the browser demo and serve it on http://localhost:$(PORT)
serve: web
	python3 -m http.server $(PORT) --directory $(SITE)

clean:
	rm -rf bin $(SITE)
