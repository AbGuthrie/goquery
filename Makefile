# goquery Makefile

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
