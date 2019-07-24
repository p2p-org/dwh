#/bin/bash

USER = "dgaming"
PASSWORD = "dgaming"
DB = "marketplace"

install: go.sum
		go install ./cmd/indexer

go.sum: go.mod
		@echo "--> Ensure dependencies have not been modified"
		GO111MODULE=on go mod verify

start-indexer: install
	indexer

stop-hasura:
	docker stop $$(docker ps -q --filter ancestor=hasura/graphql-engine:latest) || true

start-hasura:	stop-hasura
	docker run -d -p 8080:8080 \
			-e HASURA_GRAPHQL_DATABASE_URL=postgres://$(USER):$(PASSWORD)@host.docker.internal:5432/$(DB) \
			-e HASURA_GRAPHQL_ENABLE_CONSOLE=true \
			hasura/graphql-engine:latest