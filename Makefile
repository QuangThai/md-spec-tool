.PHONY: help build up down logs clean test cli eval-ai eval-ai-strict eval-ai-compare

help:
	@echo "Available commands:"
	@echo "  make build       - Build Docker images"
	@echo "  make up          - Start services with docker-compose"
	@echo "  make down        - Stop services"
	@echo "  make logs        - View logs"
	@echo "  make clean       - Remove containers and volumes"
	@echo "  make test        - Run tests"
	@echo "  make cli         - Build CLI tool"
	@echo "  make eval-ai     - Run AI eval suite and write JSON report"
	@echo "  make eval-ai-strict - Run AI eval suite with stricter gate"
	@echo "  make eval-ai-compare - Compare static_v3 against legacy_v2 prompts"
	@echo "  make install-cli - Install CLI tool to /usr/local/bin"

build:
	docker compose build

cli:
	cd backend && go build -o ../bin/mdflow ./cmd/cli

install-cli: cli
	cp bin/mdflow /usr/local/bin/mdflow
	@echo "mdflow installed to /usr/local/bin/mdflow"

up:
	docker compose up -d

down:
	docker compose down

logs:
	docker compose logs -f

clean:
	docker compose down -v

test:
	cd backend && go test ./...

eval-ai:
	cd backend && go run ./cmd/evalrunner \
		-mapping-dataset ./testdata/evals/mapping_cases.json \
		-suggestion-dataset ./testdata/evals/suggestion_cases.json \
		-output ./artifacts/ai-eval-report.json \
		-min-pass-rate 0.70

eval-ai-strict:
	cd backend && go run ./cmd/evalrunner \
		-mapping-dataset ./testdata/evals/mapping_cases.json \
		-suggestion-dataset ./testdata/evals/suggestion_cases.json \
		-output ./artifacts/ai-eval-report.json \
		-max-completion-tokens 1800 \
		-min-pass-rate 0.85

eval-ai-compare:
	cd backend && go run ./cmd/evalrunner \
		-prompt-profile static_v3 \
		-baseline-prompt-profile legacy_v2 \
		-mapping-dataset ./testdata/evals/mapping_cases.json \
		-suggestion-dataset ./testdata/evals/suggestion_cases.json \
		-output ./artifacts/ai-eval-report.json \
		-min-pass-rate 0.70

dev-backend:
	cd backend && go run ./cmd/server

dev-frontend:
	cd frontend && npm run dev

dev: build up logs
