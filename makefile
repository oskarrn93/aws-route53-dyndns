.PHONY: help
help:
	@echo "Available targets:"
	@echo "  install - Install dependencies"
	@echo "  update  - Update dependencies"
	@echo "  build   - Build the project"
	@echo "  run     - Run main.go using go run"
	@echo "  run-docker - Run the project using Docker Compose"
	@echo "  help    - Show this help message"
	@echo "  lint    - Run linter on the project"

.PHONY: install
install:
	go install ./...

.PHONY: update
update:
	go get -u ./...
	go mod tidy

.PHONY: run
run:
	go run main.go

.PHONY: run-docker
run-docker:
	docker-compose up --build -d

.PHONY: build
build:
	go build -o aws-route53-dyndns main.go

.PHONY: lint
lint:
	golangci-lint run
