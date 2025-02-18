module operator-tool

go 1.23.1

require (
	github.com/Percona-Lab/percona-version-service v0.0.0-20241013113618-2966a16cabb1
	github.com/hashicorp/go-version v1.7.0
	google.golang.org/protobuf v1.34.2
)

replace github.com/Percona-Lab/percona-version-service => ../../

require (
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.19.1 // indirect
	golang.org/x/net v0.26.0 // indirect
	golang.org/x/sys v0.21.0 // indirect
	golang.org/x/text v0.16.0 // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20240711142825-46eb208f015d // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20240701130421-f6361c86f094 // indirect
	google.golang.org/grpc v1.65.0 // indirect
)
