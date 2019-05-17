TAG := $(shell git tag --points-at=HEAD)
GIT_BRANCH := $(shell git for-each-ref --format='%(objectname) %(refname:short)' refs/heads | awk "/^$$(git rev-parse HEAD)/ {print \$$2}")
GIT_HASH := $(shell git rev-parse --short HEAD)
TAG := $(or $(TAG),${GIT_BRANCH}-${GIT_HASH})

DOCKER_IMAGE_NAME:=zufzzi/pine-trees

.PHONY: build
build: deps
	mkdir -p ./bin
	go build -o ./bin/pine-trees ./src/main.go

.PHONY: deps
deps:
	go mod vendor

.PHONY: run
run: build
	./bin/pine-trees

docker-image:
	docker build -t ${DOCKER_IMAGE_NAME}:${TAG} .
	docker tag ${DOCKER_IMAGE_NAME}:${TAG} ${DOCKER_IMAGE_NAME}:latest
