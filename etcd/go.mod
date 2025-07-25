module github.com/demdxx/cloudregistry/etcd

go 1.23.0

toolchain go1.24.4

replace github.com/demdxx/cloudregistry => ../

require (
	github.com/demdxx/cloudregistry v0.0.0-00010101000000-000000000000
	github.com/demdxx/gocast/v2 v2.10.1
	github.com/demdxx/xtypes v0.2.0
	github.com/pkg/errors v0.9.1
	go.etcd.io/etcd/api/v3 v3.5.17
	go.etcd.io/etcd/client/v3 v3.5.17
	google.golang.org/grpc v1.68.1
)

require (
	github.com/coreos/go-semver v0.3.1 // indirect
	github.com/coreos/go-systemd/v22 v22.5.0 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/golang/protobuf v1.5.4 // indirect
	go.etcd.io/etcd/client/pkg/v3 v3.5.17 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	go.uber.org/zap v1.27.0 // indirect
	golang.org/x/exp v0.0.0-20250620022241-b7579e27df2b // indirect
	golang.org/x/net v0.32.0 // indirect
	golang.org/x/sys v0.28.0 // indirect
	golang.org/x/text v0.21.0 // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20241206012308-a4fef0638583 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20241206012308-a4fef0638583 // indirect
	google.golang.org/protobuf v1.35.2 // indirect
)
