.PHONY: help build up down logs clean test cli

help:
	@echo "Available commands:"
	@echo "  make build       - Build Docker images"
	@echo "  make up          - Start services with docker-compose"
	@echo "  make down        - Stop services"
	@echo "  make logs        - View logs"
	@echo "  make clean       - Remove containers and volumes"
	@echo "  make test        - Run tests"
	@echo "  make cli         - Build CLI tool"
	@echo "  make install-cli - Install CLI tool to /usr/local/bin"

build:
	docker-compose build

cli:
	cd backend && go build -o ../bin/mdflow ./cmd/cli

install-cli: cli
	cp bin/mdflow /usr/local/bin/mdflow
	@echo "mdflow installed to /usr/local/bin/mdflow"

up:
	docker-compose up -d

down:
	docker-compose down

logs:
	docker-compose logs -f

clean:
	docker-compose down -v

test:
	cd backend && go test ./...

dev-backend:
	cd backend && go run ./cmd/server

dev-frontend:
	cd frontend && npm run dev

dev: build up logs
