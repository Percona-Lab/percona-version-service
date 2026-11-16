module operator-tool

go 1.23.4

toolchain go1.23.5

require (
	github.com/Percona-Lab/percona-version-service v0.0.0-20241013113618-2966a16cabb1
	github.com/hashicorp/go-version v1.7.0
	google.golang.org/protobuf v1.36.0
)

replace github.com/Percona-Lab/percona-version-service => ../../

require (
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.19.1 // indirect
	golang.org/x/net v0.33.0 // indirect
	golang.org/x/sys v0.28.0 // indirect
	golang.org/x/text v0.21.0 // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20241219192143-6b3ec007d9bb // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20241219192143-6b3ec007d9bb // indirect
	google.golang.org/grpc v1.69.2 // indirect
)
