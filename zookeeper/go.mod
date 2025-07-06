module github.com/demdxx/cloudregistry/zookeeper

go 1.23.0

toolchain go1.24.4

require (
	github.com/demdxx/cloudregistry v0.0.0
	github.com/demdxx/xtypes v0.3.0
	github.com/go-zookeeper/zk v1.0.3
)

require (
	github.com/demdxx/gocast/v2 v2.10.1 // indirect
	golang.org/x/exp v0.0.0-20250620022241-b7579e27df2b // indirect
)

replace github.com/demdxx/cloudregistry => ../
