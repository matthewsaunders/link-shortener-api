# LinkShortener API

## Setup

1. Environment variables

```
. ./setup_env.sh
```


2. Database

```
# Create database and user
psql
CREATE ROLE shrtnr WITH LOGIN PASSWORD 'password';
CREATE DATABASE shrtnr OWNER=shrtnr;
\l # Confirm database exists
exit

psql --host=localhost --dbname=shrtnr --username=shrtnr
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
exit
```

```
# Connection string
psql --host=localhost --dbname=shrtnr --username=shrtnr
```

```
# Migrate database
migrate -path=./migrations -database=$SHRTNR_DB_DSN up
```


## APP Routes (shortener.com)

GET     /
GET     /dashboard
GET     /links
GET     /links/:id

## API Routes (api.shortener.com)
GET     /v1/links
POST    /v1/links
GET     /v1/links/:id
UPDATE  /v1/links/:id
DELETE  /v1/links/:id

GET     /links/:id/views

## Router Routes (shrtnr.com)
GET     /:token
