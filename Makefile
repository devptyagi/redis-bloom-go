# Redis Bloom Filter Library - Makefile
# Provides convenient commands for development and testing

.PHONY: help install test clean

# Default target
help:
	@echo "Redis Bloom Filter Library - Available Commands:"
	@echo ""
	@echo "Development:"
	@echo "  install        Install dependencies"
	@echo "  test           Run all integration tests (inside Docker)"
	@echo "  clean          Remove build artifacts and containers"
	@echo ""

install:
	go mod tidy

# The only test shortcut: run all integration tests inside Docker Compose
# This will start redis, redis-cluster, and run the test service
# All networking is handled by Docker Compose

test:
	docker-compose up --build --abort-on-container-exit test

default: help

clean:
	docker-compose down -v
	rm -rf dist/ build/ *.out *.test *.log coverage.* 