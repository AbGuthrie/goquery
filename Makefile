# goquery Makefile

export GO111MODULE=on

all: build/goquery

build/goquery:
	@mkdir -p build
	go build -o $@ cmd/main.go

.PHONY: docker deploy teardown

STACK_NAME = run_goquery_infra

docker:
	docker build -f docker/goserversaml/Dockerfile -t goserversaml:latest .
	docker build -f docker/goserver/Dockerfile -t goserver:latest .
	docker build -f docker/nodes/ubuntu-18/Dockerfile -t osquerydist .

deploy:
	docker stack deploy -c docker-compose.yml $(STACK_NAME)

teardown:
	docker stack rm $(STACK_NAME)

format:
	gofmt -w ./
