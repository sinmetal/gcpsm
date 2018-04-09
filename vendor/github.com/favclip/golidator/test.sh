#!/bin/sh -eux

goimports -w .
go tool vet .
golint ./...
go test ./... $@
