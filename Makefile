BIN := bin/console

.PHONY: help build run test vet fmt fmt-check bench profile web-install web-build web-test tf-validate compose-up compose-down clean

help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-14s\033[0m %s\n", $$1, $$2}'

build: ## Build the console backend
	go build -o $(BIN) ./cmd/console

run: build ## Run the backend on the bundled fixture (no broker needed)
	$(BIN)

test: ## Run Go tests with the race detector
	go test -race ./...

vet: ## go vet the backend
	go vet ./...

fmt: ## Format Go sources
	gofmt -w .

fmt-check: ## Fail if any Go file is not gofmt-clean
	@test -z "$$(gofmt -l .)" || { echo "gofmt needed:"; gofmt -l .; exit 1; }

bench: ## Run the fleet-summary read-path benchmark
	go test -run='^$$' -bench=BenchmarkFleetSummary -benchmem -count=8 ./internal/store/

profile: ## Regenerate the committed before/after profiling artifacts
	./scripts/profile.sh

web-install: ## Install frontend dependencies from the lockfile
	cd frontend && npm ci

web-build: ## Build the production SPA bundle
	cd frontend && npm run build

web-serve: ## Run the SPA dev server (proxies /api to the backend)
	cd frontend && npm start

web-test: ## Run the SPA unit tests (headless Chrome)
	cd frontend && npm run test:ci

tf-validate: ## Validate the SLO/deployment Terraform
	cd deploy/terraform && terraform init -backend=false -input=false && terraform validate

compose-up: ## Start broker + backend + fixture producer
	docker compose up --build

compose-down: ## Stop the stack and remove volumes
	docker compose down -v

clean: ## Remove build output
	rm -rf bin frontend/dist
