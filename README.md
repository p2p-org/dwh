# Data Warehouse for Dgaming Marketplace

### Overview

DWH for Dgaming Marketplace consists of two parts:

* The `indexer` that collects data from the Cosmos Marketplace application and stores it in a Postgres database,
* The `Hasura` that provides a GraphQL querying interface for the collected data.

## Requirements
* A running node of Marketplace
* PostgreSQL
* Docker

## How to start Indexer
* Be sure that you have correct auth data for PostgreSQL connection
* Run:
```bash
make start-indexer
```
And if everything is all right you will see how indexer collects data from transactions


## How to start Hasura
* Be sure that you have correct auth data for a PostgreSQL connection
* Just run:
```bash
make start-hasura
```

The command will start a docker container with Hasura on http://localhost:8080