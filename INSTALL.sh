#!/bin/bash
go get -u github.com/jteeuwen/go-bindata/...
$GOPATH/bin/go-bindata app/...
go build -o edi *.go