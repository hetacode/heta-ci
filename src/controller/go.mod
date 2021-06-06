module github.com/hetacode/heta-ci/controller

go 1.16

require (
	github.com/alexflint/go-arg v1.4.2 // indirect
	github.com/go-git/go-git/v5 v5.3.0 // indirect
	github.com/gofrs/uuid v4.0.0+incompatible // indirect
	github.com/gorilla/mux v1.8.0 // indirect
	github.com/hashicorp/go-uuid v1.0.2
	github.com/hetacode/go-eh v0.0.2
	github.com/hetacode/heta-ci/commons v0.0.1
	github.com/hetacode/heta-ci/events v0.0.1
	github.com/hetacode/heta-ci/proto v0.0.1
	github.com/hetacode/heta-ci/structs v0.0.1
	github.com/lib/pq v1.10.2 // indirect
	github.com/robfig/cron/v3 v3.0.0 // indirect
	github.com/sijms/go-ora v1.2.0 // indirect
	github.com/xo/dburl v0.8.4 // indirect
	github.com/xo/xo v0.0.0-20210416025017-9a3ddc1e1407 // indirect
	golang.org/x/crypto v0.0.0-20210513164829-c07d793c2f9a // indirect
	google.golang.org/grpc v1.36.1
	gopkg.in/yaml.v2 v2.4.0 // indirect
)

replace (
	github.com/hetacode/heta-ci/commons => ../commons
	github.com/hetacode/heta-ci/events => ../events
	github.com/hetacode/heta-ci/proto => ../proto-gen
	github.com/hetacode/heta-ci/structs => ../structs
)
