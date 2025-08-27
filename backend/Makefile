BUILD_DIR := ./bin

build:
	@echo "Building application..."
	@mkdir -p $(BUILD_DIR)
	go build -o $(BUILD_DIR)/app cmd/main/main.go

run: build
	@echo "Starting application..."
	@$(BUILD_DIR)/app

TEST_PKGS := $(shell go list ./... | \
    grep -v /mocks | \
	grep -v /producer)

run_tests:
	@echo "==> Running tests..."
	@go test $(GOFLAGS) -coverprofile coverage_raw.out -v $(TEST_PKGS)

test: run_tests
	@echo "==> Calculating coverage..."
	@grep -v "mock" coverage_raw.out > coverage.out
	@go tool cover -func=coverage.out
	@go tool cover -html=coverage.out -o=coverage.html
	@echo "==> Done! Check coverage.html file!"

clean:
	@echo "==> Cleaning up..."
	@rm -rf $(BUILD_DIR)

docker_build:
	@echo "==> Building Docker images..."
	docker-compose build

docker_up:
	@echo "==> Starting containers..."
	docker-compose up

docker_down:
	@echo "==> Stopping containers..."
	docker-compose down

docker_restart: docker_down docker_up

docker_logs:
	@echo "==> Showing logs..."
	docker-compose logs -f

docker_rebuild: docker_down docker_build docker_up

.PHONY: build run test clean docker_build docker_up docker_down docker_restart docker_logs docker_rebuild
