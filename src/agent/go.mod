module github.com/hetacode/heta-ci/agent

go 1.16

require (
	github.com/Microsoft/go-winio v0.4.16 // indirect
	github.com/containerd/containerd v1.4.4 // indirect
	github.com/docker/distribution v2.7.1+incompatible // indirect
	github.com/docker/docker v20.10.5+incompatible // indirect
	github.com/docker/go-connections v0.4.0 // indirect
	github.com/docker/go-units v0.4.0 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/hashicorp/go-uuid v1.0.2 // indirect
	github.com/hetacode/go-eh v0.0.2 // indirect
	github.com/hetacode/heta-ci/proto v0.0.1
	github.com/hetacode/heta-ci/structs v0.0.1
	github.com/opencontainers/go-digest v1.0.0 // indirect
	github.com/opencontainers/image-spec v1.0.1 // indirect
	github.com/sirupsen/logrus v1.8.1 // indirect
	google.golang.org/grpc v1.37.0 // indirect
	github.com/hetacode/heta-ci/events v0.0.1
	github.com/hetacode/heta-ci/commons v0.0.1
)

replace (
	github.com/hetacode/heta-ci/proto => ../proto-gen
	github.com/hetacode/heta-ci/structs => ../structs
	github.com/hetacode/heta-ci/events => ../events
	github.com/hetacode/heta-ci/commons => ../commons
)
