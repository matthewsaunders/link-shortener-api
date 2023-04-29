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
CREATE EXTENSION IF NOT EXISTS "pg_trgm";
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

3. Run the api
```
go run ./cmd/api
go run ./cmd/api -cors-trusted-origins='http://localhost:3000'
```

## Seeding the database

To seed the database, run the following command from the root of the repo
```
go run ./cmd/seeder
```
