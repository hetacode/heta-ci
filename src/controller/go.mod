module github.com/hetacode/heta-ci/controller

go 1.16

require (
	github.com/gorilla/mux v1.8.0 // indirect
	github.com/hashicorp/go-uuid v1.0.2
	github.com/hetacode/go-eh v0.0.2
	github.com/hetacode/heta-ci/events v0.0.1
	github.com/hetacode/heta-ci/proto v0.0.1
	github.com/hetacode/heta-ci/structs v0.0.1
	google.golang.org/grpc v1.36.1
	github.com/hetacode/heta-ci/commons v0.0.1
)

replace (
	github.com/hetacode/heta-ci/events => ../events
	github.com/hetacode/heta-ci/proto => ../proto-gen
	github.com/hetacode/heta-ci/structs => ../structs
	github.com/hetacode/heta-ci/commons => ../commons
)
