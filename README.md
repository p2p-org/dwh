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
* Be sure that you have correct auth data for PostgreSQL
* Run:
```bash
make start-indexer
```
And if everything is all right you will see how indexer collects data from transactions


## How to start Hasura
* Be sure that you have correct auth data for a PostgreSQL
* Just run:
```bash
make start-hasura
```

The command will start a docker container with Hasura on http://localhost:8080


## Example of simple GraphQL query

Query:
```
{
  users {
    nfts {
      token_id
    }
    address
  }
}

```
Response:
```json
{
  "data": {
    "users": [
      {
        "nfts": [
          {
            "token_id": "7D5ED2AC-FF24-4321-91C5-ECB54281B38B"
          }
        ],
        "address": "cosmos1600upc35vevdd9p4jtzzq68w5p78e0sv86l200"
      }
    ]
  }
}
```

#Example of GraphQL query with gte operator

Query:

```
{
  users(where: {_or: {id: {_gte: 1}, address: {_eq: "cosmos16l2zcjlu4knx8f372wrmjwajxvfwhc3saa0zsw"}}}) {
    nfts {
      token_id
    }
    id
    address
  }
}
```

Response:

```json
{
  "data": {
    "users": [
      {
        "nfts": [],
        "id": 1,
        "address": "cosmos18dmtutv6eq3vcaqjupp3gmpy6fmn87s9cszg62"
      },
      {
        "nfts": [],
        "id": 2,
        "address": "cosmos16l2zcjlu4knx8f372wrmjwajxvfwhc3saa0zsw"
      }
    ]
  }
}
```