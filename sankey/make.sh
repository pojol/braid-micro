#! /bin/sh

rm sankey
echo "build sankey ..."
go build -o sankey /Users/pojol/work/gohome/src/braid/sankey/sankey.go

rm sankey_linux
echo "build sankey_linux ..."
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o sankey_linux /Users/pojol/work/gohome/src/braid/sankey/sankey.go
