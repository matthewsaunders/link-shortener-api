# syntax=docker/dockerfile:1

FROM golang:1.20

# Set destination for COPY
WORKDIR /app

# Copy everything from root directory into /app
COPY . .

# Download Go modules
RUN go mod download

# # Build
RUN CGO_ENABLED=0 GOOS=linux go build -o shrtnr-api ./cmd/api

EXPOSE 4000

# Run
CMD ["./shrtnr-api", "-env=production", "-cors-trusted-origins=http://localhost:3000", "-db-dsn=postgres://shrtnr:password@db:5432/shrtnr?sslmode=disable", "-migrate-db=true"]
