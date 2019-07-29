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
* Be sure that you have correct auth data for a PostgreSQL and your user has all required permissions:

### How to create a user with superuser's permissions through a command line:
```bash
createuser -l USER_NAME -s superuser -P
```

### How to give permission to already existing user with SQL:
```sql
-- create pgcrypto extension, required for UUID
CREATE EXTENSION IF NOT EXISTS pgcrypto;

-- create the schemas required by the hasura system
-- NOTE: If you are starting from scratch: drop the below schemas first, if they exist.
CREATE SCHEMA IF NOT EXISTS hdb_catalog;
CREATE SCHEMA IF NOT EXISTS hdb_views;

-- make the user an owner of system schemas
ALTER SCHEMA hdb_catalog OWNER TO USER_NAME;
ALTER SCHEMA hdb_views OWNER TO USER_NAME;

-- grant select permissions on information_schema and pg_catalog. This is
-- required for hasura to query list of available tables
GRANT SELECT ON ALL TABLES IN SCHEMA information_schema TO USER_NAME;
GRANT SELECT ON ALL TABLES IN SCHEMA pg_catalog TO USER_NAME;

-- Below permissions are optional. This is dependent on what access to your
-- tables/schemas - you want give to hasura. If you want expose the public
-- schema for GraphQL query then give permissions on public schema to the
-- hasura user.
-- Be careful to use these in your production db. Consult the postgres manual or
-- your DBA and give appropriate permissions.

-- grant all privileges on all tables in the public schema. This can be customised:
-- For example, if you only want to use GraphQL regular queries and not mutations,
-- then you can set: GRANT SELECT ON ALL TABLES...
GRANT USAGE ON SCHEMA public TO USER_NAME;
GRANT ALL ON ALL TABLES IN SCHEMA public TO USER_NAME;
GRANT ALL ON ALL SEQUENCES IN SCHEMA public TO USER_NAME;

-- Similarly add this for other schemas, if you have any.
-- GRANT USAGE ON SCHEMA <schema-name> TO USER_NAME;
-- GRANT ALL ON ALL TABLES IN SCHEMA <schema-name> TO USER_NAME;
-- GRANT ALL ON ALL SEQUENCES IN SCHEMA <schema-name> TO USER_NAME;
```

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