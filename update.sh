#!/bin/sh
rm -rf `ls -d src/proto-gen/*.go`
protoc --go_out=paths=source_relative:. proto/*.proto --go-grpc_out=paths=source_relative:. proto/*.proto
mv proto/*.go src/proto-gen/