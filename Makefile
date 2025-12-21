.PHONY: all build test clean docker-build docker-push deploy

# Variables
BINARY_NAME=webhook
DOCKER_IMAGE?=casbin/casbin-admission-webhook
VERSION?=latest
NAMESPACE?=casbin-system

all: test build

# Build the webhook binary
build:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -o $(BINARY_NAME) ./cmd/webhook

# Run tests
test:
	go test ./... -v -cover

# Run tests with coverage
coverage:
	go test ./... -coverprofile=coverage.out
	go tool cover -html=coverage.out -o coverage.html

# Clean build artifacts
clean:
	rm -f $(BINARY_NAME)
	rm -f coverage.out coverage.html

# Build Docker image
docker-build:
	docker build -t $(DOCKER_IMAGE):$(VERSION) .

# Push Docker image
docker-push: docker-build
	docker push $(DOCKER_IMAGE):$(VERSION)

# Generate TLS certificates and deploy to Kubernetes
deploy:
	./deploy/kubernetes/generate-certs.sh $(NAMESPACE)
	kubectl apply -f deploy/kubernetes/deployment.yaml

# Deploy webhook configuration
deploy-webhook-config:
	kubectl apply -f deploy/kubernetes/webhook-config.yaml

# Undeploy from Kubernetes
undeploy:
	kubectl delete -f deploy/kubernetes/webhook-config.yaml --ignore-not-found=true
	kubectl delete -f deploy/kubernetes/deployment.yaml --ignore-not-found=true
	kubectl delete namespace $(NAMESPACE) --ignore-not-found=true

# Run linting
lint:
	go fmt ./...
	go vet ./...

# Run the webhook locally (requires certs and config)
run:
	go run ./cmd/webhook

.DEFAULT_GOAL := all
