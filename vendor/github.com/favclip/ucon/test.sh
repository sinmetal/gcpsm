#!/bin/sh -eux

goimports -w .
go generate ./...
go tool vet .
golint .
golint swagger
go test ./... $@
