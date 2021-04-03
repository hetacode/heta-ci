module github.com/hetacode/heta-ci/controller

go 1.16

require (
	github.com/hashicorp/go-uuid v1.0.2 // indirect
	github.com/hetacode/go-eh v0.0.2 // indirect
	github.com/hetacode/heta-ci/proto v0.0.1
	github.com/hetacode/heta-ci/structs v0.0.1
	google.golang.org/grpc v1.36.1 // indirect
)

replace github.com/hetacode/heta-ci/structs => ../structs

replace github.com/hetacode/heta-ci/proto => ../proto-gen
