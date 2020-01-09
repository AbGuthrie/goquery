# goquery Makefile

.PHONY: docker deploy teardown

STACK_NAME = run_goquery_infra

docker:
	docker build -f docker/goserversaml/Dockerfile -t goserversaml:latest .
	docker build -f docker/goserver/Dockerfile -t goserver:latest .
	docker build -f docker/nodes/ubuntu-18/Dockerfile -t osquerydist .

deploy:
	# docker stack deploy -c docker-compose.yml $(STACK_NAME)
	docker-compose up -d

teardown:
	# docker stack rm $(STACK_NAME)
	docker-compose down

format:
	gofmt -w ./

mock:
	mkdir -p build/
	go build -o build/mock_goquery examples/mock.go

mock-external:
	mkdir -p build/
	go build -o build/mock_external_goquery examples/mock_external.go

clean:
	rm -rf build/
