.PHONY: test build lint docker-build docker-push docker-run docker-clean

IMAGE_NAME ?= mcp-go-core
IMAGE_TAG ?= v0.1.0

test:
	go test ./...

build:
	go build ./...

lint:
	go vet ./...

docker-build:
	docker build -t $(IMAGE_NAME):$(IMAGE_TAG) .

docker-push:
	docker tag $(IMAGE_NAME):$(IMAGE_TAG) $(DOCKER_REGISTRY)/$(IMAGE_NAME):$(IMAGE_TAG)
	docker push $(DOCKER_REGISTRY)/$(IMAGE_NAME):$(IMAGE_TAG)

docker-run:
	docker compose --profile production up -d

docker-clean:
	docker compose down
	docker rmi $(IMAGE_NAME):$(IMAGE_TAG) 2>/dev/null || true
