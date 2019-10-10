USER = "dgaming"
PASSWORD = "dgaming"
DB = "marketplace"

DB_HOST = "postgres"
DB_PORT = "5432"

OS := $(shell uname -s | tr A-Z a-z)

ifeq ($(OS),darwin)
	DB_HOST := "localhost"
endif

install: go.sum
		@go install ./cmd/indexer

go.sum: go.mod
		@echo "--> Ensure dependencies have not been modified"
		GO111MODULE=on go mod verify

start-indexer: install
	@indexer

stop-hasura:
	@ID=$$(docker ps -q --filter ancestor=hasura/graphql-engine:latest); \
	[[ !  -z  $$ID  ]] && echo "Stopping container..." && docker stop $$ID > /dev/null || true

start-hasura:	stop-hasura
	@echo "Starting container on http://localhost:8080..."
	@docker run -d -p 8080:8080 \
			-e HASURA_GRAPHQL_DATABASE_URL=postgres://$(USER):$(PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB) \
			-e HASURA_GRAPHQL_ENABLE_CONSOLE=true \
			hasura/graphql-engine:latest > /dev/null

.PHONY: install start-indexer stop-hasura start-hasura