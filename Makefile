#/bin/bash

USER = "dgaming"
PASSWORD = "dgaming"
DB = "marketplace"

install: go.sum
		@go install ./cmd/indexer

go.sum: go.mod
		@echo "--> Ensure dependencies have not been modified"
		GO111MODULE=on go mod verify

start-indexer: install
	@indexer

stop-hasura:
	@ID=$$(docker ps -q --filter ancestor=hasura/graphql-engine:latest); \
	[[ !  -z  $$ID  ]] && echo "Stopping container..." && docker stop $$ID || true

start-hasura:	stop-hasura
	@echo "Starting container on http://localhost:8080..."
	@docker run -d -p 8080:8080 \
			-e HASURA_GRAPHQL_DATABASE_URL=postgres://$(USER):$(PASSWORD)@host.docker.internal:5432/$(DB) \
			-e HASURA_GRAPHQL_ENABLE_CONSOLE=true \
			hasura/graphql-engine:latest