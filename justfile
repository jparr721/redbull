server:
  go run cmd/server/main.go

beacon:
  go run cmd/beacon/main.go

build:
  go build -o bin/beacon ./cmd/beacon/main.go
  go build -o bin/server ./cmd/server/main.go