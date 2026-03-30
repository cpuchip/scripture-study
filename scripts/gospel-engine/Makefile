.PHONY: build run index stats vet test clean

BINARY = gospel-engine
TAGS = -tags "fts5"

build:
	go build $(TAGS) -o $(BINARY).exe ./cmd/gospel-engine/

run: build
	./$(BINARY).exe serve

index: build
	./$(BINARY).exe index --verbose

index-sql: build
	./$(BINARY).exe index --no-vectors --verbose

stats: build
	./$(BINARY).exe stats

vet:
	go vet $(TAGS) ./...

test:
	go test $(TAGS) ./...

clean:
	rm -f $(BINARY).exe data/gospel.db
