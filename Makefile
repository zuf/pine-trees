.PHONY: build
build: deps
	mkdir -p ./bin
	go build -o bin/pine-trees main.go

.PHONY: deps
deps:
	go mod vendor

.PHONY: run
run: build
	./bin/pine-trees

docker-image:
	docker build -t zufzzi/pine-trees .